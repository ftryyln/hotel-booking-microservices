package hotel

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// Service exposes hotel catalog operations.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateHotel(ctx context.Context, req dto.HotelRequest) (uuid.UUID, error) {
	h := domain.Hotel{ID: uuid.New(), Name: req.Name, Description: req.Description, Address: req.Address}
	return h.ID, s.repo.CreateHotel(ctx, h)
}

func (s *Service) ListHotels(ctx context.Context) ([]dto.HotelResponse, error) {
	hotels, err := s.repo.ListHotels(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.HotelResponse
	for _, h := range hotels {
		roomTypes, _ := s.repo.ListRoomTypes(ctx, h.ID)
		var summaries []dto.RoomTypeSummary
		for _, rt := range roomTypes {
			summaries = append(summaries, dto.RoomTypeSummary{
				ID:       rt.ID.String(),
				Name:     rt.Name,
				Capacity: rt.Capacity,
				Price:    rt.BasePrice,
			})
		}
		resp = append(resp, dto.HotelResponse{
			ID:          h.ID.String(),
			Name:        h.Name,
			Description: h.Description,
			Address:     h.Address,
			RoomTypes:   summaries,
		})
	}
	return resp, nil
}

func (s *Service) CreateRoomType(ctx context.Context, req dto.RoomTypeRequest) (uuid.UUID, error) {
	rt := domain.RoomType{
		ID:        uuid.New(),
		HotelID:   uuid.MustParse(req.HotelID),
		Name:      req.Name,
		Capacity:  req.Capacity,
		BasePrice: req.BasePrice,
		Amenities: req.Amenities,
	}
	return rt.ID, s.repo.CreateRoomType(ctx, rt)
}

func (s *Service) CreateRoom(ctx context.Context, req dto.RoomRequest) (uuid.UUID, error) {
	room := domain.Room{ID: uuid.New(), RoomTypeID: uuid.MustParse(req.RoomTypeID), Number: req.Number, Status: req.Status}
	return room.ID, s.repo.CreateRoom(ctx, room)
}
