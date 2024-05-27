package handlers

import (
	"GoMarket/internal/entity"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"sync"
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

	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	defer close(errChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		stmt, err := o.db.Query("SELECT quantity, price FROM products WHERE id = $1", order.ProductID)
		if err != nil {
			errChan <- err
			return
		}
		defer stmt.Close()

		if stmt.Next() {
			err = stmt.Scan(&quantity, &price)
			if err != nil {
				errChan <- err
				return
			}
		} else {
			errChan <- sql.ErrNoRows
			return
		}
	}()

	wg.Wait()
	select {
	case err := <-errChan:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
	}

	if quantity < order.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нет нужного количества товара"})
		return
	}

	order.Price = price * order.Quantity

	// Горутин для вставки нового заказа
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = o.db.Exec("INSERT INTO orders (id, user_id, product_id, price, quantity, timestamp) VALUES ($1, $2, $3, $4, $5, $6)", order.ID, order.UserID, order.ProductID, order.Price, order.Quantity, time.Now())
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()
	select {
	case err := <-errChan:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
	}

	newQuantity := quantity - order.Quantity

	// Горутин для обновления количества продукта
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = o.db.Exec("UPDATE products SET quantity = $1 WHERE id = $2", newQuantity, order.ProductID)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()
	select {
	case err := <-errChan:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
	}

	c.JSON(http.StatusCreated, order.ID)
}

func (o *OrderRoutes) GetUserOrders(c *gin.Context) {
	id := c.Param("id")

	var orders entity.Order
	var ordersResponse entity.ExtendedOrder

	rows, err := o.db.Query("SELECT * FROM orders WHERE user_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&orders.ID, &orders.UserID, &orders.ProductID, &orders.Price, &orders.Quantity, &orders.Timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	rows, err := o.db.Query("SELECT * FROM orders WHERE product_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&orders.ID, &orders.UserID, &orders.ProductID, &orders.Price, &orders.Quantity, &orders.Timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ordersResponse.Orders = append(ordersResponse.Orders, orders)
	}
	ordersResponse.Total = len(ordersResponse.Orders)
	c.JSON(http.StatusOK, ordersResponse)
}
