package repository

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/utils"
	"time"
)

// GetClient クライアントIDからクライアントを取得します
func (repo *GormRepository) GetClient(id string) (*model.OAuth2Client, error) {
	if len(id) == 0 {
		return nil, ErrNotFound
	}
	oc := &model.OAuth2Client{}
	if err := repo.db.Where(model.OAuth2Client{ID: id}).Take(oc).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return oc, nil
}

// GetClientsByUser 指定した登録者のクライアントを全て取得します
func (repo *GormRepository) GetClientsByUser(userID uuid.UUID) ([]*model.OAuth2Client, error) {
	cs := make([]*model.OAuth2Client, 0)
	if userID == uuid.Nil {
		return cs, nil
	}
	return cs, repo.db.Where(&model.OAuth2Client{CreatorID: userID}).Find(&cs).Error
}

// SaveClient クライアントを保存します
func (repo *GormRepository) SaveClient(client *model.OAuth2Client) error {
	return repo.db.Create(client).Error
}

// UpdateClient クライアント情報を更新します
func (repo *GormRepository) UpdateClient(client *model.OAuth2Client) error {
	if len(client.ID) == 0 {
		return ErrNilID
	}
	return repo.db.Model(&model.OAuth2Client{ID: client.ID}).Updates(map[string]interface{}{
		"name":         client.Name,
		"description":  client.Description,
		"confidential": client.Confidential,
		"creator_id":   client.CreatorID,
		"secret":       client.Secret,
		"redirect_uri": client.RedirectURI,
		"scopes":       client.Scopes,
	}).Error
}

// DeleteClient クライアントを削除します
func (repo *GormRepository) DeleteClient(id string) error {
	if len(id) == 0 {
		return nil
	}
	err := repo.transact(func(tx *gorm.DB) error {
		errs := tx.Delete(&model.OAuth2Client{ID: id}).
			Delete(&model.OAuth2Authorize{}, &model.OAuth2Authorize{ClientID: id}).
			Delete(&model.OAuth2Token{}, &model.OAuth2Token{ClientID: id}).
			GetErrors()
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	})
	return err
}

// SaveAuthorize 認可データを保存します
func (repo *GormRepository) SaveAuthorize(data *model.OAuth2Authorize) error {
	return repo.db.Create(data).Error
}

// GetAuthorize 認可コードから認可データを取得します
func (repo *GormRepository) GetAuthorize(code string) (*model.OAuth2Authorize, error) {
	if len(code) == 0 {
		return nil, ErrNotFound
	}
	oa := &model.OAuth2Authorize{}
	if err := repo.db.Where(&model.OAuth2Authorize{Code: code}).Take(oa).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return oa, nil
}

// DeleteAuthorize 認可コードから認可データを削除します
func (repo *GormRepository) DeleteAuthorize(code string) error {
	if len(code) == 0 {
		return nil
	}
	return repo.db.Where(&model.OAuth2Authorize{Code: code}).Delete(&model.OAuth2Authorize{}).Error
}

// SaveToken トークンを発行します
func (repo *GormRepository) IssueToken(client *model.OAuth2Client, userID uuid.UUID, redirectURI string, scope model.AccessScopes, expire int, refresh bool) (*model.OAuth2Token, error) {
	newToken := &model.OAuth2Token{
		ID:             uuid.Must(uuid.NewV4()),
		UserID:         userID,
		RedirectURI:    redirectURI,
		AccessToken:    utils.RandAlphabetAndNumberString(36),
		RefreshToken:   utils.RandAlphabetAndNumberString(36),
		RefreshEnabled: refresh,
		CreatedAt:      time.Now(),
		ExpiresIn:      expire,
		Scopes:         scope,
	}

	if client != nil {
		newToken.ClientID = client.ID
	}

	return newToken, repo.db.Create(newToken).Error
}

// GetTokenByID トークンIDからトークンを取得します
func (repo *GormRepository) GetTokenByID(id uuid.UUID) (*model.OAuth2Token, error) {
	if id == uuid.Nil {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Where(&model.OAuth2Token{ID: id}).Take(ot).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ot, nil
}

// DeleteTokenByID トークンIDからトークンを削除します
func (repo *GormRepository) DeleteTokenByID(id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	return repo.db.Where(&model.OAuth2Token{ID: id}).Delete(&model.OAuth2Token{}).Error
}

// GetTokenByAccess アクセストークンからトークンを取得します
func (repo *GormRepository) GetTokenByAccess(access string) (*model.OAuth2Token, error) {
	if len(access) == 0 {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Where(&model.OAuth2Token{AccessToken: access}).Take(ot).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ot, nil
}

// DeleteTokenByAccess アクセストークンからトークンを削除します
func (repo *GormRepository) DeleteTokenByAccess(access string) error {
	if len(access) == 0 {
		return nil
	}
	return repo.db.Where(&model.OAuth2Token{AccessToken: access}).Delete(&model.OAuth2Token{}).Error
}

// GetTokenByRefresh リフレッシュトークンからトークンを取得します
func (repo *GormRepository) GetTokenByRefresh(refresh string) (*model.OAuth2Token, error) {
	if len(refresh) == 0 {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Where(&model.OAuth2Token{RefreshToken: refresh, RefreshEnabled: true}).Take(ot).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ot, nil
}

// DeleteTokenByRefresh リフレッシュトークンからトークンを削除します
func (repo *GormRepository) DeleteTokenByRefresh(refresh string) error {
	if len(refresh) == 0 {
		return nil
	}
	return repo.db.Where(&model.OAuth2Token{RefreshToken: refresh, RefreshEnabled: true}).Delete(&model.OAuth2Token{}).Error
}

// GetTokensByUser 指定したユーザーのトークンを全て取得します
func (repo *GormRepository) GetTokensByUser(userID uuid.UUID) ([]*model.OAuth2Token, error) {
	ts := make([]*model.OAuth2Token, 0)
	if userID == uuid.Nil {
		return ts, nil
	}
	return ts, repo.db.Where(&model.OAuth2Token{UserID: userID}).Find(&ts).Error
}

// DeleteTokenByUser 指定したユーザーのトークンを全て削除します
func (repo *GormRepository) DeleteTokenByUser(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return nil
	}
	return repo.db.Where(&model.OAuth2Token{UserID: userID}).Delete(&model.OAuth2Token{}).Error
}

// DeleteTokenByClient 指定したクライアントのトークンを全て削除します
func (repo *GormRepository) DeleteTokenByClient(clientID string) error {
	if len(clientID) == 0 {
		return nil
	}
	return repo.db.Where(&model.OAuth2Token{ClientID: clientID}).Delete(&model.OAuth2Token{}).Error
}