package handlers

import (
	"GoMarket/internal/entity"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	c.JSON(http.StatusCreated, product.ID)
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

func (p *ProductRoutes) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var product entity.Product

	err := c.BindJSON(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row, errDb := p.db.Query("UPDATE products SET name = $1, description = $2, price = $3, quantity = $4 WHERE id = $5 RETURNING *", product.Name, product.Description, product.Price, product.Quantity, id)
	if errDb != nil {
		c.JSON(http.StatusMultiStatus, gin.H{"database error": errDb.Error()})
		return
	}

	for row.Next() {
		err = row.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
		if err != nil {
			c.JSON(http.StatusMultiStatus, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, product)
}

func (p *ProductRoutes) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	_, err := p.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusMultiStatus, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "product delete"})
	}
}
