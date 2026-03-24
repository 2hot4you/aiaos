package postgres

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"gorm.io/gorm"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(p *domain.Project) error {
	return r.db.Create(p).Error
}

func (r *ProjectRepo) FindByID(id int64) (*domain.Project, error) {
	var p domain.Project
	err := r.db.Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) Update(p *domain.Project) error {
	return r.db.Save(p).Error
}

func (r *ProjectRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&domain.Project{}).Error
}

func (r *ProjectRepo) List(page, pageSize int, keyword, sortBy, sortOrder string) ([]domain.Project, int64, error) {
	var items []domain.Project
	var total int64

	query := r.db.Model(&domain.Project{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	if sortBy == "" {
		sortBy = "updated_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	orderStr := sortBy + " " + sortOrder

	offset := (page - 1) * pageSize
	err := query.Order(orderStr).Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}
