package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/steadfastie/gokube/docs"
	"github.com/steadfastie/gokube/handlers"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Swagger for steadfastie/gokube project
//	@version		1.0
//	@contact.email	alexander.divovich@gmail.com

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	GitHub repository
// @externalDocs.url			https://github.com/Steadfastie/gokube
func main() {
	router := gin.Default()
	var api = router.Group("/api")
	{
		api.GET("/ping", handlers.EntryHandler)
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(":8080")
}
