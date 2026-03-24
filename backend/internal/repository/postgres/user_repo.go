package postgres

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepo) List(page, pageSize int, keyword string) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.Where("deleted_at IS NULL")
	if keyword != "" {
		query = query.Where("username LIKE ? OR display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Model(&domain.User{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepo) SoftDelete(id int64) error {
	return r.db.Exec("UPDATE users SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id).Error
}

func (r *UserRepo) UpdateLastLogin(id int64) error {
	return r.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = ?", id).Error
}
