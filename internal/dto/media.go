package dto

import (
	"time"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/pkg/storage"
	"github.com/google/uuid"
)

type CreateUploadResult struct {
	UploadID    uuid.UUID         `json:"upload_id"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Upload      UploadGrant       `json:"upload"`
	Constraints UploadConstraints `json:"constraints"`
}

type UploadConstraints struct {
	MaxSizeBytes        int64    `json:"max_size_bytes"`
	AllowedContentTypes []string `json:"allowed_content_types"`
}

type UploadGrant struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ObjectKey string            `json:"-"`
	ExpiresAt time.Time         `json:"-"`
}

func NewUploadGrant(g *storage.UploadGrant) UploadGrant {
	return UploadGrant{
		Method:    g.Method,
		URL:       g.URL,
		Headers:   g.Headers,
		ObjectKey: string(g.ObjectKey),
		ExpiresAt: g.ExpiresAt,
	}
}

type MediaAsset struct {
	ID          uuid.UUID `json:"id"`
	StorageKey  string    `json:"-"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Status      string    `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func NewMediaAsset(a *ent.MediaAsset) *MediaAsset {
	asset := &MediaAsset{
		ID:          a.ID,
		StorageKey:  a.StorageKey,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
		Status:      a.Status.String(),
		CreateTime:  a.CreateTime,
		UpdateTime:  a.UpdateTime,
	}
	return asset
}
