package identity

import (
	"context"
	"fmt"
	"time"

	Identity "github.com/Diogo1080/GoLearning-IdentityMicroService/api/v1"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client Identity.InternalIdentityServiceClient
	conn   *grpc.ClientConn
}

func NewClient(addr string, metrics ...*observability.Metrics) (*Client, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	if len(metrics) > 0 && metrics[0] != nil {
		options = append(options, grpc.WithUnaryInterceptor(observability.UnaryClientMetrics(metrics[0])))
	}
	conn, err := grpc.NewClient(addr, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		client: Identity.NewInternalIdentityServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Ready(ctx context.Context) error {
	return c.WaitReady(ctx)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) WaitReady(ctx context.Context) error {
	c.conn.Connect()

	for {
		state := c.conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.Shutdown {
			return fmt.Errorf("identity service connection is shut down")
		}
		if !c.conn.WaitForStateChange(ctx, state) {
			return fmt.Errorf("identity service is not ready: %w", ctx.Err())
		}
	}
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*Identity.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ValidateToken(ctx, &Identity.ValidateTokenRequest{
		Token: token,
	})
}
