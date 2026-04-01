package main

import (
	"web_backend/internal/app/config"
	"web_backend/internal/app/dsn"
	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"
	"web_backend/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	logrus.Info("DSN: ", postgresString)

	rep, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
