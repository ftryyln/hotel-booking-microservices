package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// GormRepository implements hotel repo.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

// AutoMigrate ensures hotel related tables exist.
func AutoMigrate(db *gorm.DB) error {
	if db.Migrator().HasTable(&hotelModel{}) {
		return nil
	}
	return db.AutoMigrate(&hotelModel{}, &roomTypeModel{}, &roomModel{})
}

func (r *GormRepository) CreateHotel(ctx context.Context, h domain.Hotel) error {
	return r.db.WithContext(ctx).Create(&hotelModel{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
		Address:     h.Address,
		CreatedAt:   h.CreatedAt,
	}).Error
}

func (r *GormRepository) ListHotels(ctx context.Context, opts query.Options) ([]domain.Hotel, error) {
	var models []hotelModel
	qo := opts.Normalize(50)
	tx := r.db.WithContext(ctx)
	if qo.Limit > 0 {
		tx = tx.Limit(qo.Limit).Offset(qo.Offset)
	}
	if err := tx.Find(&models).Error; err != nil {
		return nil, err
	}
	hotels := make([]domain.Hotel, 0, len(models))
	for _, m := range models {
		hotels = append(hotels, m.toDomain())
	}
	return hotels, nil
}

func (r *GormRepository) CreateRoomType(ctx context.Context, rt domain.RoomType) error {
	return r.db.WithContext(ctx).Create(&roomTypeModel{
		ID:        rt.ID,
		HotelID:   rt.HotelID,
		Name:      rt.Name,
		Capacity:  rt.Capacity,
		BasePrice: rt.BasePrice,
		Amenities: rt.Amenities,
	}).Error
}

func (r *GormRepository) ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]domain.RoomType, error) {
	var models []roomTypeModel
	if err := r.db.WithContext(ctx).Where("hotel_id = ?", hotelID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toRoomTypes(models), nil
}

func (r *GormRepository) ListAllRoomTypes(ctx context.Context, opts query.Options) ([]domain.RoomType, error) {
	var models []roomTypeModel
	qo := opts.Normalize(50)
	tx := r.db.WithContext(ctx)
	if qo.Limit > 0 {
		tx = tx.Limit(qo.Limit).Offset(qo.Offset)
	}
	if err := tx.Find(&models).Error; err != nil {
		return nil, err
	}
	return toRoomTypes(models), nil
}

func (r *GormRepository) CreateRoom(ctx context.Context, room domain.Room) error {
	return r.db.WithContext(ctx).Create(&roomModel{
		ID:         room.ID,
		RoomTypeID: room.RoomTypeID,
		Number:     room.Number,
		Status:     room.Status,
	}).Error
}

func (r *GormRepository) GetRoomType(ctx context.Context, id uuid.UUID) (domain.RoomType, error) {
	var model roomTypeModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.RoomType{}, translateErr(err)
	}
	return model.toDomain(), nil
}

func (r *GormRepository) ListRooms(ctx context.Context, opts query.Options) ([]domain.Room, error) {
	var models []roomModel
	qo := opts.Normalize(50)
	tx := r.db.WithContext(ctx)
	if qo.Limit > 0 {
		tx = tx.Limit(qo.Limit).Offset(qo.Offset)
	}
	if err := tx.Find(&models).Error; err != nil {
		return nil, err
	}
	rooms := make([]domain.Room, 0, len(models))
	for _, m := range models {
		rooms = append(rooms, m.toDomain())
	}
	return rooms, nil
}

func (r *GormRepository) GetHotel(ctx context.Context, id uuid.UUID) (domain.Hotel, error) {
	var model hotelModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Hotel{}, translateErr(err)
	}
	return model.toDomain(), nil
}

type hotelModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string
	Description string
	Address     string
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (hotelModel) TableName() string { return "hotels" }

func (m hotelModel) toDomain() domain.Hotel {
	return domain.Hotel{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Address:     m.Address,
		CreatedAt:   m.CreatedAt,
	}
}

type roomTypeModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	HotelID   uuid.UUID `gorm:"type:uuid;index"`
	Name      string
	Capacity  int
	BasePrice float64 `gorm:"type:numeric"`
	Amenities string
}

func (roomTypeModel) TableName() string { return "room_types" }

func (m roomTypeModel) toDomain() domain.RoomType {
	return domain.RoomType{
		ID:        m.ID,
		HotelID:   m.HotelID,
		Name:      m.Name,
		Capacity:  m.Capacity,
		BasePrice: m.BasePrice,
		Amenities: m.Amenities,
	}
}

type roomModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoomTypeID uuid.UUID `gorm:"type:uuid;index"`
	Number     string
	Status     string
}

func (roomModel) TableName() string { return "rooms" }

func (m roomModel) toDomain() domain.Room {
	return domain.Room{
		ID:         m.ID,
		RoomTypeID: m.RoomTypeID,
		Number:     m.Number,
		Status:     m.Status,
	}
}

func toRoomTypes(models []roomTypeModel) []domain.RoomType {
	rts := make([]domain.RoomType, 0, len(models))
	for _, m := range models {
		rts = append(rts, m.toDomain())
	}
	return rts
}

func translateErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pkgErrors.New("not_found", "record not found")
	}
	return err
}
