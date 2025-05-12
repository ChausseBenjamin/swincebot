package intercept

import (
	"context"
	"log/slog"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/ChausseBenjamin/swincebot/internal/util"
	"github.com/hashicorp/go-uuid"
	"google.golang.org/grpc"
)

// gRPC interceptor to tag requests with a unique identifier and other unique attributes to ease logging
func Tagging(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		slog.ErrorContext(ctx, "Unable to generate UUID for request", logging.ErrKey, err)
	}
	ctx = context.WithValue(ctx, util.ReqIDKey, id)
	return handler(ctx, req)
}
