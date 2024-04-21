package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Salam4nder/user/internal/db"
	"github.com/Salam4nder/user/internal/grpc/gen"
	"github.com/Salam4nder/user/pkg/util"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCode "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	serverStr  string = "UserServer"
	handlerStr string = "CreateUser"
)

var tracer = otel.Tracer(serverStr)

func (x *UserServer) CreateUser(ctx context.Context, req *gen.CreateUserRequest) (*gen.UserID, error) {
	ctx, span := tracer.Start(
		ctx,
		handlerStr,
		trace.WithAttributes(attribute.String("full_name", req.FullName), attribute.String("email", req.Email)),
	)
	defer span.End()

	if err := validateCreateUserRequest(req); err != nil {
		span.SetStatus(otelCode.Error, "grpc: invalid request")
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := db.CreateUserParams{
		FullName:  req.GetFullName(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		CreatedAt: time.Now(),
	}
	createdUser, err := x.storage.CreateUser(ctx, params)
	if err != nil {
		span.SetStatus(otelCode.Error, "grpc: creating user")
		span.RecordError(err)
		if errors.Is(err, db.ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, internalServerError()
	}

	return &gen.UserID{Id: createdUser.ID.String()}, nil
}

// validateCreateUserRequest returns nil if the request is valid.
func validateCreateUserRequest(req *gen.CreateUserRequest) error {
	if req == nil {
		return errors.New("grpc: request can not be nil")
	}
	requestIsNilError()

	var (
		fullNameErr error
		emailErr    error
		passwordErr error
	)

	if err := util.ValidateFullName(req.GetFullName()); err != nil {
		fullNameErr = fmt.Errorf("grpc: full_name %w", err)
	}

	if err := util.ValidateEmail(req.GetEmail()); err != nil {
		emailErr = fmt.Errorf("grpc: email %w", err)
	}

	if err := util.ValidatePassword(req.GetPassword()); err != nil {
		passwordErr = fmt.Errorf("grpc: password %w", err)
	}

	return errors.Join(fullNameErr, emailErr, passwordErr)
}

func userToProtoResponse(user *db.User) *gen.UserResponse {
	return &gen.UserResponse{
		Id:        user.ID.String(),
		FullName:  user.FullName,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}
