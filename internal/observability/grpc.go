package observability

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryClientMetrics(metrics *Metrics) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, response interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, options ...grpc.CallOption) error {
		metrics.GRPCInFlight.Inc()
		started := time.Now()
		err := invoker(ctx, method, request, response, conn, options...)
		defer metrics.GRPCInFlight.Dec()

		code := status.Code(err).String()
		metrics.ObserveGRPC(method, code, time.Since(started))
		return err
	}
}
