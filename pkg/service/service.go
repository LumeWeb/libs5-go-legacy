package service

import (
	"context"
	"go.lumeweb.com/libs5-go/pkg/config"
	"go.lumeweb.com/libs5-go/pkg/encoding"
	"go.lumeweb.com/libs5-go/pkg/kv"
	"go.lumeweb.com/libs5-go/pkg/registry"
	"go.lumeweb.com/libs5-go/pkg/storage"
	"go.lumeweb.com/libs5-go/pkg/structs"
	"go.lumeweb.com/libs5-go/pkg/transport"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"old/metadata"
	"sync"
)

// BaseService defines the common operations all services must implement
type BaseService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Init(ctx context.Context) error
}
