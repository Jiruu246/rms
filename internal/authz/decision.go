package authz

// AccessMode records *why* a request was allowed, for audit logging.
type AccessMode string

const (
	AccessOwner AccessMode = "owner"
	AccessAdmin AccessMode = "admin"
)

// Decision is the result of a successful Authorize call. When access is
// denied, Authorize returns a zero Decision and a sentinel error (see errors.go)
// instead — Decision.Allowed is not used as the success signal.
type Decision struct {
	Reason string
	Mode   AccessMode
}
