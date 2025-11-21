package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// GormRepository persists bookings.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

// AutoMigrate ensures bookings table exists.
func AutoMigrate(db *gorm.DB) error {
	if db.Migrator().HasTable(&bookingModel{}) {
		return nil
	}
	return db.AutoMigrate(&bookingModel{})
}

func (r *GormRepository) Create(ctx context.Context, b domain.Booking) error {
	return r.db.WithContext(ctx).Create(&bookingModel{
		ID:          b.ID,
		UserID:      b.UserID,
		RoomTypeID:  b.RoomTypeID,
		CheckIn:     b.CheckIn,
		CheckOut:    b.CheckOut,
		Status:      b.Status,
		Guests:      b.Guests,
		TotalPrice:  b.TotalPrice,
		TotalNights: b.TotalNights,
		CreatedAt:   b.CreatedAt,
	}).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Booking, error) {
	var model bookingModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Booking{}, translateErr(err)
	}
	return model.toDomain(), nil
}

func (r *GormRepository) List(ctx context.Context, opts query.Options) ([]domain.Booking, error) {
	var models []bookingModel
	qo := opts.Normalize(50)
	tx := r.db.WithContext(ctx).Order("created_at DESC")
	if qo.Limit > 0 {
		tx = tx.Limit(qo.Limit).Offset(qo.Offset)
	}
	if err := tx.Find(&models).Error; err != nil {
		return nil, err
	}
	bookings := make([]domain.Booking, 0, len(models))
	for _, m := range models {
		bookings = append(bookings, m.toDomain())
	}
	return bookings, nil
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Model(&bookingModel{}).Where("id = ?", id).Update("status", status)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return pkgErrors.New("not_found", "booking not found")
	}
	return nil
}

type bookingModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;index"`
	RoomTypeID  uuid.UUID `gorm:"type:uuid;index"`
	CheckIn     time.Time
	CheckOut    time.Time
	Status      string `gorm:"index"`
	Guests      int
	TotalPrice  float64 `gorm:"type:numeric"`
	TotalNights int
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (bookingModel) TableName() string { return "bookings" }

func (m bookingModel) toDomain() domain.Booking {
	return domain.Booking{
		ID:          m.ID,
		UserID:      m.UserID,
		RoomTypeID:  m.RoomTypeID,
		CheckIn:     m.CheckIn,
		CheckOut:    m.CheckOut,
		Status:      m.Status,
		Guests:      m.Guests,
		TotalPrice:  m.TotalPrice,
		TotalNights: m.TotalNights,
		CreatedAt:   m.CreatedAt,
	}
}

func translateErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pkgErrors.New("not_found", "booking not found")
	}
	return err
}
