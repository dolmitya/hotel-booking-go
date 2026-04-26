package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	bookingdomain "hotelService/internal/domain/booking"
	bookingdto "hotelService/internal/dto/booking"
	guestdto "hotelService/internal/dto/guest"
	roomdto "hotelService/internal/dto/room"
	"hotelService/internal/kafka"
	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/metrics"
	"hotelService/internal/model"
	"hotelService/internal/repository"
)

type BookingService struct {
	bookingRepo *repository.BookingRepository
	guestRepo   *repository.GuestRepository
	roomRepo    *repository.RoomRepository
	publisher   kafka.BookingEventPublisher
	metrics     *metrics.BookingMetrics
}

func NewBookingService(
	bookingRepo *repository.BookingRepository,
	guestRepo *repository.GuestRepository,
	roomRepo *repository.RoomRepository,
	publisher kafka.BookingEventPublisher,
	bookingMetrics *metrics.BookingMetrics,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		guestRepo:   guestRepo,
		roomRepo:    roomRepo,
		publisher:   publisher,
		metrics:     bookingMetrics,
	}
}

func (s *BookingService) GetByID(ctx context.Context, rawID string) (bookingdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return bookingdto.Response{}, bookingdomain.ErrInvalidBookingID
	}

	bookingEntity, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if bookingEntity == nil {
		return bookingdto.Response{}, bookingdomain.ErrNotFound
	}

	roomEntity, err := s.roomRepo.FindByID(ctx, bookingEntity.RoomID)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if roomEntity == nil {
		return bookingdto.Response{}, bookingdomain.ErrRoomNotFound
	}

	guests, err := s.findGuestsByBookingID(ctx, id)
	if err != nil {
		return bookingdto.Response{}, err
	}

	return toBookingResponse(*bookingEntity, *roomEntity, guests), nil
}

func (s *BookingService) FindAvailableRooms(ctx context.Context, rawStart, rawEnd string) ([]roomdto.Response, error) {
	start, end, err := parseAndValidateRange(rawStart, rawEnd)
	if err != nil {
		return nil, err
	}

	rooms, err := s.bookingRepo.FindAvailableRooms(ctx, start, end)
	if err != nil {
		return nil, err
	}

	response := make([]roomdto.Response, 0, len(rooms))
	for _, room := range rooms {
		response = append(response, toRoomResponse(room))
	}

	return response, nil
}

func (s *BookingService) Create(ctx context.Context, request bookingdto.Request) (bookingdto.Response, error) {
	roomID, guestIDs, start, end, err := s.parseRequest(request)
	if err != nil {
		return bookingdto.Response{}, err
	}

	roomEntity, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if roomEntity == nil {
		return bookingdto.Response{}, bookingdomain.ErrRoomNotFound
	}

	guests, err := s.validateGuests(ctx, guestIDs, roomEntity.Capacity)
	if err != nil {
		return bookingdto.Response{}, err
	}

	available, err := s.bookingRepo.IsRoomAvailable(ctx, roomID, start, end)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if !available {
		return bookingdto.Response{}, bookingdomain.ErrRoomNotAvailable
	}

	entity := model.Booking{
		ID:        uuid.New(),
		RoomID:    roomID,
		StartTime: start,
		EndTime:   end,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.bookingRepo.Insert(ctx, entity); err != nil {
		return bookingdto.Response{}, err
	}
	if err := s.bookingRepo.AddGuests(ctx, entity.ID, guestIDs); err != nil {
		return bookingdto.Response{}, err
	}

	if err := s.publishBookingEvent(ctx, kafkaevent.BookingEvent{
		EventType: kafkaevent.EventTypeCreated,
		BookingID: entity.ID,
		GuestIDs:  guestIDs,
		RoomID:    entity.RoomID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return bookingdto.Response{}, err
	}
	s.metrics.IncrementCreated()

	return toBookingResponse(entity, *roomEntity, guests), nil
}

func (s *BookingService) Update(ctx context.Context, rawID string, request bookingdto.Request) (bookingdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return bookingdto.Response{}, bookingdomain.ErrInvalidBookingID
	}

	existing, err := s.bookingRepo.FindByID(ctx, id)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if existing == nil {
		return bookingdto.Response{}, bookingdomain.ErrNotFound
	}

	roomID, guestIDs, start, end, err := s.parseRequest(request)
	if err != nil {
		return bookingdto.Response{}, err
	}

	roomEntity, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if roomEntity == nil {
		return bookingdto.Response{}, bookingdomain.ErrRoomNotFound
	}

	guests, err := s.validateGuests(ctx, guestIDs, roomEntity.Capacity)
	if err != nil {
		return bookingdto.Response{}, err
	}

	available, err := s.bookingRepo.IsRoomAvailableForUpdate(ctx, id, roomID, start, end)
	if err != nil {
		return bookingdto.Response{}, err
	}
	if !available {
		return bookingdto.Response{}, bookingdomain.ErrRoomNotAvailable
	}

	existing.RoomID = roomID
	existing.StartTime = start
	existing.EndTime = end

	if err := s.bookingRepo.Update(ctx, *existing); err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return bookingdto.Response{}, bookingdomain.ErrNotFound
		}
		return bookingdto.Response{}, err
	}
	if err := s.bookingRepo.DeleteGuestsByBookingID(ctx, id); err != nil {
		return bookingdto.Response{}, err
	}
	if err := s.bookingRepo.AddGuests(ctx, id, guestIDs); err != nil {
		return bookingdto.Response{}, err
	}

	if err := s.publishBookingEvent(ctx, kafkaevent.BookingEvent{
		EventType: kafkaevent.EventTypeUpdated,
		BookingID: existing.ID,
		GuestIDs:  guestIDs,
		RoomID:    existing.RoomID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return bookingdto.Response{}, err
	}
	s.metrics.IncrementUpdated()

	return toBookingResponse(*existing, *roomEntity, guests), nil
}

func (s *BookingService) Delete(ctx context.Context, rawID string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return bookingdomain.ErrInvalidBookingID
	}

	if err := s.bookingRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return bookingdomain.ErrNotFound
		}
		return err
	}

	return nil
}

func parseAndValidateRange(rawStart, rawEnd string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, rawStart)
	if err != nil {
		return time.Time{}, time.Time{}, bookingdomain.ErrInvalidTimeRange
	}

	end, err := time.Parse(time.RFC3339, rawEnd)
	if err != nil {
		return time.Time{}, time.Time{}, bookingdomain.ErrInvalidTimeRange
	}

	if !end.After(start) {
		return time.Time{}, time.Time{}, bookingdomain.ErrInvalidTimeRange
	}

	return start.UTC(), end.UTC(), nil
}

func (s *BookingService) parseRequest(request bookingdto.Request) (uuid.UUID, []uuid.UUID, time.Time, time.Time, error) {
	roomID, err := uuid.Parse(request.RoomID)
	if err != nil {
		return uuid.Nil, nil, time.Time{}, time.Time{}, bookingdomain.ErrInvalidRoomID
	}

	start, end, err := parseAndValidateRange(request.StartTime, request.EndTime)
	if err != nil {
		return uuid.Nil, nil, time.Time{}, time.Time{}, err
	}

	guestIDs := make([]uuid.UUID, 0, len(request.GuestIDs))
	for _, rawID := range request.GuestIDs {
		guestID, parseErr := uuid.Parse(rawID)
		if parseErr != nil {
			return uuid.Nil, nil, time.Time{}, time.Time{}, bookingdomain.ErrGuestNotFound
		}
		guestIDs = append(guestIDs, guestID)
	}

	return roomID, guestIDs, start, end, nil
}

func (s *BookingService) validateGuests(ctx context.Context, guestIDs []uuid.UUID, roomCapacity int) ([]model.Guest, error) {
	if len(guestIDs) > roomCapacity {
		return nil, bookingdomain.ErrRoomCapacityExceeded
	}

	guests, err := s.guestRepo.FindAllByIDs(ctx, guestIDs)
	if err != nil {
		return nil, err
	}
	if len(guests) != len(guestIDs) {
		return nil, bookingdomain.ErrGuestNotFound
	}

	return guests, nil
}

func (s *BookingService) findGuestsByBookingID(ctx context.Context, bookingID uuid.UUID) ([]model.Guest, error) {
	guestIDs, err := s.bookingRepo.FindGuestIDsByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	guests, err := s.guestRepo.FindAllByIDs(ctx, guestIDs)
	if err != nil {
		return nil, err
	}

	return guests, nil
}

func toBookingResponse(booking model.Booking, room model.Room, guests []model.Guest) bookingdto.Response {
	guestResponses := make([]guestdto.Response, 0, len(guests))
	for _, guest := range guests {
		guestResponses = append(guestResponses, toGuestResponse(guest))
	}

	return bookingdto.Response{
		ID:        booking.ID,
		Room:      toRoomResponse(room),
		Guests:    guestResponses,
		StartTime: booking.StartTime,
		EndTime:   booking.EndTime,
	}
}

func (s *BookingService) publishBookingEvent(ctx context.Context, bookingEvent kafkaevent.BookingEvent) error {
	if s.publisher == nil {
		return nil
	}

	if err := s.publisher.PublishBookingEvent(ctx, bookingEvent); err != nil {
		return fmt.Errorf("publish booking event: %w", err)
	}

	return nil
}
