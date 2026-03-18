package api

import (
	"fmt"
	"html/template"
	"log"

	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StartServer инициализирует репозиторий, обработчики и запускает HTTP-сервер.
func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"printf": fmt.Sprintf,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", h.GetDrugs)
	r.GET("/drugs/:id", h.GetDrug)
	r.GET("/prescriptions/:id", h.GetPrescription)

	r.Run()
	log.Println("Server down")
}
