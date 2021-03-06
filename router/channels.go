package router

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/traQ/repository"
	"net/http"

	"github.com/labstack/echo"
	"github.com/traPtitech/traQ/model"
)

// PostChannel リクエストボディ用構造体
type PostChannel struct {
	Name    string      `json:"name"`
	Parent  uuid.UUID   `json:"parent"`
	Private bool        `json:"private"`
	Members []uuid.UUID `json:"member"`
}

// GetChannels GET /channels
func (h *Handlers) GetChannels(c echo.Context) error {
	userID := getRequestUserID(c)

	channelList, err := h.Repo.GetChannelsByUserID(userID)
	if err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}

	chMap := make(map[string]*channelResponse, len(channelList))
	for _, ch := range channelList {
		entry, ok := chMap[ch.ID.String()]
		if !ok {
			entry = &channelResponse{}
			chMap[ch.ID.String()] = entry
		}

		entry.ChannelID = ch.ID.String()
		entry.Name = ch.Name
		entry.Visibility = ch.IsVisible
		entry.Topic = ch.Topic
		entry.Force = ch.IsForced
		entry.Private = !ch.IsPublic
		entry.DM = ch.IsDMChannel()

		if !ch.IsPublic {
			// プライベートチャンネルのメンバー取得
			member, err := h.Repo.GetPrivateChannelMemberIDs(ch.ID)
			if err != nil {
				return internalServerError(err, h.requestContextLogger(c))
			}
			entry.Member = member
		}

		if ch.ParentID != uuid.Nil {
			entry.Parent = ch.ParentID.String()
			parent, ok := chMap[ch.ParentID.String()]
			if !ok {
				parent = &channelResponse{
					ChannelID: ch.ParentID.String(),
				}
				chMap[ch.ParentID.String()] = parent
			}
			parent.Children = append(parent.Children, ch.ID)
		} else {
			parent, ok := chMap[""]
			if !ok {
				parent = &channelResponse{}
				chMap[""] = parent
			}
			parent.Children = append(parent.Children, ch.ID)
		}
	}

	res := make([]*channelResponse, 0, len(chMap))
	for _, v := range chMap {
		res = append(res, v)
	}
	return c.JSON(http.StatusOK, res)
}

// PostChannels POST /channels
func (h *Handlers) PostChannels(c echo.Context) error {
	userID := getRequestUserID(c)

	req := PostChannel{}
	if err := bindAndValidate(c, &req); err != nil {
		return badRequest(err)
	}

	// 親チャンネルがユーザーから見えないと作成できない
	if req.Parent != uuid.Nil {
		if ok, err := h.Repo.IsChannelAccessibleToUser(userID, req.Parent); err != nil {
			return internalServerError(err, h.requestContextLogger(c))
		} else if !ok {
			return badRequest("the parent channel doesn't exist")
		}
	}

	var (
		ch  *model.Channel
		err error
	)

	if req.Private {
		// 非公開チャンネル
		ch, err = h.Repo.CreatePrivateChannel(req.Name, userID, req.Members)
		if err != nil {
			switch {
			case repository.IsArgError(err):
				return badRequest(err)
			case err == repository.ErrAlreadyExists:
				return conflict("channel name conflicts")
			case err == repository.ErrChannelDepthLimitation:
				return badRequest("channel depth limit exceeded")
			case err == repository.ErrForbidden:
				return forbidden("invalid parent channel")
			default:
				return internalServerError(err, h.requestContextLogger(c))
			}
		}
	} else {
		// 公開チャンネル
		ch, err = h.Repo.CreatePublicChannel(req.Name, req.Parent, userID)
		if err != nil {
			switch {
			case repository.IsArgError(err):
				return badRequest(err)
			case err == repository.ErrAlreadyExists:
				return conflict("channel name conflicts")
			case err == repository.ErrChannelDepthLimitation:
				return badRequest("channel depth limit exceeded")
			case err == repository.ErrForbidden:
				return forbidden("invalid parent channel")
			default:
				return internalServerError(err, h.requestContextLogger(c))
			}
		}
	}

	formatted, err := h.formatChannel(ch)
	if err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}
	return c.JSON(http.StatusCreated, formatted)
}

// GetChannelByChannelID GET /channels/:channelID
func (h *Handlers) GetChannelByChannelID(c echo.Context) error {
	ch := getChannelFromContext(c)

	formatted, err := h.formatChannel(ch)
	if err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}
	return c.JSON(http.StatusOK, formatted)
}

// PatchChannelByChannelID PATCH /channels/:channelID
func (h *Handlers) PatchChannelByChannelID(c echo.Context) error {
	channelID := getRequestParamAsUUID(c, paramChannelID)

	var req struct {
		Name       *string `json:"name"`
		Visibility *bool   `json:"visibility"`
		Force      *bool   `json:"force"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return badRequest(err)
	}

	if req.Name != nil && len(*req.Name) > 0 {
		if err := h.Repo.ChangeChannelName(channelID, *req.Name); err != nil {
			switch {
			case repository.IsArgError(err):
				return badRequest(err)
			case err == repository.ErrAlreadyExists:
				return conflict("channel name conflicts")
			case err == repository.ErrForbidden:
				return forbidden("the channel's name cannot be changed")
			default:
				return internalServerError(err, h.requestContextLogger(c))
			}
		}
	}

	if req.Force != nil || req.Visibility != nil {
		if err := h.Repo.UpdateChannelAttributes(channelID, req.Visibility, req.Force); err != nil {
			return internalServerError(err, h.requestContextLogger(c))
		}
	}
	return c.NoContent(http.StatusNoContent)
}

// PostChannelChildren POST /channels/:channelID/children
func (h *Handlers) PostChannelChildren(c echo.Context) error {
	userID := getRequestUserID(c)
	parentCh := getChannelFromContext(c)

	var req struct {
		Name string `json:"name"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return badRequest(err)
	}

	// 子チャンネル作成
	ch, err := h.Repo.CreateChildChannel(req.Name, parentCh.ID, userID)
	if err != nil {
		switch {
		case repository.IsArgError(err):
			return badRequest(err)
		case err == repository.ErrAlreadyExists:
			return conflict("channel name conflicts")
		case err == repository.ErrChannelDepthLimitation:
			return badRequest("channel depth limit exceeded")
		case err == repository.ErrForbidden:
			return forbidden("invalid parent channel")
		default:
			return internalServerError(err, h.requestContextLogger(c))
		}
	}

	formatted, err := h.formatChannel(ch)
	if err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}
	return c.JSON(http.StatusCreated, formatted)
}

// PutChannelParent PUT /channels/:channelID/parent
func (h *Handlers) PutChannelParent(c echo.Context) error {
	channelID := getRequestParamAsUUID(c, paramChannelID)

	var req struct {
		Parent uuid.UUID `json:"parent"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return badRequest(err)
	}

	if err := h.Repo.ChangeChannelParent(channelID, req.Parent); err != nil {
		switch err {
		case repository.ErrAlreadyExists:
			return conflict("channel name conflicts")
		case repository.ErrChannelDepthLimitation:
			return badRequest("channel depth limit exceeded")
		case repository.ErrForbidden:
			return forbidden("invalid parent channel")
		default:
			return internalServerError(err, h.requestContextLogger(c))
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteChannelByChannelID DELETE /channels/:channelID
func (h *Handlers) DeleteChannelByChannelID(c echo.Context) error {
	channelID := getRequestParamAsUUID(c, paramChannelID)

	if err := h.Repo.DeleteChannel(channelID); err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}

	return c.NoContent(http.StatusNoContent)
}

// GetTopic GET /channels/:channelID/topic
func (h *Handlers) GetTopic(c echo.Context) error {
	ch := getChannelFromContext(c)
	return c.JSON(http.StatusOK, map[string]string{
		"text": ch.Topic,
	})
}

// PutTopic PUT /channels/:channelID/topic
func (h *Handlers) PutTopic(c echo.Context) error {
	userID := getRequestUserID(c)
	channelID := getRequestParamAsUUID(c, paramChannelID)

	var req struct {
		Text string `json:"text"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return badRequest(err)
	}

	if err := h.Repo.UpdateChannelTopic(channelID, req.Text, userID); err != nil {
		return internalServerError(err, h.requestContextLogger(c))
	}

	return c.NoContent(http.StatusNoContent)
}
