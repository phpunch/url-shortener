package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"url-shortener/controller"
	"url-shortener/repository"
	"url-shortener/service"

	"github.com/spf13/viper"
)

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

	router := gin.Default()
	router.POST("/shorten", ctrl.Shorten)
	router.GET("/:shortCode", ctrl.Redirect)
	router.Run(":8080")
}
