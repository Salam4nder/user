package grpc

import (
	"time"

	"github.com/Salam4nder/user/internal/db"
	"github.com/Salam4nder/user/internal/grpc/gen"
	"github.com/Salam4nder/user/pkg/token"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/health"
)

// UserServer contains all necessary dependencies to serve user requests.
type UserServer struct {
	gen.UserServer

	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration

	storage    db.Storage
	tokenMaker token.Maker
	health     *health.Server
	natsConn   *nats.Conn
}

// NewUserServer returns a new UserService.
func NewUserServer(
	store db.Storage,
	health *health.Server,
	natsConn *nats.Conn,
	symmetricKey string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) (*UserServer, error) {
	tokenMaker, err := token.NewPasetoMaker(symmetricKey)
	if err != nil {
		return nil, err
	}

	return &UserServer{
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,

		storage:    store,
		tokenMaker: tokenMaker,
		health:     health,
		natsConn:   natsConn,
	}, nil
}
