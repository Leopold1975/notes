package notes

import (
	"errors"

	"notes/internal/notes/app"
	"notes/internal/notes/controller"
	"notes/internal/notes/server/ginserver"
	"notes/internal/notes/server/grpcserver"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
)

const (
	GRPCAPI = "grpc"
	RestAPI = "rest"
)

var ErrServerUnspecified = errors.New("server unspecified")

func New(serv string, cfg config.Config, app app.App, logg logger.Logger) (controller.API, error) {
	switch serv {
	case GRPCAPI:
		s := grpcserver.New(app, logg, cfg.GRPCServer)
		return s, nil
	case RestAPI:
		s := ginserver.New(app, cfg.Server, logg)
		return s, nil
	}
	return nil, ErrServerUnspecified
}
