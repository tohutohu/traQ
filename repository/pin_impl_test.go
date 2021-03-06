package repository

import (
	"github.com/gofrs/uuid"
	"testing"
)

func TestRepositoryImpl_CreatePin(t *testing.T) {
	t.Parallel()
	repo, assert, _, user, channel := setupWithUserAndChannel(t, common)

	testMessage := mustMakeMessage(t, repo, user.ID, channel.ID)

	_, err := repo.CreatePin(uuid.Nil, user.ID)
	assert.Error(err)

	_, err = repo.CreatePin(testMessage.ID, uuid.Nil)
	assert.Error(err)

	p, err := repo.CreatePin(testMessage.ID, user.ID)
	if assert.NoError(err) {
		assert.NotEmpty(p)
	}

	p2, err := repo.CreatePin(testMessage.ID, user.ID)
	if assert.NoError(err) {
		assert.EqualValues(p, p2)
	}
}

func TestRepositoryImpl_GetPin(t *testing.T) {
	t.Parallel()
	repo, assert, _, user, channel := setupWithUserAndChannel(t, common)

	testMessage := mustMakeMessage(t, repo, user.ID, channel.ID)
	p := mustMakePin(t, repo, testMessage.ID, user.ID)

	pin, err := repo.GetPin(p)
	if assert.NoError(err) {
		assert.Equal(p, pin.ID)
		assert.Equal(testMessage.ID, pin.MessageID)
		assert.Equal(user.ID, pin.UserID)
		assert.NotZero(pin.CreatedAt)
		assert.NotZero(pin.Message)
	}

	_, err = repo.GetPin(uuid.Nil)
	assert.Equal(ErrNotFound, err)

	_, err = repo.GetPin(uuid.Must(uuid.NewV4()))
	assert.Equal(ErrNotFound, err)
}

func TestRepositoryImpl_IsPinned(t *testing.T) {
	t.Parallel()
	repo, assert, _, user, channel := setupWithUserAndChannel(t, common)

	testMessage := mustMakeMessage(t, repo, user.ID, channel.ID)
	mustMakePin(t, repo, testMessage.ID, user.ID)

	ok, err := repo.IsPinned(testMessage.ID)
	if assert.NoError(err) {
		assert.True(ok)
	}

	ok, err = repo.IsPinned(uuid.Nil)
	if assert.NoError(err) {
		assert.False(ok)
	}

	ok, err = repo.IsPinned(uuid.Must(uuid.NewV4()))
	if assert.NoError(err) {
		assert.False(ok)
	}
}

func TestRepositoryImpl_DeletePin(t *testing.T) {
	t.Parallel()
	repo, assert, _, user, channel := setupWithUserAndChannel(t, common)

	testMessage := mustMakeMessage(t, repo, user.ID, channel.ID)
	p := mustMakePin(t, repo, testMessage.ID, user.ID)

	assert.Error(repo.DeletePin(uuid.Nil))

	if assert.NoError(repo.DeletePin(p)) {
		_, err := repo.GetPin(p)
		assert.Equal(ErrNotFound, err)
	}

	assert.NoError(repo.DeletePin(uuid.Must(uuid.NewV4())))
}

func TestRepositoryImpl_GetPinsByChannelID(t *testing.T) {
	t.Parallel()
	repo, assert, _, user, channel := setupWithUserAndChannel(t, common)

	testMessage := mustMakeMessage(t, repo, user.ID, channel.ID)
	mustMakePin(t, repo, testMessage.ID, user.ID)

	pins, err := repo.GetPinsByChannelID(channel.ID)
	if assert.NoError(err) {
		assert.Len(pins, 1)
	}

	pins, err = repo.GetPinsByChannelID(uuid.Nil)
	if assert.NoError(err) {
		assert.Empty(pins)
	}
}
