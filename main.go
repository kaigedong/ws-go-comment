package main

import (
	"comment/dbops"
	"comment/handler"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	viper.SetDefault("port", 9913)
	viper.SetDefault("mongodbURL", "mongodb://127.0.0.1:27017/")
	viper.SetDefault("mongodbName", "dada")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}
	dbops.InitDBCollection()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	commentGroup := e.Group("/comment")
	commentHandler := handler.CommentHandler{}
	commentGroup.GET("/comments/:project_id", commentHandler.ProjectComments) // websocket
	commentGroup.GET("/all-comments/:project_id", commentHandler.AllComments)
	commentGroup.POST("/add-comment", commentHandler.AddComment)

	go func() {
		if err := e.Start(fmt.Sprintf(":%s", viper.GetString("port"))); err != nil {
			e.Logger.Fatal(err)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
