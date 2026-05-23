package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"RateLimiterAPI/controllers"
)

func main() {
	r1 := gin.Default()
	r1.POST("/request", controllers.RequestAdder)
	r1.GET("/stats", controllers.GetStats)

	go r1.Run(":8080")
	fmt.Println("Rate Limiter running on port 8080...")

	r2 := gin.Default()
	r2.POST("/products", controllers.CreateProduct)
	r2.GET("/products", controllers.GetProducts)
	
	r2.GET("/products/:id", controllers.GetProductByID)
	r2.POST("/products/:id/media", controllers.AddMedia)
	r2.Run(":8081")
}