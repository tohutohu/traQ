package model

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"github.com/traPtitech/traQ/utils/validator"
	"time"
)

const (
	// FileTypeUserFile ユーザーアップロードファイルタイプ
	FileTypeUserFile = ""
	// FileTypeIcon ユーザーアイコンファイルタイプ
	FileTypeIcon = "icon"
	// FileTypeStamp スタンプファイルタイプ
	FileTypeStamp = "stamp"
	// FileTypeThumbnail サムネイルファイルタイプ
	FileTypeThumbnail = "thumbnail"
)

// File DBに格納するファイルの構造体
type File struct {
	ID              uuid.UUID  `gorm:"type:char(36);not null;primary_key"   json:"fileId"`
	Name            string     `gorm:"type:text;not null"                   json:"name"                  validate:"required"`
	Mime            string     `gorm:"type:text;not null"                   json:"mime"                  validate:"required"`
	Size            int64      `gorm:"type:bigint;not null"                 json:"size"                  validate:"min=0,required"`
	CreatorID       uuid.UUID  `gorm:"type:char(36);not null"               json:"-"`
	Hash            string     `gorm:"type:char(32);not null"               json:"md5"                   validate:"max=32"`
	Type            string     `gorm:"type:varchar(30);not null;default:''" json:"-"`
	HasThumbnail    bool       `gorm:"type:boolean;not null;default:false"  json:"hasThumb"`
	ThumbnailWidth  int        `gorm:"type:int;not null;default:0"          json:"thumbWidth,omitempty"  validate:"min=0"`
	ThumbnailHeight int        `gorm:"type:int;not null;default:0"          json:"thumbHeight,omitempty" validate:"min=0"`
	CreatedAt       time.Time  `gorm:"precision:6"                          json:"datetime"`
	DeletedAt       *time.Time `gorm:"precision:6"                          json:"-"`
}

// TableName dbのtableの名前を返します
func (f *File) TableName() string {
	return "files"
}

// Validate 構造体を検証します
func (f *File) Validate() error {
	return validator.ValidateStruct(f)
}

// GetKey ファイルのストレージに対するキーを返す
func (f *File) GetKey() string {
	return f.ID.String()
}

// GetThumbKey ファイルのサムネイルのストレージに対するキーを返す
func (f *File) GetThumbKey() string {
	return f.ID.String() + "-thumb"
}

// FileACLEntry ファイルアクセスコントロールリストエントリー構造体
type FileACLEntry struct {
	FileID uuid.UUID     `gorm:"type:char(36);primary_key;not null"`
	UserID uuid.NullUUID `gorm:"type:char(36);primary_key;not null"`
	Allow  sql.NullBool  `gorm:"not null"`
}

// TableName FileACLEntry構造体のテーブル名
func (f *FileACLEntry) TableName() string {
	return "files_acl"
}
