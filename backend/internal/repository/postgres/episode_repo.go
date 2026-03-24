package postgres

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"gorm.io/gorm"
)

type EpisodeRepo struct {
	db *gorm.DB
}

func NewEpisodeRepo(db *gorm.DB) *EpisodeRepo {
	return &EpisodeRepo{db: db}
}

func (r *EpisodeRepo) Create(e *domain.Episode) error {
	return r.db.Create(e).Error
}

func (r *EpisodeRepo) FindByID(id int64) (*domain.Episode, error) {
	var e domain.Episode
	err := r.db.Where("id = ?", id).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EpisodeRepo) Update(e *domain.Episode) error {
	return r.db.Save(e).Error
}

func (r *EpisodeRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&domain.Episode{}).Error
}

func (r *EpisodeRepo) ListBySeason(seasonID int64) ([]domain.Episode, error) {
	var items []domain.Episode
	err := r.db.Where("season_id = ?", seasonID).
		Order("sort_order ASC").
		Find(&items).Error
	return items, err
}

func (r *EpisodeRepo) GetMaxSortOrder(seasonID int64) (int, error) {
	var maxOrder *int
	err := r.db.Model(&domain.Episode{}).
		Where("season_id = ?", seasonID).
		Select("MAX(sort_order)").
		Scan(&maxOrder).Error
	if err != nil || maxOrder == nil {
		return 0, err
	}
	return *maxOrder, nil
}
