package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	ginadapter "github.com/yukihito-jokyu/topic2html/backend/handler/gin"
)

var (
	lookupEnvironment = os.LookupEnv
	listenAndServe    = (*http.Server).ListenAndServe
	printError        = log.Print
	exitProcess       = os.Exit
)

func main() {
	start()
}

func start() {
	if err := run(lookupEnvironment, listenAndServe); err != nil {
		printError(err)
		exitProcess(1)
	}
}

func run(lookup LookupEnv, serve func(*http.Server) error) error {
	if _, err := loadConfig(lookup); err != nil {
		return errors.New("server configuration is invalid")
	}
	server := &http.Server{Addr: "127.0.0.1:8080", Handler: ginadapter.NewRouter(), ReadHeaderTimeout: 5 * time.Second}
	if err := serve(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
