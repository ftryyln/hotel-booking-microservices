package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// GormRepository implements UserRepository via GORM.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// AutoMigrate ensures users table exists.
func AutoMigrate(db *gorm.DB) error {
	if db.Migrator().HasTable(&userModel{}) {
		return nil
	}
	return db.AutoMigrate(&userModel{})
}

func (r *GormRepository) Create(ctx context.Context, user domain.User) error {
	model := toModel(user)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
		return domain.User{}, translateErr(err)
	}
	return toDomain(model), nil
}

func (r *GormRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.User{}, translateErr(err)
	}
	return toDomain(model), nil
}

func (r *GormRepository) List(ctx context.Context, opts query.Options) ([]domain.User, error) {
	var models []userModel
	qo := opts.Normalize(50)
	tx := r.db.WithContext(ctx).Order("created_at DESC")
	if qo.Limit > 0 {
		tx = tx.Limit(qo.Limit).Offset(qo.Offset)
	}
	if err := tx.Find(&models).Error; err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(models))
	for _, m := range models {
		users = append(users, toDomain(m))
	}
	return users, nil
}

type userModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string
	Role      string
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (userModel) TableName() string { return "users" }

func toModel(u domain.User) userModel {
	return userModel{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func toDomain(m userModel) domain.User {
	return domain.User{
		ID:        m.ID,
		Email:     m.Email,
		Password:  m.Password,
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
	}
}

func translateErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pkgErrors.New("not_found", "user not found")
	}
	return err
}
