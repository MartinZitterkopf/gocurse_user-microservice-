package user

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MartinZitterkopf/gocurse_domain/domain"
	"gorm.io/gorm"
)

type (
	Repository interface {
		Create(ctx context.Context, user *domain.User) error
		GetAll(ctx context.Context, filters Fillters, limit, offset int) ([]domain.User, error)
		GetByID(ctx context.Context, id string) (*domain.User, error)
		Delete(ctx context.Context, id string) error
		Update(ctx context.Context, id string, firstName *string, lastName *string, email *string, phone *string) error
		Count(ctx context.Context, filters Fillters) (int, error)
	}

	repo struct {
		log *log.Logger
		db  *gorm.DB
	}
)

func NewRepo(log *log.Logger, db *gorm.DB) Repository {
	return &repo{
		log: log,
		db:  db,
	}
}

func (repo *repo) Create(ctx context.Context, user *domain.User) error {

	if err := repo.db.WithContext(ctx).Create(user).Error; err != nil {
		repo.log.Printf("error; %v", err)
		return err
	}
	repo.log.Println("user created with id: ", user.ID)
	return nil
}

func (repo *repo) GetAll(ctx context.Context, filters Fillters, offset, limit int) ([]domain.User, error) {

	var u []domain.User

	tx := repo.db.WithContext(ctx).Model(&u)
	tx = applyFilters(tx, filters)
	tx = tx.Limit(limit).Offset(offset)

	result := tx.Order("created_at desc").Find(&u)
	if result.Error != nil {
		repo.log.Println(result.Error)
		return nil, result.Error
	}

	return u, nil
}

func (repo *repo) GetByID(ctx context.Context, id string) (*domain.User, error) {

	user := domain.User{ID: id}

	if err := repo.db.WithContext(ctx).First(&user).Error; err != nil {
		repo.log.Println(err)
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound{id}
		}

		return nil, err
	}

	return &user, nil
}

func (repo *repo) Delete(ctx context.Context, id string) error {

	user := domain.User{ID: id}

	result := repo.db.WithContext(ctx).Delete(&user)
	if result.Error != nil {
		repo.log.Println(result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		repo.log.Printf("user %s doest't exists", id)
		return ErrNotFound{id}
	}

	return nil
}

func (repo *repo) Update(ctx context.Context, id string, firstName *string, lastName *string, email *string, phone *string) error {

	values := make(map[string]interface{})

	if firstName != nil {
		values["first_name"] = *firstName
	}

	if lastName != nil {
		values["last_name"] = *lastName
	}

	if email != nil {
		values["email"] = *email
	}

	if phone != nil {
		values["phone"] = *phone
	}

	result := repo.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		repo.log.Println(result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		repo.log.Printf("user %s doest't exists", id)
		return ErrNotFound{id}
	}

	return nil
}

func applyFilters(tx *gorm.DB, filters Fillters) *gorm.DB {

	if filters.FirstName != "" {
		filters.FirstName = fmt.Sprintf("%%%s%%", strings.ToLower(filters.FirstName))
		tx = tx.Where("lower(first_name) like ?", filters.FirstName)
	}

	if filters.LastName != "" {
		filters.LastName = fmt.Sprintf("%%%s%%", strings.ToLower(filters.LastName))
		tx = tx.Where("lower(first_name) like ?", filters.LastName)
	}

	return tx
}

func (repo *repo) Count(ctx context.Context, filters Fillters) (int, error) {

	var count int64
	tx := repo.db.WithContext(ctx).Model(domain.User{})

	tx = applyFilters(tx, filters)
	if err := tx.Count(&count).Error; err != nil {
		repo.log.Println(err)
		return 0, err
	}

	return int(count), nil
}
