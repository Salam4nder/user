package grpc

import (
	"context"
	"errors"

	"github.com/Salam4nder/user/internal/db"
	"github.com/Salam4nder/user/internal/grpc/gen"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReadByEmail returns a user by email.
func (x *UserServer) ReadByEmail(
	ctx context.Context,
	req *gen.ReadByEmailRequest,
) (*gen.UserResponse, error) {
	if req == nil {
		return nil, requestIsNilError()
	}

	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email can not be empty")
	}

	user, err := x.storage.ReadUserByEmail(ctx, req.GetEmail())
	if err != nil {
		switch {
		case errors.Is(err, db.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())

		default:
			log.Error().Err(err).Msg("failed to get user by email")

			return nil, internalServerError()
		}
	}

	return userToProtoResponse(user), nil
}
