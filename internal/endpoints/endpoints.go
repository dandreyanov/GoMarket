package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper/internal/handlers"
)

func InitEndpoints(r *gin.Engine, pr *handlers.ProductRoutes) {
	r.POST("/product/add", pr.AddProduct)
}
