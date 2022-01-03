package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Server type
type Server struct {
	r       *chi.Mux
	handler *http.Handler
}

// NewServer - Constructor
func NewServer() *Server {
	return &Server{}
}

// SetupRoutes on specified port
func (s *Server) SetupRoutes(h *AccountHandler) *chi.Mux {
	s.r = chi.NewRouter()

	// A good base middleware stack
	s.r.Use(middleware.RequestID)
	s.r.Use(middleware.RealIP)
	s.r.Use(middleware.Logger)
	s.r.Use(middleware.Recoverer)
	s.r.Use(render.SetContentType(render.ContentTypeJSON))
	s.r.Use(middleware.Timeout(time.Minute))

	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"result\":\"OK\"}"))
	})

	CreateRoute(s.r, h)

	return s.r
}

// StartWebServer - on specified port
func (s *Server) StartWebServer(port string, h *AccountHandler, logger *log.Logger) {
	log.Println("starting web server ...")

	chiRouter := s.SetupRoutes(h)

	/*
		err := http.ListenAndServe(":"+port, r)

		if err != nil {
			log.Printf("An error starting HTTP server on port: %v", port)
			log.Printf("Error: %v", err.Error())
		}
	*/

	server := &http.Server{
		Handler:      chiRouter,
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		ErrorLog:     logger,
	}

	go func() {
		log.Println("Starting server")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("An error starting HTTP server on port: %v", port)
			// log.Printf("Error: %v", err.Error())
			log.Fatal(err)
		}
	}()

	s.WaitForShutdown(server)
}

// WaitForShutdown - Wait for the shutdown
func (s *Server) WaitForShutdown(server *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interruptChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	server.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
