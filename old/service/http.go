package service

import "net/http"

type HTTPService interface {
	GetHttpRouter() map[string]http.HandlerFunc
	Service
}
