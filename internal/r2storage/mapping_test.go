package r2storage

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/pkg/storage"
)

func TestToS3Key(t *testing.T) {
	tests := []struct {
		name    string
		key     storage.ObjectKey
		want    string
		wantErr bool
	}{
		{name: "simple key", key: "media/avatars/123.png", want: "media/avatars/123.png"},
		{name: "empty key", key: "", wantErr: true},
		{name: "leading slash", key: "/media/avatars/123.png", wantErr: true},
		{name: "traversal segment", key: "media/../secrets.env", wantErr: true},
		{name: "traversal at start", key: "../secrets.env", wantErr: true},
		{name: "double dot inside a segment is fine", key: "media/..avatars../123.png", want: "media/..avatars../123.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toS3Key(tt.key)
			if tt.wantErr {
				assert.ErrorIs(t, err, apperr.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestObjectMetadataFromHead(t *testing.T) {
	lastModified := time.Date(2024, 3, 1, 8, 30, 0, 0, time.UTC)

	got := objectMetadataFromHead(&s3.HeadObjectOutput{
		ContentLength: aws.Int64(2048),
		ContentType:   aws.String("image/png"),
		LastModified:  aws.Time(lastModified),
	})

	assert.Equal(t, storage.ObjectMetadata{
		SizeBytes:    2048,
		ContentType:  "image/png",
		LastModified: lastModified,
	}, got)
}

func TestObjectMetadataFromHead_NilFields(t *testing.T) {
	got := objectMetadataFromHead(&s3.HeadObjectOutput{})

	assert.Equal(t, storage.ObjectMetadata{}, got)
}

func TestMapAWSError(t *testing.T) {
	key := storage.ObjectKey("media/avatars/123.png")

	tests := []struct {
		name         string
		err          error
		wantNotFound bool
	}{
		{
			name:         "types.NotFound",
			err:          &types.NotFound{Message: aws.String("not found")},
			wantNotFound: true,
		},
		{
			name: "generic 404 status",
			err: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: 404}},
				Err:      errors.New("NoSuchKey"),
			},
			wantNotFound: true,
		},
		{
			name: "generic 500 status is not mapped to not found",
			err: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: 500}},
				Err:      errors.New("InternalError"),
			},
			wantNotFound: false,
		},
		{
			name:         "unrelated error",
			err:          errors.New("network unreachable"),
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapAWSError(tt.err, key)
			if tt.wantNotFound {
				assert.ErrorIs(t, got, apperr.ErrNotFound)
				return
			}
			assert.NotErrorIs(t, got, apperr.ErrNotFound)
			assert.ErrorIs(t, got, tt.err)
			assert.Contains(t, got.Error(), string(key))
		})
	}
}
