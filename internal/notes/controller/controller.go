package controller

import "context"

type API interface {
	Starter
	Shutdowner
}

type Starter interface {
	Start(context.Context) error
}

type Shutdowner interface {
	Shutdown(context.Context) error
}
