package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"onestarlab/backend/internal/api"
	"onestarlab/backend/internal/db"
	"onestarlab/backend/internal/store"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	port := env("PORT", "8080")
	dbPath := env("DATABASE_PATH", "data/onestar.db")
	webDir := os.Getenv("WEB_DIR")
	apiToken := os.Getenv("API_TOKEN")
	corsOrigin := env("CORS_ORIGINS", "*")

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer conn.Close()

	server := api.NewServer(store.New(conn), webDir, apiToken, corsOrigin)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("OneStarLab API listening on http://0.0.0.0:%s (db=%s, web=%q)", port, dbPath, webDir)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务启动失败: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("服务关闭异常: %v", err)
	}
	log.Println("OneStarLab API stopped")
}
