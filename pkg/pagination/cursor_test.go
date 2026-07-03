package pagination_test

import (
	"testing"

	"github.com/Jiruu246/rms/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCursor_DecodeCursor_RoundTrip(t *testing.T) {
	original := pagination.Cursor{
		Sort: []pagination.SortSpec{
			{Field: "created_at", Desc: true},
			{Field: "name", Desc: false},
		},
		Values: map[string]any{
			"created_at": "2024-01-01T12:00:00Z",
			"name":       "some-name",
		},
		ID: uuid.New().String(),
	}

	token, err := pagination.EncodeCursor(original)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	decoded, err := pagination.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Sort, decoded.Sort)
	assert.Equal(t, original.Values["name"], decoded.Values["name"])
}

func TestDecodeCursor_InvalidToken(t *testing.T) {
	_, err := pagination.DecodeCursor("not-valid-base64!!")
	assert.Error(t, err)
}
