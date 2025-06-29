package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/vesselchuckk/go-social/cmd/api/config"
	"github.com/vesselchuckk/go-social/internal/auth"
	"github.com/vesselchuckk/go-social/internal/mails"
	"github.com/vesselchuckk/go-social/internal/store"
	"github.com/vesselchuckk/go-social/internal/store/cache"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type Server struct {
	Config  *config.Config
	Store   *store.Store
	Logger  *zap.SugaredLogger
	Mailer  *mails.SendGridMailer
	JWTAuth *auth.JWTAuth
	Redis   *cache.Storage
}

func NewServer(cfg *config.Config, db *store.Store, logger *zap.SugaredLogger, mailer *mails.SendGridMailer, auth *auth.JWTAuth, rdb *redis.Client) *Server {
	return &Server{
		Config:  cfg,
		Store:   db,
		Logger:  logger,
		Mailer:  mailer,
		JWTAuth: auth,
		Redis:   cache.NewCacheStore(rdb),
	}
}

func (s *Server) Run() error {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Route("/v1", func(router chi.Router) {
		router.With(s.BasicAuth()).Get("/health", s.healthHandler)

		router.Route("/posts", func(router chi.Router) {
			router.Use(s.AuthMiddleware)
			router.Post("/", s.createPostHandler)

			router.Route("/{postID}", func(router chi.Router) {
				router.Use(s.postContextFetch)

				router.Get("/", s.getPostByID)
				router.Delete("/", s.checkPostOwnership("admin", s.deletePostHandler))
				router.Patch("/", s.checkPostOwnership("moderator", s.updatePostHandler))
			})
		})

		router.Route("/users", func(router chi.Router) {
			router.Put("/activate/{token}", s.activateUser)

			router.Route("/{userID}", func(router chi.Router) {
				router.Use(s.AuthMiddleware)
				router.Use(s.userContext)

				router.Get("/", s.getUserHandler)

				router.Put("/follow", s.followUserHandler)
				router.Put("/unfollow", s.unfollowUserHandler)
			})

			router.Group(func(router chi.Router) {
				router.Use(s.AuthMiddleware)
				router.Get("/feed", s.getUserFeed)
			})
		})

		//pub
		router.Route("/auth", func(router chi.Router) {
			router.Post("/user", s.registerHandler)
			router.Post("/token", s.createTokenHandler)
		})

	})

	srv := &http.Server{
		Addr:    net.JoinHostPort(s.Config.ServerHost, s.Config.ServerPort),
		Handler: middleware.Logger(router),
	}
	s.Logger.Infow("sever has started", "addr", s.Config.ServerHost, "env", s.Config.ENV)

	return srv.ListenAndServe()
}
