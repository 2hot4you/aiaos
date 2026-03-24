package postgres

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"gorm.io/gorm"
)

type ModelConfigRepo struct {
	db *gorm.DB
}

func NewModelConfigRepo(db *gorm.DB) *ModelConfigRepo {
	return &ModelConfigRepo{db: db}
}

func (r *ModelConfigRepo) Create(m *domain.AIModelConfig) error {
	return r.db.Create(m).Error
}

func (r *ModelConfigRepo) FindByID(id int64) (*domain.AIModelConfig, error) {
	var m domain.AIModelConfig
	err := r.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ModelConfigRepo) Update(m *domain.AIModelConfig) error {
	return r.db.Save(m).Error
}

func (r *ModelConfigRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&domain.AIModelConfig{}).Error
}

func (r *ModelConfigRepo) List() ([]domain.AIModelConfig, error) {
	var items []domain.AIModelConfig
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ModelConfigRepo) ClearDefault(modelType string) error {
	return r.db.Model(&domain.AIModelConfig{}).
		Where("model_type = ? AND is_default = ?", modelType, true).
		Update("is_default", false).Error
}
