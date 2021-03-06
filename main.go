package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"url-shortener/controller"
	"url-shortener/repository"
	"url-shortener/service"

	"github.com/spf13/viper"

	"github.com/swaggo/files"       // swagger embed files
	"github.com/swaggo/gin-swagger" // gin-swagger middleware

	_ "url-shortener/docs" // docs is generated by Swag CLI, you have to import it.
)

// @title Swagger Example API
// @version 1.0
// @description Basic url shortener.

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %v \n", err)
	}
}

func main() {
	redisAddress := viper.GetString("REDIS_ADDRESS")

	repo, err := repository.NewPool(redisAddress)
	if err != nil {
		log.Fatalf("failed to init repository, err: %v", err)
	}
	serv := service.New(repo)
	ctrl := controller.New(serv)

	url := ginSwagger.URL("doc.json") // The url pointing to API definition

	router := gin.Default()
	router.POST("/shorten", ctrl.Shorten)
	router.GET("/:shortCode", ctrl.Redirect)
	router.GET("/admin/urls", ctrl.GetUrls)
	router.DELETE("/:shortCode", ctrl.DeleteUrl)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	router.Run(":8080")
}
