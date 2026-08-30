package r2storage

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/pkg/storage"
)

// toS3Key maps a provider-neutral ObjectKey to the literal S3 object key.
//
// This is a defense-in-depth guard, not a source of key generation: callers
// (the service layer) are expected to generate keys themselves, never from
// unvalidated client input. This only rejects keys that a well-behaved
// caller should never produce.
func toS3Key(key storage.ObjectKey) (string, error) {
	if err := key.Validate(); err != nil {
		return "", err
	}
	s := string(key)
	if strings.HasPrefix(s, "/") {
		return "", apperr.Invalid("object key %q must not start with '/'", s)
	}
	for segment := range strings.SplitSeq(s, "/") {
		if segment == ".." {
			return "", apperr.Invalid("object key %q must not contain '..' segments", s)
		}
	}
	return s, nil
}

// objectMetadataFromHead maps a HeadObject response to provider-neutral
// ObjectMetadata.
func objectMetadataFromHead(out *s3.HeadObjectOutput) storage.ObjectMetadata {
	return storage.ObjectMetadata{
		SizeBytes:    aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		LastModified: aws.ToTime(out.LastModified),
	}
}

// httpStatusCoder is satisfied by the AWS SDK's HTTP response error types.
// Declared locally so mapAWSError doesn't need to import smithy's transport
// package just for this one check.
type httpStatusCoder interface {
	HTTPStatusCode() int
}

// mapAWSError translates an error from an S3-compatible SDK call into the
// application's error vocabulary. A missing object maps to apperr.ErrNotFound
// regardless of whether R2 reported it via the exact AWS NotFound shape or
// just a bare 404 status (R2's error bodies are not guaranteed to match AWS's
// byte-for-byte). Anything else is wrapped as an opaque infrastructure error.
func mapAWSError(err error, key storage.ObjectKey) error {
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return apperr.NotFound("object %q", key)
	}

	var statusErr httpStatusCoder
	if errors.As(err, &statusErr) && statusErr.HTTPStatusCode() == http.StatusNotFound {
		return apperr.NotFound("object %q", key)
	}

	return fmt.Errorf("r2: object %q: %w", key, err)
}
