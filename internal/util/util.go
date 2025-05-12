package util

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrOutOfBoundsPort = errors.New("given port is out of bounds (1024-65535)")

type ContextKey uint8

const (
	DBKey ContextKey = iota
	ReqIDKey
)

type ParseUUIDParams struct {
	Str         string
	Subject     string
	Implication codes.Code
	Critical    bool
}

type ConfigStore struct {
	Admins      []uint64
	DBCacheSize int
}

func GetFromContext[T any](ctx context.Context, key any) *T {
	if value := ctx.Value(key); value != nil {
		if asserted, ok := value.(*T); ok {
			return asserted
		}
	}
	slog.WarnContext(ctx, "Failed to retrieve item from context", "key", key)
	return nil
}

// ParseUUID avoids doing the logging and error handling for protobuf every
// f*cking time a uuid is being passed as a string.
// - str is the string formatted uuid
// - subject is what the uuid represents (ex: user_id, task_id, jwt_id)
// - critical helps define how to handle errors around the token.
//   - user can f*ck up -> !critical as the error might not be our fault -> warning
//   - DBs shouldn't fuck up -> critical as this implies a greater internal issue -> error
//
// - implications define what an error implies for the user (which protobuf status code response)
func ParseUUID(ctx context.Context, p ParseUUIDParams) (uuid.UUID, error) {
	id, err := uuid.Parse(p.Str)
	if err != nil {
		log := slog.With(
			p.Subject, p.Str,
			"error_message", err,
		)
		if p.Critical {
			log.ErrorContext(ctx, "failed to parse provided UUID")
			return uuid.UUID{}, status.Errorf(p.Implication,
				"failed to parse %s: '%s'", p.Subject, p.Str,
			)
		} else {
			log.WarnContext(ctx, "failed to parse provided UUID")
			return uuid.UUID{}, status.Errorf(p.Implication,
				"failed to parse %s: '%s'", p.Subject, p.Str,
			)
		}
	}
	return id, nil
}
