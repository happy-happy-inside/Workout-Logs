package aiclient

import (
	pb "bot/protoai"
	proto "bot/protoai"

	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	pbproto "google.golang.org/protobuf/proto"
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

	// если нет дедлайна — ставим дефолтный
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	// correlation id
	reqID := uuid.New().String()
	req.Reqid = reqID

	data, err := pbproto.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(reqID),
		Value: data,
	})
	if err != nil {
		return nil, err
	}

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return nil, err
		}

		if string(msg.Key) != reqID {
			continue
		}

		var resp pb.GetResponse
		if err := pbproto.Unmarshal(msg.Value, &resp); err != nil {
			return nil, err
		}

		if resp.Error != "" {
			return nil, errors.New(resp.Error)
		}

		return &resp, nil
	}
}
