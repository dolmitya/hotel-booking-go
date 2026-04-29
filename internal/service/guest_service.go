package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	guestdomain "hotelService/internal/domain/guest"
	guestdto "hotelService/internal/dto/guest"
	"hotelService/internal/kafka"
	kafkaevent "hotelService/internal/kafka/event"
	"hotelService/internal/model"
	"hotelService/internal/repository"
)

const guestBirthDateLayout = "2006-01-02"

type GuestService struct {
	repo      guestRepository
	publisher kafka.GuestEventPublisher
}

func NewGuestService(repo guestRepository, publisher kafka.GuestEventPublisher) *GuestService {
	return &GuestService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *GuestService) GetByID(ctx context.Context, rawID string) (guestdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return guestdto.Response{}, guestdomain.ErrInvalidGuestID
	}

	guest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return guestdto.Response{}, err
	}
	if guest == nil {
		return guestdto.Response{}, guestdomain.ErrNotFound
	}

	return toGuestResponse(*guest), nil
}

func (s *GuestService) Create(ctx context.Context, request guestdto.Request) (guestdto.Response, error) {
	birthDate, err := parseBirthDate(request.BirthDate)
	if err != nil {
		return guestdto.Response{}, err
	}

	guest := model.Guest{
		ID:          uuid.New(),
		LastName:    request.LastName,
		FirstName:   request.FirstName,
		MiddleName:  request.MiddleName,
		BirthDate:   birthDate,
		PhoneNumber: request.PhoneNumber,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Insert(ctx, guest); err != nil {
		return guestdto.Response{}, err
	}

	if err := s.publishGuestEvent(ctx, kafkaevent.GuestEvent{
		EventType: kafkaevent.EventTypeCreated,
		GuestID:   guest.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return guestdto.Response{}, err
	}

	return toGuestResponse(guest), nil
}

func (s *GuestService) Update(ctx context.Context, rawID string, request guestdto.Request) (guestdto.Response, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return guestdto.Response{}, guestdomain.ErrInvalidGuestID
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return guestdto.Response{}, err
	}
	if existing == nil {
		return guestdto.Response{}, guestdomain.ErrNotFound
	}

	birthDate, err := parseBirthDate(request.BirthDate)
	if err != nil {
		return guestdto.Response{}, err
	}

	existing.LastName = request.LastName
	existing.FirstName = request.FirstName
	existing.MiddleName = request.MiddleName
	existing.BirthDate = birthDate
	existing.PhoneNumber = request.PhoneNumber

	if err := s.repo.Update(ctx, *existing); err != nil {
		if errors.Is(err, repository.ErrGuestNotFound) {
			return guestdto.Response{}, guestdomain.ErrNotFound
		}
		return guestdto.Response{}, err
	}

	if err := s.publishGuestEvent(ctx, kafkaevent.GuestEvent{
		EventType: kafkaevent.EventTypeUpdated,
		GuestID:   existing.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return guestdto.Response{}, err
	}

	return toGuestResponse(*existing), nil
}

func parseBirthDate(value string) (time.Time, error) {
	birthDate, err := time.Parse(guestBirthDateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: expected yyyy-mm-dd", guestdomain.ErrInvalidBirthDate)
	}

	return birthDate.UTC(), nil
}

func toGuestResponse(guest model.Guest) guestdto.Response {
	return guestdto.Response{
		ID:          guest.ID,
		LastName:    guest.LastName,
		FirstName:   guest.FirstName,
		MiddleName:  guest.MiddleName,
		BirthDate:   guest.BirthDate.Format(guestBirthDateLayout),
		PhoneNumber: guest.PhoneNumber,
	}
}

func (s *GuestService) publishGuestEvent(ctx context.Context, guestEvent kafkaevent.GuestEvent) error {
	if s.publisher == nil {
		return nil
	}

	if err := s.publisher.PublishGuestEvent(ctx, guestEvent); err != nil {
		return fmt.Errorf("publish guest event: %w", err)
	}

	return nil
}
