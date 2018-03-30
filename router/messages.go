package router

import (
	"net/http"
	"time"

	"github.com/traPtitech/traQ/notification"
	"github.com/traPtitech/traQ/notification/events"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/traPtitech/traQ/model"
)

//MessageForResponse :クライアントに返す形のメッセージオブジェクト
type MessageForResponse struct {
	MessageID       string                `json:"messageId"`
	UserID          string                `json:"userId"`
	ParentChannelID string                `json:"parentChannelId"`
	Content         string                `json:"content"`
	Datetime        time.Time             `json:"datetime"`
	Pin             bool                  `json:"pin"`
	StampList       []*model.MessageStamp `json:"stampList"`
}

// GetMessageByID GET /messages/{messageID} のハンドラ
func GetMessageByID(c echo.Context) error {
	messageID := c.Param("messageID")
	m, err := validateMessageID(messageID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, formatMessage(m))
}

// GetMessagesByChannelID GET /channels/{channelID}/messages のハンドラ
func GetMessagesByChannelID(c echo.Context) error {
	queryParam := &struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}{}
	if err := c.Bind(queryParam); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid format")
	}

	// channelIDの検証
	userID := c.Get("user").(*model.User).ID
	res, err := getMessages(c.Param("channelID"), userID, queryParam.Limit, queryParam.Offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// GetMessagesFromPrivateChannel GET /users/{userID}/messages のハンドラ
func GetMessagesFromPrivateChannel(c echo.Context) error {
	myID := c.Get("user").(*model.User).ID
	userID := c.Param("userID")
	if _, err := validateUserID(userID); err != nil {
		return err
	}

	q := &struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}{}
	if err := c.Bind(q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid format")
	}

	upc, err := model.GetPrivateChannel(myID, userID)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			return c.JSON(http.StatusNotFound, "There is no message")
		default:
			c.Logger().Errorf("failed to get private channel: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "An error occurred while getting private channel")
		}
	}

	// messagesを取得・整形
	res, err := getMessages(upc.ChannelID, myID, q.Limit, q.Offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// PostMessage POST /channels/{channelID}/messages のハンドラ
func PostMessage(c echo.Context) error {
	post := &struct {
		Text string `json:"text"`
	}{}
	if err := c.Bind(post); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid format")
	}

	m, err := createMessage(post.Text, c.Get("user").(*model.User).ID, c.Param("channelID"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, m)
}

// PostPrivateMessage POST /users/{userID}/messages のハンドラ
func PostPrivateMessage(c echo.Context) error {
	userID := c.Param("userID")
	if _, err := validateUserID(userID); err != nil {
		return err
	}

	// privateチャンネルの存在確認と作成
	var channelID string
	myID := c.Get("user").(*model.User).ID
	upc, err := model.GetPrivateChannel(myID, userID)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			ch, err := createChannel("private", myID, "private", "", []string{myID, userID})
			if err != nil {
				return err
			}
			channelID = ch.ID
		default:
			c.Logger().Errorf("failed to get private channel: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "An error occurred while getting private channel")
		}
	} else {
		channelID = upc.ChannelID
	}

	post := &struct {
		Text string `json:"text"`
	}{}
	if err := c.Bind(post); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid format")
	}

	m, err := createMessage(post.Text, myID, channelID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, m)
}

// PutMessageByID PUT /messages/{messageID}のハンドラ
func PutMessageByID(c echo.Context) error {
	m, err := validateMessageID(c.Param("messageID"))
	if err != nil {
		return err
	}

	req := &struct {
		Text string `json:"text"`
	}{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid format")
	}

	m.Text = req.Text
	m.UpdaterID = c.Get("user").(*model.User).ID

	if err := m.Update(); err != nil {
		c.Logger().Errorf("message.Update() returned an error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update the message")
	}

	go notification.Send(events.MessageUpdated, events.MessageEvent{Message: *m})
	return c.JSON(http.StatusOK, formatMessage(m))
}

// DeleteMessageByID : DELETE /message/{messageID} のハンドラ
func DeleteMessageByID(c echo.Context) error {
	userID := c.Get("user").(*model.User).ID
	messageID := c.Param("messageID")

	m, err := validateMessageID(messageID)
	if err != nil {
		return err
	}
	if m.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "you are not allowed to delete this message")
	}

	m.IsDeleted = true
	if err := m.Update(); err != nil {
		c.Logger().Errorf("message.Update() returned an error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update the message")
	}

	if err := model.DeleteUnreadsByMessageID(messageID); err != nil {
		c.Logger().Errorf("model.DeleteUnreadsByMessageID returned an error: %v", err) //500エラーにはしない
	}

	go notification.Send(events.MessageDeleted, events.MessageEvent{Message: *m})
	return c.NoContent(http.StatusNoContent)
}

// dbにデータを入れる
func createMessage(text, userID, channelID string) (*MessageForResponse, error) {
	m := &model.Message{
		UserID:    userID,
		Text:      text,
		ChannelID: channelID,
	}
	if err := m.Create(); err != nil {
		log.Errorf("Message.Create() returned an error: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to insert your message")
	}

	go notification.Send(events.MessageCreated, events.MessageEvent{Message: *m})
	return formatMessage(m), nil
}

// チャンネルのデータを取得する
func getMessages(channelID, userID string, limit, offset int) ([]*MessageForResponse, error) {
	if _, err := validateChannelID(channelID, userID); err != nil {
		return nil, err
	}

	messages, err := model.GetMessagesByChannelID(channelID, limit, offset)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Channel is not found")
	}

	res := make([]*MessageForResponse, 0)
	for _, message := range messages {
		res = append(res, formatMessage(message))
	}
	return res, nil
}

func formatMessage(raw *model.Message) *MessageForResponse {
	isPinned, err := raw.IsPinned()
	if err != nil {
		log.Error(err)
	}

	stampList, err := model.GetMessageStamps(raw.ID)
	if err != nil {
		log.Error(err)
	}

	res := &MessageForResponse{
		MessageID:       raw.ID,
		UserID:          raw.UserID,
		ParentChannelID: raw.ChannelID,
		Pin:             isPinned,
		Content:         raw.Text,
		Datetime:        raw.CreatedAt.Truncate(time.Second).UTC(),
		StampList:       stampList,
	}
	return res
}

// リクエストで飛んできたmessageIDを検証する。存在する場合はそのメッセージを返す
func validateMessageID(messageID string) (*model.Message, error) {
	m := &model.Message{ID: messageID}
	ok, err := m.Exists()
	if err != nil {
		log.Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot find message")
	}
	if !ok {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Message is not found")
	}
	return m, nil
}
