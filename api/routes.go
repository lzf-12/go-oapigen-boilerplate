package api

import (
	"oapi-to-rest/api/auth"
	"oapi-to-rest/api/user"
	"oapi-to-rest/pkg/env"
	"oapi-to-rest/pkg/errlib"

	"github.com/gin-gonic/gin"
)

// represents the API server with all dependencies
type Server struct {
	Config *env.Config
	Router *gin.Engine

	// add dependencies here (DB clients, services, etc.)
	User user.ServerInterface
	Auth auth.ServerInterface

	// standardized error handler
	ErrorHandler errlib.ErrorHandler
}

// creates a new server instance
func NewServer(cfg *env.Config) *Server {

	dep := InitDependencies(cfg)

	userImpl := user.UserImpl{Db: dep.DbSqlite}
	userStrictHandler := user.NewStrictHandler(&userImpl, []user.StrictMiddlewareFunc{})

	authImpl := auth.AuthImpl{Db: dep.DbSqlite}
	authStrictHandler := auth.NewStrictHandler(&authImpl, []auth.StrictMiddlewareFunc{})

	return &Server{
		Config: cfg,
		Router: gin.New(),

		User: userStrictHandler,
		Auth: authStrictHandler,

		ErrorHandler: *dep.ErrorHandler,
	}
}

func (s *Server) RegisterRoutes() {

	s.Router.Use(gin.Logger())
	s.Router.Use(gin.Recovery())

	// standardized error response middleware
	s.Router.Use(errlib.ErrorHandlerGinMiddleware(s.ErrorHandler))

	api := s.Router.Group("api")
	v1 := api.Group("v1")

	user.RegisterHandlers(v1, s.User)
	auth.RegisterHandlers(v1, s.Auth)
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}
