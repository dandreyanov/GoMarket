package endpoints

import (
	"GoMarket/internal/handlers"
	"github.com/gin-gonic/gin"
)

func InitEndpoints(r *gin.Engine, pr *handlers.ProductRoutes) {
	r.POST("/product/add", pr.AddProduct)
	r.POST("/product/update/:id", pr.UpdateProduct)
	r.DELETE("/product/delete/:id", pr.DeleteProduct)
	r.GET("/product/all", pr.GetAllProducts)
	r.GET("/product/:id", pr.GetProductById)
}
