package model

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
	"github.com/traPtitech/traQ/utils"
	"github.com/traPtitech/traQ/utils/validator"
	"net/http"
	"net/url"
	"time"
)

var (
	// ErrUserBotTryLogin : ユーザーエラー botユーザーでログインを試みました。botユーザーはログインできません。
	ErrUserBotTryLogin = errors.New("bot user is not allowed to login")
	// ErrUserWrongIDOrPassword : ユーザーエラー IDかパスワードが間違っています。
	ErrUserWrongIDOrPassword = errors.New("password or id is wrong")
)

// UserAccountStatus ユーザーアカウント状態
type UserAccountStatus int

// Valid 有効な値かどうか
func (v UserAccountStatus) Valid() bool {
	return userAccountStatuses[v]
}

const (
	// UserAccountStatusDeactivated ユーザーアカウント状態: 凍結
	UserAccountStatusDeactivated UserAccountStatus = 0
	// UserAccountStatusActive ユーザーアカウント状態: 有効
	UserAccountStatusActive UserAccountStatus = 1
	// UserAccountStatusSuspended ユーザーアカウント状態: 一時停止
	UserAccountStatusSuspended UserAccountStatus = 2
)

var userAccountStatuses = map[UserAccountStatus]bool{
	UserAccountStatusDeactivated: true,
	UserAccountStatusActive:      true,
	UserAccountStatusSuspended:   true,
}

// User userの構造体
type User struct {
	ID          uuid.UUID         `gorm:"type:char(36);not null;primary_key"`
	Name        string            `gorm:"type:varchar(32);not null;unique"     validate:"required,name"`
	DisplayName string            `gorm:"type:varchar(64);not null;default:''" validate:"max=64"`
	Password    string            `gorm:"type:char(128);not null;default:''"   validate:"required,max=128"`
	Salt        string            `gorm:"type:char(128);not null;default:''"   validate:"required,max=128"`
	Icon        uuid.UUID         `gorm:"type:char(36);not null"`
	Status      UserAccountStatus `gorm:"type:tinyint;not null;default:0"`
	Bot         bool              `gorm:"type:boolean;not null;default:false"`
	Role        string            `gorm:"type:varchar(30);not null;default:'user'"    validate:"required"`
	TwitterID   string            `gorm:"type:varchar(15);not null;default:''" validate:"twitterid"`
	LastOnline  *time.Time        `gorm:"precision:6"`
	CreatedAt   time.Time         `gorm:"precision:6"`
	UpdatedAt   time.Time         `gorm:"precision:6"`
}

// TableName dbの名前を指定する
func (user *User) TableName() string {
	return "users"
}

// Validate 構造体を検証します
func (user *User) Validate() error {
	return validator.ValidateStruct(user)
}

// AuthenticateUser ユーザー構造体とパスワードを照合します
func AuthenticateUser(user *User, password string) error {
	if user == nil {
		return ErrUserWrongIDOrPassword
	}
	// Botはログイン不可
	if user.Bot {
		return ErrUserBotTryLogin
	}

	if viper.GetBool("externalAuthentication.enabled") {
		values := url.Values{}
		values.Set(viper.GetString("externalAuthentication.authPost.formUserNameKey"), user.Name)
		values.Set(viper.GetString("externalAuthentication.authPost.formPasswordKey"), password)
		resp, err := http.PostForm(viper.GetString("externalAuthentication.authPost.url"), values)
		if err != nil {
			return ErrUserWrongIDOrPassword
		}
		defer resp.Body.Close()
		if resp.StatusCode != viper.GetInt("externalAuthentication.authPost.successfulCode") {
			return ErrUserWrongIDOrPassword
		}
	} else {
		if len(user.Password) == 0 || len(user.Salt) == 0 {
			return ErrUserWrongIDOrPassword
		}

		storedPassword, err := hex.DecodeString(user.Password)
		if err != nil {
			return ErrUserWrongIDOrPassword
		}
		salt, err := hex.DecodeString(user.Salt)
		if err != nil {
			return ErrUserWrongIDOrPassword
		}

		if subtle.ConstantTimeCompare(storedPassword, utils.HashPassword(password, salt)) != 1 {
			return ErrUserWrongIDOrPassword
		}
	}
	return nil
}

// UserStatus userの状態
type UserStatus struct {
	UserID   uuid.UUID `json:"userId"`
	Status   string    `json:"status"`
	LastTime time.Time `json:"-"`
}

// HeartbeatStatus Heartbeatの状態
type HeartbeatStatus struct {
	UserStatuses []*UserStatus `json:"userStatuses"`
	ChannelID    uuid.UUID     `json:"channelId"`
}
