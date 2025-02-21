package service

import (
	"context"
)

// BaseService defines the common operations all services must implement
type BaseService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Init(ctx context.Context) error
}
