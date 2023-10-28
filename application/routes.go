package application

import (
	"net/http"

	"github.com/PRS1161/go-micro-service/handler"
	"github.com/PRS1161/go-micro-service/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", a.loadOrderRoutes)

	a.router = router
}

func (a *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &handler.Order{
		Helper: &helpers.RedisHelper{Client: a.rdb},
	}

	router.Post("/", orderHandler.GenerateOrder)
	router.Get("/", orderHandler.GetOrders)
	router.Get("/{id}", orderHandler.GetSingleOrder)
	router.Put("/{id}", orderHandler.UpdateOrder)
	router.Delete("/{id}", orderHandler.RemoveOrder)
}
