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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product.ID)
}

func (p *ProductRoutes) GetAllProducts(c *gin.Context) {
	var product entity.Product
	var ProductResponse entity.ExtendedProduct

	rows, err := p.db.Query("SELECT * FROM products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	row := p.db.QueryRow("SELECT * FROM products WHERE id = $1", id)
	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
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

	row := p.db.QueryRow("UPDATE products SET name = $1, description = $2, price = $3, quantity = $4 WHERE id = $5 RETURNING id, name, description, price, quantity", product.Name, product.Description, product.Price, product.Quantity, id)
	err = row.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, product)
}

func (p *ProductRoutes) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	result, err := p.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар удален"})
}
