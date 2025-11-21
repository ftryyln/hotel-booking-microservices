package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// GormRepository persists payments using GORM.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

// AutoMigrate ensures schema is present.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&paymentModel{})
}

func (r *GormRepository) Create(ctx context.Context, p domain.Payment) error {
	model := toModel(p)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	var model paymentModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Payment{}, translateErr(err)
	}
	return toDomain(model), nil
}

func (r *GormRepository) FindByBookingID(ctx context.Context, bookingID uuid.UUID) (domain.Payment, error) {
	var model paymentModel
	if err := r.db.WithContext(ctx).First(&model, "booking_id = ?", bookingID).Error; err != nil {
		return domain.Payment{}, translateErr(err)
	}
	return toDomain(model), nil
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentURL string) error {
	updates := map[string]any{"status": status}
	if paymentURL != "" {
		updates["payment_url"] = paymentURL
	}
	res := r.db.WithContext(ctx).Model(&paymentModel{}).Where("id = ?", id).Updates(updates)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return pkgErrors.New("not_found", "payment not found")
	}
	return nil
}

type paymentModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	BookingID  uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	Amount     float64   `gorm:"type:numeric"`
	Currency   string
	Status     string `gorm:"index"`
	Provider   string
	PaymentURL string
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (paymentModel) TableName() string { return "payments" }

func toModel(p domain.Payment) paymentModel {
	return paymentModel{
		ID:         p.ID,
		BookingID:  p.BookingID,
		Amount:     p.Amount,
		Currency:   p.Currency,
		Status:     p.Status,
		Provider:   p.Provider,
		PaymentURL: p.PaymentURL,
		CreatedAt:  p.CreatedAt,
	}
}

func toDomain(m paymentModel) domain.Payment {
	return domain.Payment{
		ID:         m.ID,
		BookingID:  m.BookingID,
		Amount:     m.Amount,
		Currency:   m.Currency,
		Status:     m.Status,
		Provider:   m.Provider,
		PaymentURL: m.PaymentURL,
		CreatedAt:  m.CreatedAt,
	}
}

func translateErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pkgErrors.New("not_found", "payment not found")
	}
	return err
}
