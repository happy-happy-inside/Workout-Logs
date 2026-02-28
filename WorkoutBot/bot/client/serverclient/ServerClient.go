package serverclient

import (
	"bot/proto"
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
		log.Print("oshibka porluch")
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

// func wrappers
func (c *Client) AddRes(ctx context.Context, req *proto.AddResRequest) (*proto.AddResResponse, error) {
	return c.client.AddRes(ctx, req)
}

func (c *Client) GetRes(ctx context.Context, req *proto.GetResRequest) (*proto.GetResResponse, error) {
	return c.client.GetRes(ctx, req)
}

func (c *Client) TopUsers(ctx context.Context, req *proto.Uprajnenie) (*proto.Top, error) {
	return c.client.TopUsers(ctx, req)
}

func (c *Client) Stat(ctx context.Context, in *proto.StatRequest, opts ...grpc.CallOption) (*proto.StatResponse, error) {
	return c.client.Stat(ctx, in)
}
