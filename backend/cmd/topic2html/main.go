package main

import (
	"log"
	"net/http"
	"time"

	"github.com/yukihito-jokyu/topic2html/backend/internal/config"
	httphandler "github.com/yukihito-jokyu/topic2html/backend/internal/handler/http"
	"github.com/yukihito-jokyu/topic2html/backend/internal/usecase"
)

func main() {
	settings := config.Load()
	readiness := usecase.NewReadinessService()

	server := &http.Server{
		Addr:              settings.ListenAddress,
		Handler:           httphandler.NewRouter(readiness),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
