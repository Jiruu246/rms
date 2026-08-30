# Media Upload Framework

Direct-to-storage media uploads: the client uploads bytes straight to the
object store (Cloudflare R2) using a short-lived presigned URL, and the
application server never proxies file bytes. The server's job is limited to
issuing that URL, and later verifying — from the storage provider's own
metadata, never from the client's say-so — that what actually landed there
matches what was promised.

## Responsibility split

```
Client          -> requests an upload for one resource's image slot (purpose
                    is implied by the slot; the client declares neither
                    content-type nor size — see "Purpose policy" below)
ResourceHandler -> auth + request shape only, calls its own ResourceService
                    (e.g. RestaurantHandler -> RestaurantService)
ResourceService -> authorizes the actor against the resource, resolves purpose
                    from resource + slot, then calls MediaService
MediaService    -> purpose policy, key generation, session lifecycle — internal
                    only, no HTTP route of its own
StorageProvider -> presigned grants, StatObject/DeleteObject (pkg/storage)
Repository      -> MediaUpload (session) / MediaAsset (finalized) persistence
```

Concretely:

- **`MediaService`** It is an internal-only service, always reached through a resource-scoped route
- **`internal/services/media_service.go`** (`MediaService`) is where all the
  actual policy lives: which content-types/sizes a purpose accepts,
  authorization, object-key generation, and the finalize-time verification
  that the uploaded object satisfies that policy.
- **`pkg/storage`** (`StorageProvider` interface) is the provider-neutral
  boundary. **`internal/r2storage`** is the only implementation, backed by
  Cloudflare R2's S3-compatible API.
- **`internal/repos/media_upload_repo.go`** / **`media_asset_repo.go`**
  persist the two entities and own the transactional finalize step.

## Data model

Two entities, both under `internal/ent/schema/`:

| Entity | Schema file | Purpose |
|--------|-------------|---------|
| `MediaUpload` | `media_upload.go` | A short-lived upload *session*: the object key and expiry granted to one caller for one purpose. Single-use. Content-type and size are not persisted per session — see "Purpose policy" below for why. |
| `MediaAsset` | `media_asset.go` | A *finalized* upload: an object confirmed to exist in storage, ready for a resource to reference. |

`MediaUpload` lifecycle (`status` field):

```
issued --(ConsumeUpload succeeds)-------------------------------> consumed
issued --(ConsumeUpload rejects the uploaded object)------------> failed *
```

Both non-`issued` states are terminal — neither can ever transition back to
`issued` or into the other. There is no `expired` status: expiry is not
tracked as a state transition at all, only as the `expires_at` field.
`ConsumeUpload` treats "issued but past `expires_at`" as a read-time
conflict (see below) — an expired-but-never-consumed session is simply left
in `issued` state forever, since there is no sweep job to tidy it up.

\* `failed` means `ConsumeUpload` found an object in storage but rejected it
(content-type/size outside what the purpose allows). The
rejected object is deleted from storage and the session is marked `failed`
in the same step (`MediaService.rejectUpload`) — see Flow 3. A
`failed` session can never be consumed again; the caller must request a
fresh upload via `CreateUpload`.

`MediaAsset.status`: `active` -> `deleted` (soft delete; storage cleanup for
deleted assets is a later stage — see `MediaAssetRepository.Delete`'s doc
comment).

Key invariants encoded in the schema, not just in application code:

- `MediaUpload.object_key` is `Unique` — two sessions can never collide on
  the same storage key.
- `MediaAsset.upload_id` is `Unique` — at most one asset per upload session,
  which is what makes `MediaUploadRepository.Consume` safe under concurrent
  callers (see its doc comment: two transactions racing to consume the same
  upload serialize on the row, and the loser's conditional `UPDATE` affects
  zero rows).
- `MediaAsset.uploaded_by_user_id` is **provenance only, not an ACL** — once
  an asset is attached to a resource (a menu item, a restaurant), that
  resource's own ownership/restaurant scope governs access, not who
  originally uploaded the file. `MediaService.DeleteMedia` reflects this: it
  performs no authorization check of its own (see below).

## Purpose policy

`mediaupload.Purpose` (ent-generated enum) has three values, each mapped to
an allow-list of content-types and a max size in `purposeConstraints`
(`internal/services/media_service.go`):

| Purpose | Allowed content-types | Max size |
|---------|------------------------|----------|
| `menu_item_image` | image/jpeg, image/png, image/webp | 10 MiB |
| `restaurant_logo` | image/jpeg, image/png, image/webp | 5 MiB |
| `restaurant_cover_image` | image/jpeg, image/png, image/webp | 10 MiB |

SVG is deliberately excluded from every purpose, including logos — it can
embed scripts and this content is potentially served back to other users.

This table is code, not config: the set of purposes is fixed and small, and
each one already maps to a specific place in the product that changing
requires a code change anyway. Adding a new purpose means adding an ent enum
value (`internal/ent/schema/media_upload.go`, then `go generate
./internal/ent`) and a matching entry in `purposeConstraints`.

## Storage abstraction (`pkg/storage`)

| File | Type | Notes |
|------|------|-------|
| `provider.go` | `StorageProvider` interface | `CreateUploadGrant`, `StatObject`, `DeleteObject`. Implementations live in infrastructure packages (`internal/r2storage`); nothing provider-specific leaks through this interface. |
| `object.go` | `ObjectKey`, `ObjectMetadata`, `StoredObject` | `ObjectKey` is provider-neutral — no bucket/path semantics implied here. |
| `upload.go` | `UploadRequest`, `UploadGrant` | `UploadGrant` is everything a client needs for a direct upload: method, URL, headers, expiry. `UploadRequest` deliberately has no `ContentType` field — see below. |

`internal/r2storage.Provider` is the only implementation today, over
Cloudflare R2's S3-compatible API. Two properties of it matter for how
`MediaService` is written:

- **`CreateUploadGrant` never touches the network** — presigning is a local
  computation. Only `StatObject`/`DeleteObject` make real HTTP calls (with a
  10s timeout). This is why constructing a `Provider` at server startup with
  fake-but-valid test credentials is safe (see `config.LoadTestConfig`) —
  nothing actually talks to R2 unless a test calls `StatObject`/`DeleteObject`.
- **The presigned PUT does not cryptographically bind content-type or size.**
  The AWS SDK strips `Content-Type` from the signature before presigning, and
  a presigned PUT URL has no notion of a max size at all. This is *why*
  `ConsumeUpload`'s post-hoc verification against `StatObject` exists — see
  "Never trust the client" below.

## Flow 1 — `CreateUpload` (request an upload)

`MediaService.CreateUpload` has no HTTP route of its own — like
`ConsumeUpload` (Flow 3), every upload purpose is always attached to a
specific resource, so requesting an upload is always a step inside that
resource's own flow. Today the only caller is Restaurant's
`POST /restaurants/{id}/images/{slot}/uploads` (see Flow 4):
`RestaurantHandler.CreateRestaurantImageUpload` ->
`RestaurantService.CreateImageUpload` -> `MediaService.CreateUpload`.
`RestaurantService` resolves the purpose from the resource + `:slot` (the
same `restaurantImageSlotPurposes` mapping `UpdateImage` uses) and passes it
in — the client never supplies purpose directly.

```
1. Authorize the actor against the resource     (e.g. RestaurantService.AuthorizeOwnership)
2. Resolve the purpose from resource + slot     (e.g. restaurantImageSlotPurposes)
3. Validate the purpose                          (mediaupload.PurposeValidator)
4. Authorize the actor for media_upload:create   (authz.PolicyAuthorizer, self-ownership)
5. Look up the purpose's constraints             (purposeConstraints for that purpose)
6. Generate the object key                       (media/{ownerID}/{uuid} — backend-only)
7. Request an upload grant                       (StorageProvider.CreateUploadGrant)
8. Persist the MediaUpload session               (ExpiresAt = the grant's actual expiry)
9. Return {upload_id, expires_at, upload, constraints}
```

Notes on ordering vs. what you might expect from the step list above: the
grant is requested *before* the session is persisted (not after), and the
session's `expires_at` is taken from the grant's actual (possibly
provider-capped) expiry — not a locally-computed guess. This avoids ever
persisting a session for which no grant was successfully issued.

The client never supplies the object key — it is always backend-generated
as `media/{ownerID}/{uuid}`, and never derived from the original filename
(filenames are not relied on for uniqueness; the `filename` request field is
optional and unpersisted — schema has no filename column at all).

**Request** (`POST /restaurants/{id}/images/{slot}/uploads`):

There is deliberately no `purpose` field — the route's own `{id}/images/{slot}`
already determines it.

There is also deliberately no client-declared `content_type` or `size_bytes`:
a presigned R2 PUT URL doesn't cryptographically bind either one (see
`r2storage.Provider.CreateUploadGrant`), so a value the client claims ahead of
time can't be enforced any harder than the purpose's own policy already is at
consume time (Flow 3, step 7) — the client just needs to know what that
policy is, which the `constraints` field in the response below provides.
`filename` is the only field on the request, and it's an optional,
unpersisted hint (not shown above — the request body is typically empty).

**Response** (`201 Created`, wrapped in the standard `utils.APIResponse`
envelope):

```json
{
  "success": true,
  "data": {
    "upload_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "expires_at": "2026-08-16T18:00:00Z",
    "upload": {
      "method": "PUT",
      "url": "https://<account>.r2.cloudflarestorage.com/media/...&X-Amz-Signature=...",
      "headers": { "Host": "<account>.r2.cloudflarestorage.com" }
    },
    "constraints": {
      "max_size_bytes": 10485760,
      "allowed_content_types": ["image/jpeg", "image/png", "image/webp"]
    }
  },
  "timestamp": "2026-08-16T17:45:00Z"
}
```

`constraints` is derived from the same per-purpose policy (`purposeConstraints`
in `MediaService`) that Flow 3 enforces server-side, so the two can never
drift apart — it exists purely so the client can show/validate limits without
hardcoding them.

The storage object key is **deliberately never included** in this response
(`dto.UploadGrant.ObjectKey` is `json:"-"`) — the client never needs it, it
only ever refers to this upload again by `upload_id`.

## Flow 2 — the client uploads (outside this API)

The client issues the HTTP request described by `upload` directly to the
storage provider (`PUT <url>` with the given headers and the file body).
The application server is not involved in this step at all — this is the
entire point of a presigned-URL flow.

## Flow 3 — `ConsumeUpload` (finalize)

`MediaService.ConsumeUpload` has no HTTP route of its own — every upload
purpose is always attached to a specific resource, so finalizing a session
is always a step inside that resource's own flow (e.g. Restaurant's
`PUT /restaurants/{id}/images/{slot}` — see Flow 4), never a standalone
"finalize this upload" action a client calls directly. The upload ID plus
the JWT actor is everything `ConsumeUpload` needs.

```
1. Load the upload session, scoped to the caller     (MediaUploadRepository.GetByID)
2. Ownership is folded into step 1 — a session owned by someone
   else is indistinguishable from one that doesn't exist (apperr.ErrNotFound)
3. Fast-path check: session still "issued"?          (else apperr.Conflict —
   "failed" and "consumed" get distinct messages, see below)
4. Fast-path check: session not past expires_at?      (else apperr.Conflict)
5. StatObject against the storage provider            (confirms the object exists)
6. Self-check: the returned object's key matches the session's own key
   (an opaque 500 if not — this guards against a bug in our own code or the
   provider, not against anything a client controls)
7. Compare actual metadata against the purpose's own policy (apperr.Invalid, 400, if not):
     - size <= the purpose's max
     - content-type is one of the purpose's allowed content-types
   Neither check is against a value declared at CreateUpload time — see
   "Purpose policy" above for why the purpose's own policy is the sole
   source of truth here, not what the session promised.
   On failure: clean up (see below), then return the validation error.
8. MediaUploadRepository.Consume: re-checks issued/not-expired
   transactionally, creates the MediaAsset, marks the session consumed
```

**Return value**: the finalized `dto.MediaAsset` — `id`, `content_type`,
`size_bytes`, `status`, `create_time`, `update_time` — for the
caller (e.g. `RestaurantService`) to attach to its own resource.
`storage_key` is **not** serialized (`json:"-"`) — same rule as
`CreateUpload`: the client references media by ID from here on, never by
the internal storage key.

**Never trust the client's "upload succeeded" report.** Steps 5–7 are the
entire reason this method exists rather than trusting whatever the frontend
claims. Two caveats worth knowing:

- The content-type check is an **allow-list membership check, not a
  security guarantee** — R2's presigned PUT doesn't cryptographically bind
  content-type (see above), so "actual" is just whatever header the
  client's PUT happened to send. It catches disallowed types, not
  adversarial content smuggling within an allowed type.
- Steps 3–4 are a fast path so an already-settled session fails before ever
  touching the storage provider. The *authoritative*, race-safe version of
  the same checks runs again inside `MediaUploadRepository.Consume`'s
  transaction — see that method's doc comment for how it stays safe under
  two callers racing to consume the same session.

### Cleanup on rejected uploads

When step 7 rejects the object, `MediaService.rejectUpload`:

1. Deletes the rejected object from storage (`StorageProvider.DeleteObject`)
   — an invalid upload must not sit in the bucket indefinitely.
2. Terminally marks the session `failed` (`MediaUploadRepository.Fail`) —
   the client cannot retry this session; they must call `CreateUpload` again
   for a fresh key. `Fail` uses the same conditional-`UPDATE`-under-a-race
   pattern as `Consume` (`Where(status = issued)`), so it can't clobber a
   session a concurrent call already legitimately consumed.

Both steps are **best-effort**: either one failing is logged
(`log.Printf`), never surfaced to the client. The client already has the
real reason (the validation error from step 7) — an unrelated storage/DB
hiccup during cleanup must not replace or mask that with a confusing 500.

A subsequent call to `ConsumeUpload` against a `failed` session gets a
distinct, clear message ("was rejected and cannot be retried; request a new
upload") rather than being lumped in with the generic "already consumed"
conflict message a `StatusConsumed` session gets.

## Flow 3b — purpose enforcement at consume time

`MediaService.ConsumeUpload` takes an `expectedPurpose mediaupload.Purpose`
parameter. When non-empty, it must match the purpose the session was
actually created with (`mediaupload.Purpose`, loaded from the session row)
or the call is rejected with `apperr.Invalid` (400) before the storage
provider is ever consulted.

This closes a real gap: without it, a caller could open a
`menu_item_image` upload session (10 MiB cap) and hand that same
`upload_id` to a feature that expects a `restaurant_logo` (5 MiB cap),
silently inheriting the wrong purpose's constraints. Every caller resolves
an upload on behalf of a specific attachment point — the Restaurant
integration below is the first one — and passes its own purpose here.
`""` (skipping the check) is only used by tests exercising the rest of
`ConsumeUpload`'s validation independently of purpose enforcement; there is
no real caller that needs it.

## Flow 4 — Restaurant logo / cover image as independent subresources

Restaurant is the first (and, per current scope, only) resource wired up to
this framework. It has two independent image slots — `logo_media_asset_id`
and `cover_image_media_asset_id` — each a nullable FK to `MediaAsset`,
resolved through `mediaupload.PurposeRestaurantLogo` /
`PurposeRestaurantCoverImage` respectively. They replaced the old
`logo_url`/`cover_image_url` raw-string columns entirely (a breaking schema
change, not a dual-write migration — there was no production data to
preserve). The client never sends a URL or a storage key for either slot,
only an `upload_id` obtained from that slot's own upload route.

Each slot is its own subresource, addressed by `dto.RestaurantImageSlot` —
an allow-list (`logo`, `cover`; see `dto.ParseRestaurantImageSlot`), never an
arbitrary client-supplied column name:

- `POST /restaurants/{id}/images/{slot}/uploads` — request a presigned
  upload for that slot (see Flow 1); the purpose is derived from the slot,
  never supplied by the client.
- `PUT /restaurants/{id}/images/{slot}` — assign the asset resolved from
  `upload_id` to that slot.
- `DELETE /restaurants/{id}/images/{slot}` — clear whatever is assigned to
  that slot.

Both live entirely outside `PATCH /restaurants/{id}`: an image assignment
and a plain-attribute update never share a request or a failure boundary.
Updating the name and replacing the logo are two independent HTTP calls,
each succeeding or failing on its own.

`RestaurantHandler` and `RestaurantRepository` know nothing about R2, S3, or
`MediaService` — all of that lives in `RestaurantService`, which is the only
caller of `MediaService` in this stage. This keeps the "never import R2/S3
code into Restaurant services" and "never pass raw object keys through
Restaurant handlers" rules structurally true rather than just documented.

### Create and the plain Update never touch media

`POST /restaurants` (`dto.CreateRestaurantRequest`) has no logo/cover fields
at all and never calls `MediaService`. `RestaurantService.Create` is a
plain pass-through to `RestaurantRepository.Create`. Likewise,
`dto.UpdateRestaurantRequest` (the body of `PATCH /restaurants/{id}`) has no
upload-ID fields — `RestaurantService.Update` is a plain pass-through to
`RestaurantRepository.Update` and never resolves an upload.

This is deliberate: it keeps a restaurant's plain-attribute writes
independent of R2/media health, and it means an API consumer only ever has
one way to attach an image (a follow-up call to its own slot endpoint), not
two competing ones that invite a fragile flow being built around whichever
happens to work.

The recommended client flow is therefore: `POST /restaurants` with plain
data (fast, and its success is independent of R2 entirely) → for each image
the user wants attached, `POST /restaurants/{id}/images/{slot}/uploads` →
upload to R2 → `PUT /restaurants/{id}/images/logo` or
`PUT /restaurants/{id}/images/cover` with that slot's `upload_id`. Each
image can be done independently and in any
order (including concurrently — see below), and a failed image attach only
needs *that* step retried, not restaurant creation or any other attribute
update.

### UpdateImage / ClearImage each own exactly one image slot

A restaurant is never expected to replace both images in the same request,
so both endpoints take the slot as a path parameter and only ever touch that
one column:

- `PUT /restaurants/{id}/images/{slot}` (`dto.UpdateRestaurantImageRequest{upload_id}`)
  → `RestaurantHandler.UpdateRestaurantImage` → `RestaurantService.UpdateImage`
- `DELETE /restaurants/{id}/images/{slot}` →
  `RestaurantHandler.DeleteRestaurantImage` → `RestaurantService.ClearImage`

Attaching a restaurant's first-ever logo and replacing an existing one go
through the exact same `UpdateImage` call — the operation is a plain
overwrite of the column, so there is no "first attach" special case.
Because each call only ever writes its own slot's column, calling both slots
concurrently against the same restaurant (e.g. the logo and cover image
uploaded independently) is safe — there's no read-modify-write overlap
between them.

`UpdateImage`:

```
1. AuthorizeOwnership (existing behavior, unchanged)
2. Resolve the upload via
   MediaService.ConsumeUpload(ctx, actor, uploadID, expectedPurpose)
     - this is "resolve/consume" + "verify the object exists" + "create
       MediaAsset" collapsed into one call — see Flow 3 — and enforces the
       matching purpose (restaurant_logo / restaurant_cover_image), same
       protection described in Flow 3b
3. RestaurantRepository.SetImage persists the new asset ID onto that one
   slot's column, overwriting whatever was there before
```

Consuming an upload (step 2) does a real network call (`StatObject`) and
commits its own transaction inside `MediaUploadRepository.Consume` — it
cannot be wrapped in the same DB transaction as the restaurant row update
without violating "never hold a DB transaction open across a remote storage
call".

**Whatever was previously assigned to a slot is never touched by
`UpdateImage` or `ClearImage` — not deleted from storage, not even
soft-deleted in the database.** Once step 3 repoints (or, for `ClearImage`,
clears) the column, the old `MediaAsset` row becomes an orphan: an `active`
row (and its R2 object) that nothing references anymore. This is a
deliberate scope cut for this iteration, not an oversight — reconciling
these orphans (soft-delete, then eventually the R2 object itself) is left
for a later pass; see "Deferred / not yet done" below.

Concretely:

- **Plain-attribute update**: `PATCH /restaurants/{id}` never touches an
  image slot — `MediaService` is never called from that path, regardless of
  what either slot currently holds.
- **First image attach / image replacement succeeds**: the new asset is
  consumed and the slot's column is overwritten. Whatever the slot held
  before (if anything) is left exactly as it was — see above.
- **Consuming the upload fails**: `ConsumeUpload` in step 2 returns an error
  (content-type/size/purpose mismatch, or an already-consumed/expired
  conflict) — the function returns immediately; step 3 never runs, so the
  row is untouched.
- **The row write fails after a new upload was consumed**: step 2 succeeded,
  step 3 fails — the newly-consumed asset is compensated (soft-deleted), the
  same synchronous-compensation shape used everywhere else in this
  framework. This is unrelated to the old-asset orphaning above: it only
  cleans up the *new* asset this request itself created but never managed to
  attach anywhere.

### Idempotency

- **`DELETE .../images/{slot}`** is idempotent for free — clearing an
  already-nil column is not an error, so a repeated clear of an
  already-empty slot just succeeds again.
- **`PUT .../images/{slot}`** is not idempotent under retry: `MediaUpload`'s
  single-use invariant (see "Data model" above) means a session can only
  ever be consumed once, so retrying the exact same request (same
  `upload_id`) after an earlier attempt already consumed it surfaces
  `ConsumeUpload`'s ordinary "already consumed" conflict, even if that
  earlier attempt's response never reached the client. Clients must request
  a fresh upload session to retry an image attach.

### Display URLs

`dto.Restaurant` exposes only derived, read-only `logo_url` /
`cover_image_url` strings — never the asset ID and never the storage key.
The client has no use for the `MediaAsset` ID itself, only the image it
resolves to, so `LogoMediaAssetID`/`CoverImageMediaAssetID` on the Go struct
are `json:"-"`: they exist purely so `RestaurantService` can read the
currently-attached asset off the same response type (e.g. to check an image
update against it for idempotency, see above) without ever putting it on
the wire. `RestaurantRepository` builds the URLs from `R2Config.PublicBaseURL`
plus the attached `MediaAsset.StorageKey`,
loaded via ent's `WithLogoAsset()`/`WithCoverImageAsset()` edge
eager-loading on every read path (`GetByID`, `GetAllForUser`, and the
re-read `Create`/`Update`/`SetImage`/`ClearImage` do after writing). If
`PublicBaseURL` isn't configured, both URL fields are simply omitted — this
mirrors `dto.MediaAsset`'s own doc comment, which left "resolving the
storage key to a servable URL" to whatever attaches the asset to a
resource.

### Schema migration note

The old `logo_url`/`cover_image_url` string columns were dropped outright
(`internal/ent/schema/restaurant.go`), not kept alongside the new FK
columns. `cmd/migrate apply` uses ent's declarative schema sync with
`migrate.WithDropColumn(true)`, so running it against a database that still
has the old columns will drop them along with any data in them.

## Flow 5 — `DeleteMedia` — service only, no HTTP route yet

`MediaService.DeleteMedia(ctx, actor, mediaID)` soft-deletes a `MediaAsset`
(`MediaAssetRepository.Delete`) and performs **no authorization check of its
own**. `actor` is accepted only for audit/logging symmetry with the other
two methods.

This is deliberate, not an oversight: callers reach a specific asset through
a resource (a menu item, a restaurant) that has already authorized the actor
against *that resource's own* ownership/restaurant scope, per the schema
comment on `MediaAsset.uploaded_by_user_id` (see "Data model" above). Adding
a redundant uploader-ownership check here would be actively wrong once
assets are shared/reassigned across resources — it would block a legitimate
resource-owner deletion just because they didn't personally upload the file.

Storage cleanup (actually deleting the R2 object) is explicitly deferred —
see the doc comment on `MediaAssetRepository.Delete`. Note this is a
different situation from Flow 3's cleanup: that deletes a *rejected,
never-finalized* object; this would (eventually) delete a *finalized*
asset's object after the fact.

## Error mapping

Same sentinels/HTTP mapping as everywhere else in the API (see
`documentation/ErrorHandlingFramework.md`) — `MediaService` constructs
these, and the calling handler (e.g. `RestaurantHandler`) never inspects
error content itself:

| Situation | Sentinel | HTTP |
|-----------|----------|------|
| Unknown purpose, uploaded object violates the purpose's content-type/size policy (Flow 3, step 7), upload's purpose doesn't match the caller's expected purpose (Flow 3b) | `apperr.ErrInvalid` | 400 |
| Malformed upload ID or restaurant ID in the URL path, unrecognized `:slot` (anything other than `logo`/`cover`) | none — handler-level `utils.WriteBadRequest` | 400 |
| Actor doesn't own the upload session (folded into the `GetByID` lookup) | `apperr.ErrNotFound` | 404 |
| Session already consumed (by this request or an earlier retry), previously failed/rejected, expired, or no object uploaded yet | `apperr.ErrConflict` | 409 |
| Actor fails `authz.PolicyAuthorizer` (self-ownership check) | `apperr.ErrForbidden` | 404 (never 403 — see error-handling doc) |
| Missing/invalid `Authorization` header | `apperr.ErrUnauthorized` (via JWT middleware) | 401 |
| Anything else (provider/network failure, unexpected bug, including the object-key self-check) | none — opaque `error` | 500, generic fallback message; real error only logged server-side |

## Configuration (`internal/config/r2Config.go`)

Two independently-configurable expiries, easy to confuse:

| Field | Env var | Default | Meaning |
|-------|---------|---------|---------|
| `MaxGrantExpiry` | `R2_MAX_GRANT_EXPIRY` | 15m | **Provider-side cap.** `r2storage.Provider.CreateUploadGrant` clamps any requested expiry to this, regardless of what's asked for. |
| `UploadGrantExpiry` | `R2_UPLOAD_GRANT_EXPIRY` | 15m | **Application-level default.** What `MediaService.CreateUpload` actually requests. Independent of the cap above — tunable without touching provider settings. |

`R2Config.Validate()` requires both to be positive, plus the usual
account/credential/bucket fields.

## API surface today

| Method | Path | Auth | Handler |
|--------|------|------|---------|
| `POST` | `/api/restaurants/{id}/images/{slot}/uploads` | JWT required | `RestaurantHandler.CreateRestaurantImageUpload` |
| `PUT` | `/api/restaurants/{id}/images/{slot}` | JWT required | `RestaurantHandler.UpdateRestaurantImage` |
| `DELETE` | `/api/restaurants/{id}/images/{slot}` | JWT required | `RestaurantHandler.DeleteRestaurantImage` |

There is deliberately no generic, resource-agnostic "create an upload" or
"finalize this upload" route: every upload purpose is always attached to a
specific resource, so both `MediaService.CreateUpload` and
`MediaService.ConsumeUpload` are only ever invoked internally, as steps
inside that resource's own flow. The Restaurant image-slot routes are
ordinary resource verbs against a subresource: `POST .../uploads` requests
a presigned grant for that slot, `PUT` assigns (replacing whatever was
there, consuming the upload as part of the same request), `DELETE`
detaches — see Flow 4.

`DeleteMedia` has no HTTP route of its own and, as of this iteration, is no
longer called from the Restaurant image flow at all — `UpdateImage` and
`ClearImage` never retire the asset a slot previously held (see "Whatever
was previously assigned..." in Flow 4). It is still used internally to
compensate a freshly-consumed asset when the row write in the *same request*
fails partway through. A direct client-facing route would still need to be
added if some future feature needs to delete a `MediaAsset` outright.

## Security properties

- The client can never choose or influence the storage object key — it is
  always `media/{ownerID}/{uuid}`, generated server-side, never derived from
  a client-supplied filename. `internal/r2storage/mapping.go`'s `toS3Key`
  additionally rejects any key starting with `/` or containing a `..`
  segment before it reaches an S3 call — defense-in-depth against a bug in
  our own key generation, not a client-facing control.
- The storage object key is never returned to the client either — it isn't
  needed (the client only ever refers to an upload by `upload_id`), and
  keeping it server-only reduces what a client could infer about internal
  storage layout.
- The server's own R2 credentials (account ID, access key, secret) never
  appear in any request or response — only a scoped, time-limited presigned
  URL does, which is the entire point of the presigned-URL pattern.
- Handlers never log response bodies (gin's request logger logs
  method/path/status/latency only), so presigned URLs are never written to
  logs by this code path.
- A client's claim that an upload "succeeded" is never trusted — see Flow 3.
- A rejected upload's object is deleted from storage and its session
  permanently barred from retry — see "Cleanup on rejected uploads" in
  Flow 3.

## Deferred / not yet done

- **No sweep for expired sessions.** There is no `expired` status; expiry is
  derived purely from `expires_at`. `ConsumeUpload` handles "issued but past
  `expires_at`" as a read-time check, and nothing ever sweeps stale `issued`
  rows out of the table.
- **No storage cleanup on `DeleteMedia`.** `MediaAssetRepository.Delete` only
  soft-deletes the DB row; the underlying R2 object is left in place. (This
  is distinct from Flow 3's cleanup, which does delete storage — for a
  *rejected* object, not a finalized asset being deleted later.)
- **`DeleteMedia` has no HTTP route of its own** — see "API surface today".
  It is only invoked internally, to compensate a freshly-consumed asset when
  a Restaurant image-slot write fails partway through the same request — not
  exposed for a client to call directly yet.
- **Replaced/cleared Restaurant images are not reconciled at all.**
  `UpdateImage`/`ClearImage` (Flow 4) never soft-delete, let alone actually
  delete from storage, whatever asset a slot previously held — they simply
  overwrite or clear the column. Every image replacement or clear leaves
  behind an orphaned, still-`active` `MediaAsset` row (and its R2 object),
  with neither a synchronous step nor a background sweep to reconcile it.
  Explicitly deferred to the next iteration.
- **No async cleanup / sweep for orphaned assets generally.** Beyond the
  image-slot case above, the synchronous compensation that *does* still
  exist (soft-deleting a newly-consumed asset when the same request's row
  write fails) is best-effort and scoped to that one request. There is no
  background job that finds and retires orphans missed by that path (e.g. a
  crash between consuming and writing the row) — deliberately out of scope
  for this stage.
- **Only Restaurant is integrated.** `MenuItem.image_url` (and any future
  resource) still stores/accepts a raw URL and does not go through this
  framework — migrating it would repeat Flow 4's schema-change + service
  work for that entity, deliberately not done as part of this stage.
- **Purpose policy (`purposeConstraints`) is not config-driven** — adding a
  new purpose is a code change (ent enum + constraints map), by design (see
  "Purpose policy" above).
