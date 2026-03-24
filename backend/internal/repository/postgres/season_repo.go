package postgres

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"gorm.io/gorm"
)

type SeasonRepo struct {
	db *gorm.DB
}

func NewSeasonRepo(db *gorm.DB) *SeasonRepo {
	return &SeasonRepo{db: db}
}

func (r *SeasonRepo) Create(s *domain.Season) error {
	return r.db.Create(s).Error
}

func (r *SeasonRepo) FindByID(id int64) (*domain.Season, error) {
	var s domain.Season
	err := r.db.Where("id = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SeasonRepo) Update(s *domain.Season) error {
	return r.db.Save(s).Error
}

func (r *SeasonRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&domain.Season{}).Error
}

func (r *SeasonRepo) ListByProject(projectID int64) ([]domain.Season, error) {
	var items []domain.Season
	err := r.db.Where("project_id = ?", projectID).
		Order("sort_order ASC").
		Preload("Episodes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Find(&items).Error
	return items, err
}

func (r *SeasonRepo) GetMaxSortOrder(projectID int64) (int, error) {
	var maxOrder *int
	err := r.db.Model(&domain.Season{}).
		Where("project_id = ?", projectID).
		Select("MAX(sort_order)").
		Scan(&maxOrder).Error
	if err != nil || maxOrder == nil {
		return 0, err
	}
	return *maxOrder, nil
}
