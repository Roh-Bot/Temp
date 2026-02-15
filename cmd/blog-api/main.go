package main

import (
	"log"

	"github.com/Roh-Bot/blog-api/cmd/api"
	_ "github.com/Roh-Bot/blog-api/docs"
	servicesv1 "github.com/Roh-Bot/blog-api/internal/application"
	"github.com/Roh-Bot/blog-api/internal/auth"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/database"
	"github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/internal/validator"
	"github.com/Roh-Bot/blog-api/internal/worker"
	"github.com/Roh-Bot/blog-api/pkg/global"
	"github.com/Roh-Bot/blog-api/pkg/logger"
)

// @title Task Management API
// @version 1.0
// @description Task Management REST API with JWT authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host
// @BasePath /api
// @schemes https http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token
func main() {
	global.ParseFlags()
	appCtx := global.NewApplicationContext()

	appCtx.Add(1)
	cfg, err := config.LoadConfiguration(appCtx.Context())
	if err != nil {
		log.Fatal(err)
	}

	newLogger, err := logger.ZapNew(cfg.Get().Logger)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := newLogger.Flush()
		if err != nil {
			log.Fatal(err)
		}
	}()

	db, err := database.NewMasterConnection(cfg.Get().Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Flush()

	newStore := store.NewStorage(db, cfg)

	jwt := auth.NewJWTAuthenticator(cfg, newStore)
	auth2 := auth.NewAuthentication(jwt)

	services := servicesv1.NewService(cfg, auth2, newStore, newLogger)
	validator2 := validator.NewValidator()

	taskWorker := worker.NewTaskWorker(newStore.Tasks, newLogger, cfg.Get().AutoCompleteMin)
	appCtx.Add(1)
	taskWorker.Start(appCtx.Context())

	server := api.NewServer(cfg, services, validator2, newLogger, appCtx)

	appCtx.Add(1)
	go server.Run()

	appCtx.HandleShutdownSignal()
	appCtx.WaitForShutdown()

	log.Println("Goodbye")
}
