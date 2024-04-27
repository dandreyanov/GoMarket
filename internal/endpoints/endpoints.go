package endpoints

import (
	"GoMarket/internal/handlers"
	"github.com/gin-gonic/gin"
)

func InitEndpoints(r *gin.Engine, pr *handlers.ProductRoutes, or *handlers.OrderRoutes) {
	r.POST("/product/add", pr.AddProduct)
	r.POST("/product/update/:id", pr.UpdateProduct)
	r.DELETE("/product/delete/:id", pr.DeleteProduct)
	r.GET("/product/all", pr.GetAllProducts)
	r.GET("/product/:id", pr.GetProductById)
	r.POST("/order/add", or.MakeOrder)
	r.GET("order/user/:id", or.GetUserOrders)
	r.GET("order/product/:id", or.GetProductOrders)
}
