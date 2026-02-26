package client

import (
	"ai-saas/proto"
	"context"
	"time"

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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.AddRes(ctx, req)
}

func (c *Client) GetRes(ctx context.Context, req *proto.GetResRequest) (*proto.GetResResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetRes(ctx, req)
}

func (c *Client) TopUsers(ctx context.Context, req *proto.Uprajnenie) (*proto.Top, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.TopUsers(ctx, req)
}
