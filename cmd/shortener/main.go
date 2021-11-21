package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/DrGermanius/Shortener/internal/app/handlers"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

func main() {
	store.InitLinksMap()

	r := gin.Default()

	r.Any("/*path", gin.WrapF(handlers.ShortenerHandler))

	log.Println("API started on " + handlers.Port)
	log.Fatalln(r.Run(":" + handlers.Port))
}
