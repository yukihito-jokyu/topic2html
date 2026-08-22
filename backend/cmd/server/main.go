package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	ginadapter "github.com/yukihito-jokyu/topic2html/backend/handler/gin"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	"github.com/yukihito-jokyu/topic2html/backend/repository/google"
	"github.com/yukihito-jokyu/topic2html/backend/repository/postgres"
	"github.com/yukihito-jokyu/topic2html/backend/repository/security"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

var (
	lookupEnvironment           = os.LookupEnv
	listenAndServe              = (*http.Server).ListenAndServe
	serverLogWriter   io.Writer = os.Stderr
	exitProcess                 = os.Exit
	loadEnvironment             = func() error { return loadDotEnv(func(filename string) error { return godotenv.Load(filename) }) }
	runServer                   = run
)

type dependencies struct {
	newPool       func(context.Context, string) (*pgxpool.Pool, error)
	closePool     func(*pgxpool.Pool)
	newProtection func(string) (*security.Service, error)
	newService    func(usecaseauth.Dependencies, string, string) (*usecaseauth.Service, error)
	verifyBroker  func(context.Context, string) error
}

func productionDependencies() dependencies {
	return dependencies{
		newPool:       pgxpool.New,
		closePool:     (*pgxpool.Pool).Close,
		newProtection: security.New,
		newService:    usecaseauth.NewService,
		verifyBroker:  brokerVerifier(),
	}
}

func main() {
	start()
}

func start() {
	if err := loadEnvironment(); err != nil {
		observability.NewLogger(serverLogWriter).Error(context.Background(), "server.start.failed", err)
		exitProcess(1)

		return
	}
	if err := runServer(lookupEnvironment, listenAndServe); err != nil {
		observability.NewLogger(serverLogWriter).Error(context.Background(), "server.start.failed", err)
		exitProcess(1)
	}
}

func loadDotEnv(loadFile func(string) error) error {
	for _, filename := range []string{"../.env", ".env"} {
		err := loadFile(filename)
		if err == nil {
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func run(lookup LookupEnv, serve func(*http.Server) error) error {
	return runWithDependencies(lookup, serve, productionDependencies())
}

func runWithDependencies(lookup LookupEnv, serve func(*http.Server) error, dependencies dependencies) error {
	config, err := loadConfig(lookup)
	if err != nil {
		return apperr.New(apperr.CodeInvalidConfiguration)
	}
	if err := dependencies.verifyBroker(context.Background(), config.CodexBrokerEndpoint); err != nil {
		return apperr.New(apperr.CodeUnavailable)
	}
	database, err := dependencies.newPool(context.Background(), config.DatabaseURL)
	if err != nil {
		return apperr.New(apperr.CodeUnavailable)
	}
	defer dependencies.closePool(database)
	protectedRecords := postgres.NewStore(database)
	protection, err := dependencies.newProtection(config.ProtectionKey)
	if err != nil {
		return apperr.New(apperr.CodeUnavailable)
	}
	provider := google.NewProvider(google.NewClient(nil), google.ProviderConfig{
		ClientID:          config.GoogleClientID,
		ClientSecret:      config.GoogleSecret,
		RedirectURI:       config.OAuthCallbackURI,
		DiscoveryEndpoint: config.GoogleDiscoveryEndpoint,
	})
	logger := observability.NewLogger(os.Stderr)
	service, err := dependencies.newService(usecaseauth.Dependencies{
		Store:    protectedRecords,
		Provider: provider,
		Security: protection,
		Clock:    usecaseauth.SystemClock{},
		Logger:   logger,
	}, config.TrustedAppOrigin, config.AllowedEmail)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           ginadapter.NewRouter(service, service, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := serve(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return apperr.New(apperr.CodeUnavailable)
	}

	return nil
}
