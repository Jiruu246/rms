package storage_test

import (
	"testing"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestObjectKey_Validate(t *testing.T) {
	tests := []struct {
		name    string
		key     storage.ObjectKey
		wantErr bool
	}{
		{name: "empty", key: "", wantErr: true},
		{name: "non-empty", key: "media/avatars/123.png", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, apperr.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
