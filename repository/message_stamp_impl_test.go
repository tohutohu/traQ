package repository

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/traQ/model"
	"testing"
)

func TestRepositoryImpl_AddStampToMessage(t *testing.T) {
	t.Parallel()
	repo, _, _, user, channel := setupWithUserAndChannel(t, common)

	message := mustMakeMessage(t, repo, user.ID, channel.ID)
	stamp := mustMakeStamp(t, repo, random, uuid.Nil)

	t.Run("Nil id", func(t *testing.T) {
		t.Parallel()

		_, err := repo.AddStampToMessage(uuid.Nil, uuid.Nil, uuid.Nil)
		assert.EqualError(t, err, ErrNilID.Error())
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		assert, _ := assertAndRequire(t)
		{
			ms, err := repo.AddStampToMessage(message.ID, stamp.ID, user.ID)
			if assert.NoError(err) {
				assert.Equal(message.ID, ms.MessageID)
				assert.Equal(stamp.ID, ms.StampID)
				assert.Equal(user.ID, ms.UserID)
				assert.Equal(1, ms.Count)
				assert.NotEmpty(ms.CreatedAt)
				assert.NotEmpty(ms.UpdatedAt)
			}
		}
		{
			ms, err := repo.AddStampToMessage(message.ID, stamp.ID, user.ID)
			if assert.NoError(err) {
				assert.Equal(message.ID, ms.MessageID)
				assert.Equal(stamp.ID, ms.StampID)
				assert.Equal(user.ID, ms.UserID)
				assert.Equal(2, ms.Count)
				assert.NotEmpty(ms.CreatedAt)
				assert.NotEmpty(ms.UpdatedAt)
			}
		}
	})
}

func TestRepositoryImpl_RemoveStampFromMessage(t *testing.T) {
	t.Parallel()
	repo, _, _, user, channel := setupWithUserAndChannel(t, common)

	message := mustMakeMessage(t, repo, user.ID, channel.ID)
	stamp := mustMakeStamp(t, repo, random, uuid.Nil)

	t.Run("Nil id", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, repo.RemoveStampFromMessage(message.ID, stamp.ID, uuid.Nil), ErrNilID.Error())
		assert.EqualError(t, repo.RemoveStampFromMessage(message.ID, uuid.Nil, user.ID), ErrNilID.Error())
		assert.EqualError(t, repo.RemoveStampFromMessage(uuid.Nil, stamp.ID, user.ID), ErrNilID.Error())
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		mustAddMessageStamp(t, repo, message.ID, stamp.ID, user.ID)
		mustAddMessageStamp(t, repo, message.ID, stamp.ID, user.ID)

		if assert.NoError(t, repo.RemoveStampFromMessage(message.ID, stamp.ID, user.ID)) {
			assert.Equal(t, 0, count(t, getDB(repo).Model(&model.MessageStamp{}).Where(&model.MessageStamp{MessageID: message.ID, StampID: stamp.ID, UserID: user.ID})))
		}
	})
}

func TestRepositoryImpl_GetMessageStamps(t *testing.T) {
	t.Parallel()
	repo, _, _, user, channel := setupWithUserAndChannel(t, common)

	message := mustMakeMessage(t, repo, user.ID, channel.ID)
	stamp1 := mustMakeStamp(t, repo, random, uuid.Nil)
	stamp2 := mustMakeStamp(t, repo, random, uuid.Nil)
	mustAddMessageStamp(t, repo, message.ID, stamp1.ID, user.ID)
	mustAddMessageStamp(t, repo, message.ID, stamp2.ID, user.ID)
	mustAddMessageStamp(t, repo, message.ID, stamp1.ID, user.ID)

	t.Run("Nil id", func(t *testing.T) {
		t.Parallel()
		ms, err := repo.GetMessageStamps(uuid.Nil)
		if assert.NoError(t, err) {
			assert.Empty(t, ms)
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		ms, err := repo.GetMessageStamps(message.ID)
		if assert.NoError(t, err) {
			assert.Len(t, ms, 2)
		}
	})
}

func TestRepositoryImpl_GetUserStampHistory(t *testing.T) {
	t.Parallel()
	repo, _, _, user, channel := setupWithUserAndChannel(t, common)

	message := mustMakeMessage(t, repo, user.ID, channel.ID)
	stamp1 := mustMakeStamp(t, repo, random, uuid.Nil)
	stamp2 := mustMakeStamp(t, repo, random, uuid.Nil)
	stamp3 := mustMakeStamp(t, repo, random, uuid.Nil)
	mustAddMessageStamp(t, repo, message.ID, stamp1.ID, user.ID)
	mustAddMessageStamp(t, repo, message.ID, stamp3.ID, user.ID)
	mustAddMessageStamp(t, repo, message.ID, stamp2.ID, user.ID)

	t.Run("Nil id", func(t *testing.T) {
		t.Parallel()
		ms, err := repo.GetUserStampHistory(uuid.Nil)
		if assert.NoError(t, err) {
			assert.Empty(t, ms)
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		ms, err := repo.GetUserStampHistory(user.ID)
		if assert.NoError(t, err) && assert.Len(t, ms, 3) {
			assert.Equal(t, ms[0].StampID, stamp2.ID)
			assert.Equal(t, ms[1].StampID, stamp3.ID)
			assert.Equal(t, ms[2].StampID, stamp1.ID)
		}
	})
}
