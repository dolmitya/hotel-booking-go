package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"

	"hotelService/internal/config"
	"hotelService/internal/kafka/event"
)

type GuestEventPublisher interface {
	PublishGuestEvent(ctx context.Context, guestEvent event.GuestEvent) error
}

type RoomEventPublisher interface {
	PublishRoomEvent(ctx context.Context, roomEvent event.RoomEvent) error
}

type Producer struct {
	brokers     []string
	topics      config.KafkaTopicsConfig
	guestWriter *kafka.Writer
	roomWriter  *kafka.Writer
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		brokers: cfg.BootstrapServers,
		topics:  cfg.Topics,
		guestWriter: &kafka.Writer{
			Addr:         kafka.TCP(cfg.BootstrapServers...),
			Topic:        cfg.Topics.GuestEvents,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.LeastBytes{},
		},
		roomWriter: &kafka.Writer{
			Addr:         kafka.TCP(cfg.BootstrapServers...),
			Topic:        cfg.Topics.RoomEvents,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) EnsureTopics(ctx context.Context) error {
	if p == nil || len(p.brokers) == 0 {
		return nil
	}

	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", p.brokers[0], err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get kafka controller: %w", err)
	}

	address := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := kafka.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("dial kafka controller %s: %w", address, err)
	}
	defer controllerConn.Close()

	topics := []kafka.TopicConfig{
		{
			Topic:             p.topics.GuestEvents,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             p.topics.RoomEvents,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	if p.topics.BookingEvents != "" {
		topics = append(topics, kafka.TopicConfig{
			Topic:             p.topics.BookingEvents,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if err := controllerConn.CreateTopics(topics...); err != nil {
		return fmt.Errorf("create kafka topics: %w", err)
	}

	return nil
}

func (p *Producer) PublishGuestEvent(ctx context.Context, guestEvent event.GuestEvent) error {
	if p == nil || p.guestWriter == nil {
		return nil
	}

	payload, err := json.Marshal(guestEvent)
	if err != nil {
		return fmt.Errorf("marshal guest event: %w", err)
	}

	if err := p.guestWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(guestEvent.GuestID.String()),
		Value: payload,
	}); err != nil {
		return fmt.Errorf("write guest event: %w", err)
	}

	return nil
}

func (p *Producer) PublishRoomEvent(ctx context.Context, roomEvent event.RoomEvent) error {
	if p == nil || p.roomWriter == nil {
		return nil
	}

	payload, err := json.Marshal(roomEvent)
	if err != nil {
		return fmt.Errorf("marshal room event: %w", err)
	}

	if err := p.roomWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(roomEvent.RoomID.String()),
		Value: payload,
	}); err != nil {
		return fmt.Errorf("write room event: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if p == nil {
		return nil
	}

	if p.guestWriter != nil {
		if err := p.guestWriter.Close(); err != nil {
			return err
		}
	}

	if p.roomWriter != nil {
		if err := p.roomWriter.Close(); err != nil {
			return err
		}
	}

	return nil
}
