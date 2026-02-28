package aiclient

import (
	proto "bot/protoai"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultTimeout = 60 * time.Second

type Client struct {
	conn   *grpc.ClientConn
	client proto.OrderServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to grpc server")
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: proto.NewOrderServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// ===== Wrapper for OrderService =====

func (c *Client) Get(ctx context.Context, req *proto.GetRequest, opts ...grpc.CallOption) (*proto.GetResponse, error) {

	// если у входящего ctx нет дедлайна — ставим дефолтный
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	return c.client.Get(ctx, req, opts...)
}
