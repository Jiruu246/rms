package r2storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/pkg/storage"
)

// callTimeout bounds every outbound network call R2 makes on behalf of
// StatObject and DeleteObject. CreateUploadGrant is excluded: presigning a
// request is a local computation and never touches the network.
const callTimeout = 10 * time.Second

// headObjectAPI is the subset of *s3.Client used by StatObject, narrowed so
// tests can supply a fake instead of hitting a real bucket.
type headObjectAPI interface {
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// deleteObjectAPI is the subset of *s3.Client used by DeleteObject, narrowed
// so tests can supply a fake instead of hitting a real bucket.
type deleteObjectAPI interface {
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// presignPutObjectAPI is the subset of *s3.PresignClient used by
// CreateUploadGrant, narrowed so tests can supply a fake.
type presignPutObjectAPI interface {
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// Provider implements pkg/storage.StorageProvider against a Cloudflare R2 bucket
type Provider struct {
	head           headObjectAPI
	del            deleteObjectAPI
	presign        presignPutObjectAPI
	bucket         string
	maxGrantExpiry time.Duration
}

func NewProvider(cfg config.R2Config) (*Provider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("r2: %w", err)
	}

	awsCfg := aws.Config{
		Region:      "auto",
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})

	return &Provider{
		head:           client,
		del:            client,
		presign:        s3.NewPresignClient(client),
		bucket:         cfg.BucketName,
		maxGrantExpiry: cfg.MaxGrantExpiry,
	}, nil
}

// CreateUploadGrant issues a short-lived presigned PUT URL for req.Key.
func (p *Provider) CreateUploadGrant(ctx context.Context, req storage.UploadRequest) (*storage.UploadGrant, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	key, err := toS3Key(req.Key)
	if err != nil {
		return nil, err
	}

	expiry := min(req.Expiry, p.maxGrantExpiry)

	signed, err := p.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return nil, fmt.Errorf("r2: create upload grant for %q: %w", req.Key, err)
	}

	headers := make(map[string]string, len(signed.SignedHeader))
	for name, values := range signed.SignedHeader {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return &storage.UploadGrant{
		Method:    signed.Method,
		URL:       signed.URL,
		Headers:   headers,
		ObjectKey: req.Key,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

func (p *Provider) StatObject(ctx context.Context, key storage.ObjectKey) (*storage.StoredObject, error) {
	s3Key, err := toS3Key(key)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	out, err := p.head.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return nil, mapAWSError(err, key)
	}

	return &storage.StoredObject{
		Key:      key,
		Metadata: objectMetadataFromHead(out),
	}, nil
}

func (p *Provider) DeleteObject(ctx context.Context, key storage.ObjectKey) error {
	s3Key, err := toS3Key(key)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	if _, err := p.del.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(s3Key),
	}); err != nil {
		return mapAWSError(err, key)
	}
	return nil
}
