package identity

import (
	"context"
	"time"

	Identity "github.com/Diogo1080/GoLearning-IdentityMicroService/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client Identity.InternalIdentityServiceClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	return &Client{
		client: Identity.NewInternalIdentityServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*Identity.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ValidateToken(ctx, &Identity.ValidateTokenRequest{
		Token: token,
	})
}
