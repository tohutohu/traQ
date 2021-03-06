package repository

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/utils"
	"time"
)

// GetClient implements OAuth2Repository interface.
func (repo *GormRepository) GetClient(id string) (*model.OAuth2Client, error) {
	if len(id) == 0 {
		return nil, ErrNotFound
	}
	oc := &model.OAuth2Client{}
	if err := repo.db.Take(oc, &model.OAuth2Client{ID: id}).Error; err != nil {
		return nil, convertError(err)
	}
	return oc, nil
}

// GetClientsByUser implements OAuth2Repository interface.
func (repo *GormRepository) GetClientsByUser(userID uuid.UUID) ([]*model.OAuth2Client, error) {
	cs := make([]*model.OAuth2Client, 0)
	if userID == uuid.Nil {
		return cs, nil
	}
	return cs, repo.db.Where(&model.OAuth2Client{CreatorID: userID}).Find(&cs).Error
}

// SaveClient implements OAuth2Repository interface.
func (repo *GormRepository) SaveClient(client *model.OAuth2Client) error {
	return repo.db.Create(client).Error
}

// UpdateClient implements OAuth2Repository interface.
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

// DeleteClient implements OAuth2Repository interface.
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

// SaveAuthorize implements OAuth2Repository interface.
func (repo *GormRepository) SaveAuthorize(data *model.OAuth2Authorize) error {
	return repo.db.Create(data).Error
}

// GetAuthorize implements OAuth2Repository interface.
func (repo *GormRepository) GetAuthorize(code string) (*model.OAuth2Authorize, error) {
	if len(code) == 0 {
		return nil, ErrNotFound
	}
	oa := &model.OAuth2Authorize{}
	if err := repo.db.Take(oa, &model.OAuth2Authorize{Code: code}).Error; err != nil {
		return nil, convertError(err)
	}
	return oa, nil
}

// DeleteAuthorize implements OAuth2Repository interface.
func (repo *GormRepository) DeleteAuthorize(code string) error {
	if len(code) == 0 {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Authorize{Code: code}).Error
}

// IssueToken implements OAuth2Repository interface.
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

// GetTokenByID implements OAuth2Repository interface.
func (repo *GormRepository) GetTokenByID(id uuid.UUID) (*model.OAuth2Token, error) {
	if id == uuid.Nil {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Take(ot, &model.OAuth2Token{ID: id}).Error; err != nil {
		return nil, convertError(err)
	}
	return ot, nil
}

// DeleteTokenByID implements OAuth2Repository interface.
func (repo *GormRepository) DeleteTokenByID(id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Token{}, &model.OAuth2Token{ID: id}).Error
}

// GetTokenByAccess implements OAuth2Repository interface.
func (repo *GormRepository) GetTokenByAccess(access string) (*model.OAuth2Token, error) {
	if len(access) == 0 {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Take(ot, &model.OAuth2Token{AccessToken: access}).Error; err != nil {
		return nil, convertError(err)
	}
	return ot, nil
}

// DeleteTokenByAccess implements OAuth2Repository interface.
func (repo *GormRepository) DeleteTokenByAccess(access string) error {
	if len(access) == 0 {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Token{}, &model.OAuth2Token{AccessToken: access}).Error
}

// GetTokenByRefresh implements OAuth2Repository interface.
func (repo *GormRepository) GetTokenByRefresh(refresh string) (*model.OAuth2Token, error) {
	if len(refresh) == 0 {
		return nil, ErrNotFound
	}
	ot := &model.OAuth2Token{}
	if err := repo.db.Take(ot, &model.OAuth2Token{RefreshToken: refresh, RefreshEnabled: true}).Error; err != nil {
		return nil, convertError(err)
	}
	return ot, nil
}

// DeleteTokenByRefresh implements OAuth2Repository interface.
func (repo *GormRepository) DeleteTokenByRefresh(refresh string) error {
	if len(refresh) == 0 {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Token{}, &model.OAuth2Token{RefreshToken: refresh, RefreshEnabled: true}).Error
}

// GetTokensByUser implements OAuth2Repository interface.
func (repo *GormRepository) GetTokensByUser(userID uuid.UUID) ([]*model.OAuth2Token, error) {
	ts := make([]*model.OAuth2Token, 0)
	if userID == uuid.Nil {
		return ts, nil
	}
	return ts, repo.db.Where(&model.OAuth2Token{UserID: userID}).Find(&ts).Error
}

// DeleteTokenByUser implements OAuth2Repository interface.
func (repo *GormRepository) DeleteTokenByUser(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Token{}, &model.OAuth2Token{UserID: userID}).Error
}

// DeleteTokenByClient implements OAuth2Repository interface.
func (repo *GormRepository) DeleteTokenByClient(clientID string) error {
	if len(clientID) == 0 {
		return nil
	}
	return repo.db.Delete(&model.OAuth2Token{}, &model.OAuth2Token{ClientID: clientID}).Error
}
