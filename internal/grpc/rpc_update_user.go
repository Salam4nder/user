package grpc

import (
	"context"
	"errors"

	"github.com/Salam4nder/user/internal/db"
	"github.com/Salam4nder/user/internal/grpc/gen"
	"github.com/Salam4nder/user/pkg/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateUser updates a user by ID.
func (x *UserServer) UpdateUser(
	ctx context.Context,
	req *gen.UpdateUserRequest,
) (*gen.UserResponse, error) {
	authPayload, err := x.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err = validateUpdateUserRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if authPayload.Email != req.GetEmail() {
		return nil, status.Errorf(
			codes.PermissionDenied,
			"invoker is not owner of provided email",
		)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ID is invalid")
	}

	params := db.UpdateUserParams{
		ID:       id,
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
	}

	updatedUser, err := x.storage.UpdateUser(ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())

		default:
			log.Error().Err(err).Msg("grpc: failed to update user")
			return nil, internalServerError()
		}
	}

	return userToProtoResponse(updatedUser), nil
}

// validateUpdateUserRequest returns nil if the request is valid.
func validateUpdateUserRequest(req *gen.UpdateUserRequest) error {
	if req == nil {
		return errors.New("request can not be nil")
	}

	if req.Id == "" {
		return errors.New("ID can not be empty")
	}

	var (
		fullNameErr error
		emailErr    error
	)

	if req.GetFullName() != "" {
		if err := util.ValidateFullName(req.GetFullName()); err != nil {
			fullNameErr = err
		}
	}

	if req.GetEmail() != "" {
		if err := util.ValidateEmail(req.GetEmail()); err != nil {
			emailErr = err
		}
	}

	return errors.Join(fullNameErr, emailErr)
}
