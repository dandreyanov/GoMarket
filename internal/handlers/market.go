package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper/internal/entity"
	"net/http"
)

type ProductRoutes struct {
	db *sql.DB
}

func NewProductRoutes(database *sql.DB) *ProductRoutes {
	return &ProductRoutes{
		db: database,
	}
}

func (p *ProductRoutes) AddProduct(c *gin.Context) {
	var product entity.Product
	err := c.BindJSON(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	product.ID = uuid.New().String()
	_, err = p.db.Exec("INSERT INTO products (id, name, description, price, quantity) VALUES ($1, $2, $3, $4, $5)", product.ID, product.Name, product.Description, product.Price, product.Quantity)
}

func (p *ProductRoutes) GetAllProducts(c *gin.Context) {
	var product entity.Product
	var ProductResponse entity.ExtendedProduct

	rows, err := p.db.Query("SELECT * FROM products")
	if err != nil {
		return
	}
	for rows.Next() {
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
		if err != nil {
			return
		}
		ProductResponse.Products = append(ProductResponse.Products, product)
	}

	ProductResponse.Total = len(ProductResponse.Products)

	c.JSON(http.StatusOK, ProductResponse)
}

func (p *ProductRoutes) GetProductById(c *gin.Context) {
	id := c.Param("id")

	var product entity.Product

	rows, _ := p.db.Query("SELECT * FROM products WHERE id = $1", id)
	_ = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
	for rows.Next() {
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
		if err != nil {
			return
		}
	}
	c.JSON(http.StatusOK, product)
}
