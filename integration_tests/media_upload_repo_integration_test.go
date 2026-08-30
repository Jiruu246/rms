package integration_tests

import (
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// MediaUploadRepositoryTestSuite exercises MediaUploadRepository directly
// against a real Postgres instance — there is no HTTP handler for media
// uploads yet (this stage only adds persistence), so unlike the other
// integration suites this one calls the repository, not the server.
type MediaUploadRepositoryTestSuite struct {
	IntegrationTestSuite
}

func TestMediaUploadRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MediaUploadRepositoryTestSuite))
}

func (s *MediaUploadRepositoryTestSuite) TestCreateAndGetByID() {
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)

	created, err := repo.Create(s.T().Context(), repos.CreateMediaUploadParams{
		OwnerID:   user.ID,
		Purpose:   mediaupload.PurposeMenuItemImage,
		ObjectKey: "uploads/" + uuid.NewString(),
		ExpiresAt: time.Now().Add(time.Hour),
	})
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusIssued, created.Status)

	fetched, err := repo.GetByID(s.T().Context(), user.ID, created.ID)
	s.Require().NoError(err)
	s.Equal(created.ID, fetched.ID)
	s.Equal(created.ObjectKey, fetched.ObjectKey)
}

func (s *MediaUploadRepositoryTestSuite) TestCreate_DuplicateObjectKey_Conflict() {
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	key := "uploads/" + uuid.NewString()

	_, err = repo.Create(s.T().Context(), repos.CreateMediaUploadParams{
		OwnerID:   user.ID,
		Purpose:   mediaupload.PurposeMenuItemImage,
		ObjectKey: key,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	s.Require().NoError(err)

	_, err = repo.Create(s.T().Context(), repos.CreateMediaUploadParams{
		OwnerID:   user.ID,
		Purpose:   mediaupload.PurposeMenuItemImage,
		ObjectKey: key,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrConflict)
}

func (s *MediaUploadRepositoryTestSuite) TestGetByID_WrongOwner_NotFound() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	otherUser, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	_, err = repo.GetByID(s.T().Context(), otherUser.ID, upload.ID)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)
}

func (s *MediaUploadRepositoryTestSuite) TestConsume_Success() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	asset, err := repo.Consume(s.T().Context(), upload.OwnerID, upload.ID, repos.ConsumeMediaUploadParams{
		ContentType: "image/png",
		SizeBytes:   4096,
	})
	s.Require().NoError(err)
	s.Equal(upload.ID, asset.UploadID)
	s.Equal(upload.ObjectKey, asset.StorageKey)
	s.Equal(upload.OwnerID, asset.UploadedByUserID)
	s.Equal(int64(4096), asset.SizeBytes)

	updatedUpload, err := s.client.MediaUpload.Get(s.T().Context(), upload.ID)
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusConsumed, updatedUpload.Status)
}

func (s *MediaUploadRepositoryTestSuite) TestConsume_Twice_Conflict() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	params := repos.ConsumeMediaUploadParams{ContentType: "image/png", SizeBytes: 1024}

	_, err = repo.Consume(s.T().Context(), upload.OwnerID, upload.ID, params)
	s.Require().NoError(err)

	_, err = repo.Consume(s.T().Context(), upload.OwnerID, upload.ID, params)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrConflict)
}

func (s *MediaUploadRepositoryTestSuite) TestConsume_WrongOwner_NotFound() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	otherUser, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	_, err = repo.Consume(s.T().Context(), otherUser.ID, upload.ID, repos.ConsumeMediaUploadParams{
		ContentType: "image/png",
		SizeBytes:   1024,
	})
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)

	// The session must still be issued — a failed consume attempt by a
	// non-owner must not affect its state.
	unchanged, err := s.client.MediaUpload.Get(s.T().Context(), upload.ID)
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusIssued, unchanged.Status)
}

func (s *MediaUploadRepositoryTestSuite) TestConsume_Expired_Conflict() {
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	expired, err := s.client.MediaUpload.Create().
		SetOwnerID(user.ID).
		SetPurpose(mediaupload.PurposeMenuItemImage).
		SetObjectKey("uploads/" + uuid.NewString()).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Save(s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	_, err = repo.Consume(s.T().Context(), user.ID, expired.ID, repos.ConsumeMediaUploadParams{
		ContentType: "image/png",
		SizeBytes:   1024,
	})
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrConflict)
}

func (s *MediaUploadRepositoryTestSuite) TestFail_Success() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	err = repo.Fail(s.T().Context(), upload.OwnerID, upload.ID)
	s.Require().NoError(err)

	updated, err := s.client.MediaUpload.Get(s.T().Context(), upload.ID)
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusFailed, updated.Status)
}

func (s *MediaUploadRepositoryTestSuite) TestFail_AlreadyConsumed_Conflict() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	_, err = repo.Consume(s.T().Context(), upload.OwnerID, upload.ID, repos.ConsumeMediaUploadParams{
		ContentType: "image/png",
		SizeBytes:   1024,
	})
	s.Require().NoError(err)

	err = repo.Fail(s.T().Context(), upload.OwnerID, upload.ID)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrConflict)

	// A losing Fail attempt must not clobber the already-consumed status.
	unchanged, err := s.client.MediaUpload.Get(s.T().Context(), upload.ID)
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusConsumed, unchanged.Status)
}

func (s *MediaUploadRepositoryTestSuite) TestFail_WrongOwner_NotFound() {
	upload, err := SetupMediaUpload(s.client, s.T().Context())
	s.Require().NoError(err)

	otherUser, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)

	repo := repos.NewEntMediaUploadRepository(s.client)
	err = repo.Fail(s.T().Context(), otherUser.ID, upload.ID)
	s.Require().Error(err)
	s.ErrorIs(err, apperr.ErrNotFound)

	unchanged, err := s.client.MediaUpload.Get(s.T().Context(), upload.ID)
	s.Require().NoError(err)
	s.Equal(mediaupload.StatusIssued, unchanged.Status)
}
