package api

import (
	"fmt"
	"oapi-to-rest/api/order"
	"oapi-to-rest/api/user"
	"oapi-to-rest/pkg/env"

	"github.com/gin-gonic/gin"
)

// represents the API server with all dependencies
type Server struct {
	Router *gin.Engine

	// add dependencies here (DB clients, services, etc.)
	User  user.ServerInterface
	Order order.ServerInterface
}

// creates a new server instance
func NewServer(cfg *env.Config) *Server {

	dep := InitDependencies(cfg)

	userImpl := user.UserImpl{Db: dep.DbSqlite}
	userStrictHandler := user.NewStrictHandler(&userImpl, []user.StrictMiddlewareFunc{})

	orderImpl := order.OrderImpl{Db: dep.DbSqlite}
	orderStrictHandler := order.NewStrictHandler(&orderImpl, []order.StrictMiddlewareFunc{})

	return &Server{
		Router: gin.New(),

		User:  userStrictHandler,
		Order: orderStrictHandler,
	}
}

func (s *Server) RegisterRoutes() {

	s.Router.Use(gin.Logger())
	s.Router.Use(gin.Recovery())

	api := s.Router.Group("api")
	v1 := api.Group("v1")

	user.RegisterHandlers(v1, s.User)
	order.RegisterHandlers(v1, s.Order)
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}

func (s *Server) PrintRoutes() {
	for _, route := range s.Router.Routes() {
		fmt.Printf("%s %s\n", route.Method, route.Path)
	}
}
