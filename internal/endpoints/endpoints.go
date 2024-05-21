package endpoints

import (
	"GoMarket/internal/handlers"
	"github.com/gin-gonic/gin"
)

func InitEndpoints(r *gin.Engine, pr *handlers.ProductRoutes, or *handlers.OrderRoutes, ur *handlers.UserRoutes) {
	productGroup := r.Group("/product")
	productGroup.Use(ur.AuthMiddleware())
	{
		productGroup.POST("/add", pr.AddProduct)
		productGroup.POST("/update/:id", pr.UpdateProduct)
		productGroup.DELETE("/delete/:id", pr.DeleteProduct)
		productGroup.GET("/all", pr.GetAllProducts)
		productGroup.GET("/:id", pr.GetProductById)
	}

	orderGroup := r.Group("/order")
	orderGroup.Use(ur.AuthMiddleware())
	{
		orderGroup.POST("/add", or.MakeOrder)
		orderGroup.GET("/user/:id", or.GetUserOrders)
		orderGroup.GET("/product/:id", or.GetProductOrders)
	}

	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", ur.RegisterUser)
		userGroup.POST("/login", ur.LoginUser)
	}
}
