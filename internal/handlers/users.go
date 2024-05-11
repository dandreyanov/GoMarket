package handlers

import (
	"GoMarket/internal/entity"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
)

type UserRoutes struct {
	db *sql.DB
}

func NewUserRoutes(database *sql.DB) *UserRoutes {
	return &UserRoutes{
		db: database,
	}
}

func (u *UserRoutes) RegisterUser(c *gin.Context) {
	var user entity.User
	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`)
	if !passwordRegex.MatchString(user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен содержать только латиницу и цифры и быть не короче 8 символов"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = strings.TrimRight(string(hashedPassword), "\n")

	user.ID = uuid.New().String()
	_, err = u.db.Exec("INSERT INTO users (id, username, password, email) VALUES ($1, $2, $3, $4)", user.ID, user.Username, user.Password, user.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user.ID)
}
