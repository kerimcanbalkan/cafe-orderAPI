package test

import (
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/order"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/user"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

type UsersResponse struct {
	Data []user.User `json:"data"`
}
type OrderResponse struct {
	Data []order.Order `json:"data"`
}
type MenuResponse struct {
	Data []menu.MenuItem `json:"data"`
}
