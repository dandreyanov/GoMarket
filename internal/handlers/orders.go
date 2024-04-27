package handlers

import (
	"GoMarket/internal/entity"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type OrderRoutes struct {
	db *sql.DB
}

func NewOrderRoutes(database *sql.DB) *OrderRoutes {
	return &OrderRoutes{
		db: database,
	}
}

func (o *OrderRoutes) MakeOrder(c *gin.Context) {
	var order entity.Order
	var quantity, price uint8
	err := c.Bind(&order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order.ID = uuid.New().String()

	stmt, err := o.db.Query("SELECT quantity, price FROM products WHERE id = $1", order.ProductID)
	if err != nil {
		panic(err)
	}
	defer func(stmt *sql.Rows) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	for stmt.Next() {
		err = stmt.Scan(&quantity, &price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if quantity < order.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough quantity"})
		return
	}

	order.Price = price * order.Quantity

	_, err = o.db.Exec("INSERT INTO orders (id, user_id, product_id, price, quantity, timestamp) VALUES ($1, $2, $3, $4, $5, $6)", order.ID, order.UserID, order.ProductID, order.Price, order.Quantity, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newQuantity := quantity - order.Quantity

	fmt.Println(newQuantity, order.ProductID)
	_, err = o.db.Exec("UPDATE products SET quantity = $1 WHERE id = $2", newQuantity, order.ProductID)
	if err != nil {
		c.JSON(http.StatusMultiStatus, gin.H{"database error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order.ID)
}

func (o *OrderRoutes) GetUserOrders(c *gin.Context) {
	id := c.Param("id")

	var orders entity.Order
	var ordersResponse entity.ExtendedOrder

	rows, _ := o.db.Query("SELECT * FROM orders WHERE user_id = $1", id)
	for rows.Next() {
		err := rows.Scan(&orders.ID, &orders.UserID, &orders.ProductID, &orders.Price, &orders.Quantity, &orders.Timestamp)
		if err != nil {
			return
		}
		ordersResponse.Orders = append(ordersResponse.Orders, orders)
	}
	ordersResponse.Total = len(ordersResponse.Orders)
	c.JSON(http.StatusOK, ordersResponse)
}

func (o *OrderRoutes) GetProductOrders(c *gin.Context) {
	id := c.Param("id")

	var orders entity.Order
	var ordersResponse entity.ExtendedOrder

	rows, _ := o.db.Query("SELECT * FROM orders WHERE product_id = $1", id)
	for rows.Next() {
		err := rows.Scan(&orders.ID, &orders.UserID, &orders.ProductID, &orders.Price, &orders.Quantity, &orders.Timestamp)
		if err != nil {
			return
		}
		ordersResponse.Orders = append(ordersResponse.Orders, orders)
	}
	ordersResponse.Total = len(ordersResponse.Orders)
	c.JSON(http.StatusOK, ordersResponse)
}
