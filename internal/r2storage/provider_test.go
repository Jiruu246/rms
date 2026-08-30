package r2storage

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/pkg/storage"
)

func testR2Config() config.R2Config {
	return config.R2Config{
		AccountID:         "acct-123",
		AccessKeyID:       "test-access-key-id",
		SecretAccessKey:   "test-secret-access-key",
		BucketName:        "media",
		MaxGrantExpiry:    15 * time.Minute,
		UploadGrantExpiry: 15 * time.Minute,
	}
}

func TestNewProvider_RejectsInvalidConfig(t *testing.T) {
	cfg := testR2Config()
	cfg.BucketName = ""

	_, err := NewProvider(cfg)

	require.Error(t, err)
}

func TestNewProvider_BuildsFromValidConfig(t *testing.T) {
	provider, err := NewProvider(testR2Config())

	require.NoError(t, err)
	assert.Equal(t, "media", provider.bucket)
	assert.Equal(t, 15*time.Minute, provider.maxGrantExpiry)
}

// CreateUploadGrant presigns locally and never touches the network, so it is
// exercised against the real *s3.PresignClient built by NewProvider, with
// fake (non-functional) credentials.
func TestCreateUploadGrant(t *testing.T) {
	provider, err := NewProvider(testR2Config())
	require.NoError(t, err)

	req := storage.UploadRequest{
		Key:    "media/avatars/123.png",
		Expiry: 5 * time.Minute,
	}

	grant, err := provider.CreateUploadGrant(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, grant.Method)
	assert.Equal(t, storage.ObjectKey("media/avatars/123.png"), grant.ObjectKey)
	assert.Contains(t, grant.URL, "acct-123.r2.cloudflarestorage.com")
	assert.Contains(t, grant.URL, "media/avatars/123.png")
	assert.WithinDuration(t, time.Now().Add(5*time.Minute), grant.ExpiresAt, 5*time.Second)

	// The AWS SDK strips Content-Type before signing a presigned PutObject
	// request regardless — see the doc comment on CreateUploadGrant. Only
	// Host ends up signed.
	require.Contains(t, grant.Headers, "Host")
	assert.NotContains(t, grant.Headers, "Content-Type")
}

func TestCreateUploadGrant_ClampsExpiryToConfiguredMax(t *testing.T) {
	provider, err := NewProvider(testR2Config())
	require.NoError(t, err)

	req := storage.UploadRequest{
		Key:    "media/avatars/123.png",
		Expiry: 24 * time.Hour, // far beyond the 15-minute cap
	}

	grant, err := provider.CreateUploadGrant(context.Background(), req)
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now().Add(15*time.Minute), grant.ExpiresAt, 5*time.Second)

	q := grant.URL[strings.Index(grant.URL, "?")+1:]
	assert.Contains(t, q, "X-Amz-Expires=900") // 15 minutes, in seconds
}

func TestCreateUploadGrant_RejectsInvalidRequest(t *testing.T) {
	provider, err := NewProvider(testR2Config())
	require.NoError(t, err)

	_, err = provider.CreateUploadGrant(context.Background(), storage.UploadRequest{})

	require.Error(t, err)
}

func TestCreateUploadGrant_RejectsUnsafeKey(t *testing.T) {
	provider, err := NewProvider(testR2Config())
	require.NoError(t, err)

	req := storage.UploadRequest{
		Key:    "../secrets.env",
		Expiry: time.Minute,
	}

	_, err = provider.CreateUploadGrant(context.Background(), req)

	require.Error(t, err)
}

type fakeHeadObjectAPI struct {
	out *s3.HeadObjectOutput
	err error
}

func (f *fakeHeadObjectAPI) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return f.out, f.err
}

func TestStatObject(t *testing.T) {
	lastModified := time.Date(2024, 3, 1, 8, 30, 0, 0, time.UTC)
	provider := &Provider{
		bucket: "media",
		head: &fakeHeadObjectAPI{
			out: &s3.HeadObjectOutput{
				ContentLength: aws.Int64(2048),
				ContentType:   aws.String("image/png"),
				LastModified:  aws.Time(lastModified),
			},
		},
	}

	got, err := provider.StatObject(context.Background(), "media/avatars/123.png")
	require.NoError(t, err)

	assert.Equal(t, storage.ObjectKey("media/avatars/123.png"), got.Key)
	assert.Equal(t, storage.ObjectMetadata{
		SizeBytes:    2048,
		ContentType:  "image/png",
		LastModified: lastModified,
	}, got.Metadata)
}

func TestStatObject_NotFound(t *testing.T) {
	provider := &Provider{
		bucket: "media",
		head:   &fakeHeadObjectAPI{err: &types.NotFound{}},
	}

	_, err := provider.StatObject(context.Background(), "media/missing.png")

	require.Error(t, err)
}

func TestStatObject_RejectsUnsafeKey(t *testing.T) {
	provider := &Provider{bucket: "media", head: &fakeHeadObjectAPI{}}

	_, err := provider.StatObject(context.Background(), "/leading-slash.png")

	require.Error(t, err)
}

type fakeDeleteObjectAPI struct {
	calledWithKey string
	err           error
}

func (f *fakeDeleteObjectAPI) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.calledWithKey = aws.ToString(params.Key)
	return &s3.DeleteObjectOutput{}, f.err
}

func TestDeleteObject(t *testing.T) {
	fake := &fakeDeleteObjectAPI{}
	provider := &Provider{bucket: "media", del: fake}

	err := provider.DeleteObject(context.Background(), "media/avatars/123.png")

	require.NoError(t, err)
	assert.Equal(t, "media/avatars/123.png", fake.calledWithKey)
}

// DeleteObject on a key the SDK reports as missing is not an error: R2/S3
// delete is idempotent, and the fake here stands in for that native behavior
// by simply returning success (no error) as the real API would.
func TestDeleteObject_MissingKeyIsNotAnError(t *testing.T) {
	provider := &Provider{bucket: "media", del: &fakeDeleteObjectAPI{}}

	err := provider.DeleteObject(context.Background(), "media/never-existed.png")

	require.NoError(t, err)
}

func TestDeleteObject_PropagatesUnexpectedErrors(t *testing.T) {
	provider := &Provider{bucket: "media", del: &fakeDeleteObjectAPI{err: errors.New("access denied")}}

	err := provider.DeleteObject(context.Background(), "media/avatars/123.png")

	require.Error(t, err)
}

func TestDeleteObject_RejectsUnsafeKey(t *testing.T) {
	provider := &Provider{bucket: "media", del: &fakeDeleteObjectAPI{}}

	err := provider.DeleteObject(context.Background(), "")

	require.Error(t, err)
}
