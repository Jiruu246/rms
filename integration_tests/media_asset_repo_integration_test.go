package integration_tests

import (
	"testing"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// MediaAssetRepositoryTestSuite exercises MediaAssetRepository directly
// against a real Postgres instance — see MediaUploadRepositoryTestSuite for
// why this bypasses the (nonexistent) HTTP layer.
type MediaAssetRepositoryTestSuite struct {
	IntegrationTestSuite
}

func TestMediaAssetRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MediaAssetRepositoryTestSuite))
}

func (s *MediaAssetRepositoryTestSuite) TestGetByIDAndGetByUploadID() {
	asset, err := SetupMediaAsset(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaAssetRepository(s.client)

	byID, err := repo.GetByID(s.T().Context(), asset.ID)
	s.Require().NoError(err)
	s.Equal(asset.StorageKey, byID.StorageKey)

	byUpload, err := repo.GetByUploadID(s.T().Context(), asset.UploadID)
	s.Require().NoError(err)
	s.Equal(asset.ID, byUpload.ID)
}

func (s *MediaAssetRepositoryTestSuite) TestGetByID_NotFound() {
	repo := repos.NewEntMediaAssetRepository(s.client)
	_, err := repo.GetByID(s.T().Context(), uuid.New())
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *MediaAssetRepositoryTestSuite) TestDelete_SoftDeletesAndHidesFromGetByID() {
	asset, err := SetupMediaAsset(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaAssetRepository(s.client)
	err = repo.Delete(s.T().Context(), asset.ID)
	s.Require().NoError(err)

	_, err = repo.GetByID(s.T().Context(), asset.ID)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)

	// The row itself still exists, just marked deleted — this is a soft
	// delete, not a hard delete; orphan/storage cleanup is a later stage.
	raw, err := s.client.MediaAsset.Get(s.T().Context(), asset.ID)
	s.Require().NoError(err)
	s.Equal("deleted", raw.Status.String())
}

func (s *MediaAssetRepositoryTestSuite) TestDelete_AlreadyDeleted_NotFound() {
	asset, err := SetupMediaAsset(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaAssetRepository(s.client)
	s.Require().NoError(repo.Delete(s.T().Context(), asset.ID))

	err = repo.Delete(s.T().Context(), asset.ID)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)
}

// TestUploadIDUniqueConstraint verifies the DB-level guard directly: even
// bypassing MediaUploadRepository.Consume (which never issues a second
// insert for the same upload because the status check happens first), the
// unique index on upload_id independently rejects a second media_assets row
// referencing the same upload.
func (s *MediaAssetRepositoryTestSuite) TestUploadIDUniqueConstraint() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	_, err = s.client.MediaAsset.Create().
		SetUploadedByUserID(upload.OwnerID).
		SetUploadID(upload.ID).
		SetStorageKey(upload.ObjectKey).
		SetContentType("image/png").
		SetSizeBytes(1024).
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = s.client.MediaAsset.Create().
		SetUploadedByUserID(upload.OwnerID).
		SetUploadID(upload.ID).
		SetStorageKey("uploads/"+upload.ID.String()+"-duplicate").
		SetContentType("image/png").
		SetSizeBytes(1024).
		Save(s.T().Context())
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))
}
