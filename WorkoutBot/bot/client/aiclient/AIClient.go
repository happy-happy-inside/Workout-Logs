package aiclient

import (
	proto "bot/protoai"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

const defaultTimeout = 60 * time.Second

type Client struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewClient(broker string) (*Client, error) {

	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "ai.requests",
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "ai.responses",
		GroupID: "bot-group",
	})

	return &Client{
		writer: writer,
		reader: reader,
	}, nil
}

func (c *Client) Close() error {
	if err := c.writer.Close(); err != nil {
		return err
	}
	return c.reader.Close()
}

func (c *Client) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	reqID := uuid.New().String()

	// добавляем correlation id в metadata
	req.RequestId = reqID

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(reqID),
		Value: data,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to send kafka message")
		return nil, err
	}

	// ждём ответ
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return nil, err
		}

		if string(msg.Key) != reqID {
			continue
		}

		var resp proto.GetResponse
		if err := proto.Unmarshal(msg.Value, &resp); err != nil {
			return nil, err
		}

		if resp.Error != "" {
			return nil, errors.New(resp.Error)
		}

		return &resp, nil
	}
}