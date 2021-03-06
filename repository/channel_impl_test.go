package repository

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/utils"
	"strconv"
	"testing"
)

func TestRepositoryImpl_UpdateChannelTopic(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupWithUser(t, common)

	cases := []struct {
		topic string
	}{
		{"test"},
		{""},
	}

	for i, v := range cases {
		v := v
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			ch := mustMakeChannel(t, repo, random)
			if assert.NoError(t, repo.UpdateChannelTopic(ch.ID, v.topic, user.ID)) {
				ch, err := repo.GetChannel(ch.ID)
				require.NoError(t, err)
				assert.Equal(t, v.topic, ch.Topic)
			}
		})
	}
}

func TestRepositoryImpl_UpdateChannelAttributes(t *testing.T) {
	t.Parallel()
	repo, _, _ := setup(t, common)

	cases := []struct {
		flag1 bool
		flag2 bool
	}{
		{true, true},
		{true, false},
		{false, true},
		{false, false},
	}

	for i, v := range cases {
		v := v
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			ch := mustMakeChannel(t, repo, random)
			if assert.NoError(t, repo.UpdateChannelAttributes(ch.ID, &v.flag1, &v.flag2)) {
				c, err := repo.GetChannel(ch.ID)
				require.NoError(t, err)
				assert.Equal(t, v.flag1, c.IsVisible)
				assert.Equal(t, v.flag2, c.IsForced)
			}
		})
	}
}

func TestRepositoryImpl_GetChannelByMessageID(t *testing.T) {
	t.Parallel()
	repo, _, _, user, channel := setupWithUserAndChannel(t, common)

	t.Run("Exists", func(t *testing.T) {
		t.Parallel()

		message := mustMakeMessage(t, repo, user.ID, channel.ID)
		ch, err := repo.GetChannelByMessageID(message.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, channel.ID, ch.ID)
		}
	})

	t.Run("NotExists", func(t *testing.T) {
		t.Parallel()

		_, err := repo.GetChannelByMessageID(uuid.Nil)
		assert.Error(t, err)
	})
}

func TestRepositoryImpl_GetChannel(t *testing.T) {
	t.Parallel()
	repo, _, _, channel := setupWithChannel(t, common)

	t.Run("Exists", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		ch, err := repo.GetChannel(channel.ID)
		if assert.NoError(err) {
			assert.Equal(channel.ID, ch.ID)
			assert.Equal(channel.Name, ch.Name)
		}
	})

	t.Run("NotExists", func(t *testing.T) {
		_, err := repo.GetChannel(uuid.Nil)
		assert.Error(t, err)
	})
}

func TestRepositoryImpl_GetChannelPath(t *testing.T) {
	t.Parallel()
	repo, _, _ := setup(t, common)

	ch1 := mustMakeChannelDetail(t, repo, uuid.Nil, random, uuid.Nil)
	ch2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, ch1.ID)
	ch3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, ch2.ID)

	t.Run("ch1", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		path, err := repo.GetChannelPath(ch1.ID)
		if assert.NoError(err) {
			assert.Equal(ch1.Name, path)
		}
	})

	t.Run("ch2", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		path, err := repo.GetChannelPath(ch2.ID)
		if assert.NoError(err) {
			assert.Equal(fmt.Sprintf("%s/%s", ch1.Name, ch2.Name), path)
		}
	})

	t.Run("ch3", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		path, err := repo.GetChannelPath(ch3.ID)
		if assert.NoError(err) {
			assert.Equal(fmt.Sprintf("%s/%s/%s", ch1.Name, ch2.Name, ch3.Name), path)
		}
	})

	t.Run("NotExists1", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		_, err := repo.GetChannelPath(uuid.Nil)
		assert.Error(err)
	})

	t.Run("NotExists2", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		_, err := repo.GetChannelPath(uuid.Must(uuid.NewV4()))
		assert.Error(err)
	})
}

func TestRepositoryImpl_ChangeChannelName(t *testing.T) {
	t.Parallel()
	repo, _, _, parent := setupWithChannel(t, common)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, "test2", parent.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, "test3", c2.ID)
	mustMakeChannelDetail(t, repo, uuid.Nil, "test4", c2.ID)

	t.Run("fail", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		assert.Error(repo.ChangeChannelName(parent.ID, ""))
		assert.Error(repo.ChangeChannelName(parent.ID, "あああ"))
		assert.Error(repo.ChangeChannelName(parent.ID, "test2???"))
	})

	t.Run("c2", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		if assert.NoError(repo.ChangeChannelName(c2.ID, "aiueo")) {
			c, err := repo.GetChannel(c2.ID)
			require.NoError(t, err)
			assert.Equal("aiueo", c.Name)
		}
	})

	t.Run("c3", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		assert.Error(repo.ChangeChannelName(c3.ID, "test4"))
		if assert.NoError(repo.ChangeChannelName(c3.ID, "test2")) {
			c, err := repo.GetChannel(c3.ID)
			require.NoError(t, err)
			assert.Equal("test2", c.Name)
		}
	})
}

func TestRepositoryImpl_ChangeChannelParent(t *testing.T) {
	t.Parallel()
	repo, _, _ := setup(t, common)

	chName := utils.RandAlphabetAndNumberString(20)
	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, chName, uuid.Nil)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, chName, c3.ID)

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, repo.ChangeChannelParent(c4.ID, uuid.Nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		if assert.NoError(t, repo.ChangeChannelParent(c3.ID, uuid.Nil)) {
			c, err := repo.GetChannel(c3.ID)
			require.NoError(t, err)
			assert.Equal(t, uuid.Nil, c.ParentID)
		}
	})
}

func TestRepositoryImpl_DeleteChannel(t *testing.T) {
	t.Parallel()
	repo, _, _, c1 := setupWithChannel(t, common)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c1.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c3.ID)

	if assert.NoError(t, repo.DeleteChannel(c1.ID)) {
		t.Run("c1", func(t *testing.T) {
			t.Parallel()
			_, err := repo.GetChannel(c1.ID)
			assert.Error(t, err)
		})
		t.Run("c2", func(t *testing.T) {
			t.Parallel()
			_, err := repo.GetChannel(c2.ID)
			assert.Error(t, err)
		})
		t.Run("c3", func(t *testing.T) {
			t.Parallel()
			_, err := repo.GetChannel(c3.ID)
			assert.Error(t, err)
		})
		t.Run("c4", func(t *testing.T) {
			t.Parallel()
			_, err := repo.GetChannel(c4.ID)
			assert.Error(t, err)
		})
	}
}

func TestRepositoryImpl_GetChildrenChannelIDs(t *testing.T) {
	t.Parallel()
	repo, _, _, c1 := setupWithChannel(t, common)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c1.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)

	cases := []struct {
		name   string
		ch     uuid.UUID
		expect []uuid.UUID
	}{
		{"c1", c1.ID, []uuid.UUID{c2.ID}},
		{"c2", c2.ID, []uuid.UUID{c3.ID, c4.ID}},
		{"c3", c3.ID, []uuid.UUID{}},
		{"c4", c4.ID, []uuid.UUID{}},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			ids, err := repo.GetChildrenChannelIDs(v.ch)
			if assert.NoError(t, err) {
				assert.ElementsMatch(t, ids, v.expect)
			}
		})
	}
}

func TestRepositoryImpl_GetPrivateChannelMemberIDs(t *testing.T) {
	t.Parallel()
	repo, assert, _ := setup(t, common)

	user1 := mustMakeUser(t, repo, random)
	user2 := mustMakeUser(t, repo, random)
	ch := mustMakePrivateChannel(t, repo, random, []uuid.UUID{user1.ID, user2.ID})

	member, err := repo.GetPrivateChannelMemberIDs(ch.ID)
	if assert.NoError(err) {
		assert.Len(member, 2)
	}
}

func TestRepositoryImpl_SubscribeChannel(t *testing.T) {
	t.Parallel()
	repo, assert, _, user1, ch := setupWithUserAndChannel(t, common)

	if assert.NoError(repo.SubscribeChannel(user1.ID, ch.ID)) {
		assert.Equal(1, count(t, getDB(repo).Model(model.UserSubscribeChannel{}).Where(&model.UserSubscribeChannel{UserID: user1.ID})))
	}
	assert.NoError(repo.SubscribeChannel(user1.ID, ch.ID))
}

func TestRepositoryImpl_UnsubscribeChannel(t *testing.T) {
	t.Parallel()
	repo, _, require := setup(t, common)

	user1 := mustMakeUser(t, repo, random)
	user2 := mustMakeUser(t, repo, random)
	ch1 := mustMakeChannel(t, repo, random)
	ch2 := mustMakeChannel(t, repo, random)
	require.NoError(repo.SubscribeChannel(user1.ID, ch1.ID))
	require.NoError(repo.SubscribeChannel(user1.ID, ch2.ID))
	require.NoError(repo.SubscribeChannel(user2.ID, ch2.ID))

	cases := []struct {
		name   string
		user   uuid.UUID
		ch     uuid.UUID
		expect int
	}{
		{"user2-channel2", user2.ID, ch2.ID, 2},
		{"user1-channel2", user1.ID, ch2.ID, 1},
		{"user1-channel1", user1.ID, ch1.ID, 0},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			if assert.NoError(t, repo.UnsubscribeChannel(v.user, v.ch)) {
				assert.Equal(t, v.expect, count(t, getDB(repo).Model(model.UserSubscribeChannel{}).Where("user_id IN (?, ?)", user1.ID, user2.ID)))
			}
		})
	}
}

func TestRepositoryImpl_GetSubscribingUserIDs(t *testing.T) {
	t.Parallel()
	repo, _, require := setup(t, common)

	user1 := mustMakeUser(t, repo, random)
	user2 := mustMakeUser(t, repo, random)
	ch1 := mustMakeChannel(t, repo, random)
	ch2 := mustMakeChannel(t, repo, random)
	require.NoError(repo.SubscribeChannel(user1.ID, ch1.ID))
	require.NoError(repo.SubscribeChannel(user1.ID, ch2.ID))
	require.NoError(repo.SubscribeChannel(user2.ID, ch2.ID))

	cases := []struct {
		name   string
		ch     uuid.UUID
		expect int
	}{
		{"ch1", ch1.ID, 1},
		{"ch2", ch2.ID, 2},
		{"nil ch", uuid.Nil, 0},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			arr, err := repo.GetSubscribingUserIDs(v.ch)
			if assert.NoError(t, err) {
				assert.Len(t, arr, v.expect)
			}
		})
	}
}

func TestRepositoryImpl_GetSubscribedChannelIDs(t *testing.T) {
	t.Parallel()
	repo, _, require := setup(t, common)

	user1 := mustMakeUser(t, repo, random)
	user2 := mustMakeUser(t, repo, random)
	ch1 := mustMakeChannel(t, repo, random)
	ch2 := mustMakeChannel(t, repo, random)
	require.NoError(repo.SubscribeChannel(user1.ID, ch1.ID))
	require.NoError(repo.SubscribeChannel(user1.ID, ch2.ID))
	require.NoError(repo.SubscribeChannel(user2.ID, ch2.ID))

	cases := []struct {
		name   string
		user   uuid.UUID
		expect int
	}{
		{"user1", user1.ID, 2},
		{"user2", user2.ID, 1},
		{"nil user", uuid.Nil, 0},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			arr, err := repo.GetSubscribedChannelIDs(v.user)
			if assert.NoError(t, err) {
				assert.Len(t, arr, v.expect)
			}
		})
	}
}

func TestRepositoryImpl_CreatePublicChannel(t *testing.T) {
	t.Parallel()
	repo, assert, _, user := setupWithUser(t, common)

	name := utils.RandAlphabetAndNumberString(20)
	c, err := repo.CreatePublicChannel(name, uuid.Nil, user.ID)
	if assert.NoError(err) {
		assert.NotEmpty(c.ID)
		assert.Equal(name, c.Name)
		assert.Equal(user.ID, c.CreatorID)
		assert.EqualValues(uuid.Nil, c.ParentID)
		assert.True(c.IsPublic)
		assert.True(c.IsVisible)
		assert.False(c.IsForced)
		assert.Equal(user.ID, c.UpdaterID)
		assert.Empty(c.Topic)
		assert.NotZero(c.CreatedAt)
		assert.NotZero(c.UpdatedAt)
		assert.Nil(c.DeletedAt)
	}

	_, err = repo.CreatePublicChannel(name, uuid.Nil, user.ID)
	assert.Equal(ErrAlreadyExists, err)

	_, err = repo.CreatePublicChannel("ああああ", uuid.Nil, user.ID)
	assert.Error(err)

	c2, err := repo.CreatePublicChannel("Parent2", c.ID, user.ID)
	assert.NoError(err)
	c3, err := repo.CreatePublicChannel("Parent3", c2.ID, user.ID)
	assert.NoError(err)
	c4, err := repo.CreatePublicChannel("Parent4", c3.ID, user.ID)
	assert.NoError(err)
	_, err = repo.CreatePublicChannel("Parent4", c3.ID, user.ID)
	assert.Equal(ErrAlreadyExists, err)
	c5, err := repo.CreatePublicChannel("Parent5", c4.ID, user.ID)
	assert.NoError(err)
	_, err = repo.CreatePublicChannel("Parent6", c5.ID, user.ID)
	assert.Equal(ErrChannelDepthLimitation, err)
}

func TestRepositoryImpl_getParentChannel(t *testing.T) {
	t.Parallel()
	r, _, _ := setup(t, common)
	repo := r.(*GormRepository)

	parentChannel := mustMakeChannel(t, repo, random)
	childChannel := mustMakeChannelDetail(t, repo, uuid.Nil, random, parentChannel.ID)

	t.Run("child", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		parent, err := repo.getParentChannel(repo.db, childChannel.ID)
		if assert.NoError(err) {
			assert.Equal(parent.ID, parentChannel.ID)
		}
	})

	t.Run("parent", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		parent, err := repo.getParentChannel(repo.db, parentChannel.ID)
		if assert.NoError(err) {
			assert.Nil(parent)
		}
	})

	t.Run("NotExists1", func(t *testing.T) {
		t.Parallel()

		_, err := repo.getParentChannel(repo.db, uuid.Nil)
		assert.Error(t, err)
	})

	t.Run("NotExists2", func(t *testing.T) {
		t.Parallel()

		_, err := repo.getParentChannel(repo.db, uuid.Must(uuid.NewV4()))
		assert.Error(t, err)
	})
}

func TestRepositoryImpl_getDescendantChannelIDs(t *testing.T) {
	t.Parallel()
	r, _, _, c1 := setupWithChannel(t, common)
	repo := r.(*GormRepository)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c1.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c5 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c3.ID)

	cases := []struct {
		name   string
		ch     uuid.UUID
		expect []uuid.UUID
	}{
		{"c1", c1.ID, []uuid.UUID{c2.ID, c3.ID, c4.ID, c5.ID}},
		{"c2", c2.ID, []uuid.UUID{c3.ID, c4.ID, c5.ID}},
		{"c3", c3.ID, []uuid.UUID{c5.ID}},
		{"c4", c4.ID, []uuid.UUID{}},
		{"c5", c5.ID, []uuid.UUID{}},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			ids, err := repo.getDescendantChannelIDs(repo.db, v.ch)
			if assert.NoError(t, err) {
				assert.ElementsMatch(t, ids, v.expect)
			}
		})
	}
}

func TestRepositoryImpl_getAscendantChannelIDs(t *testing.T) {
	t.Parallel()
	r, _, _, c1 := setupWithChannel(t, common)
	repo := r.(*GormRepository)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c1.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c5 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c3.ID)

	cases := []struct {
		name   string
		ch     uuid.UUID
		expect []uuid.UUID
	}{
		{"c1", c1.ID, []uuid.UUID{}},
		{"c2", c2.ID, []uuid.UUID{c1.ID}},
		{"c3", c3.ID, []uuid.UUID{c1.ID, c2.ID}},
		{"c4", c4.ID, []uuid.UUID{c1.ID, c2.ID}},
		{"c5", c5.ID, []uuid.UUID{c1.ID, c2.ID, c3.ID}},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			ids, err := repo.getAscendantChannelIDs(repo.db, v.ch)
			if assert.NoError(t, err) {
				assert.ElementsMatch(t, ids, v.expect)
			}
		})
	}
}

func TestRepositoryImpl_getChannelDepth(t *testing.T) {
	t.Parallel()
	r, _, _, c1 := setupWithChannel(t, common)
	repo := r.(*GormRepository)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c1.ID)
	c3 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c4 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c2.ID)
	c5 := mustMakeChannelDetail(t, repo, uuid.Nil, random, c3.ID)

	cases := []struct {
		name string
		ch   uuid.UUID
		num  int
	}{
		{"c1", c1.ID, 4},
		{"c2", c2.ID, 3},
		{"c3", c3.ID, 2},
		{"c4", c4.ID, 1},
		{"c5", c5.ID, 1},
	}

	for _, v := range cases {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			d, err := repo.getChannelDepth(repo.db, v.ch)
			if assert.NoError(t, err) {
				assert.Equal(t, v.num, d)
			}
		})
	}
}

func TestRepositoryImpl_isChannelPresent(t *testing.T) {
	t.Parallel()
	r, _, _, parent := setupWithChannel(t, common)
	repo := r.(*GormRepository)

	c2 := mustMakeChannelDetail(t, repo, uuid.Nil, "test2", parent.ID)
	mustMakeChannelDetail(t, repo, uuid.Nil, "test3", c2.ID)

	cases := []struct {
		parentID uuid.UUID
		name     string
		expect   bool
	}{
		{parent.ID, "test2", true},
		{parent.ID, "test3", false},
		{c2.ID, "test3", true},
		{c2.ID, "test4", false},
	}

	for i, v := range cases {
		v := v
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ok, err := repo.isChannelPresent(repo.db, v.name, v.parentID)
			if assert.NoError(t, err) {
				assert.Equal(t, v.expect, ok)
			}
		})
	}
}
