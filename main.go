package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog-api/auth"
	"blog-api/handlers"
	"blog-api/logger"
	"blog-api/storage"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dataDir := "data"
	store := storage.NewFileStorage(dataDir)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key"
	}
	tokenAuth := auth.NewTokenAuth(jwtSecret)

	eventLogger := logger.NewEventLogger("log.txt")
	eventLogger.Start()

	h := handlers.NewHandler(store, tokenAuth, eventLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("POST /register", h.RegisterHandler)
	mux.HandleFunc("POST /login", h.LoginHandler)
	mux.HandleFunc("POST /posts", h.CreatePostHandler)
	mux.HandleFunc("GET /posts", h.GetPostsHandler)
	mux.HandleFunc("GET /posts/{id}", h.GetPostHandler)
	mux.HandleFunc("POST /posts/{id}/comments", h.CreateCommentHandler)
	mux.HandleFunc("GET /posts/{id}/comments", h.GetCommentsHandler)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		eventLogger.Stop()
		log.Println("Event logger stopped")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	fmt.Printf("Blog API server starting on %s\n", addr)
	log.Println("Available endpoints:")
	log.Println("  GET    /health")
	log.Println("  POST   /register")
	log.Println("  POST   /login")
	log.Println("  POST   /posts")
	log.Println("  GET    /posts")
	log.Println("  GET    /posts/{id}")
	log.Println("  POST   /posts/{id}/comments")
	log.Println("  GET    /posts/{id}/comments")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
