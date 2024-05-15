package handlers

import (
	"GoMarket/internal/entity"
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type UserRoutes struct {
	db *sql.DB
}

func NewUserRoutes(database *sql.DB) *UserRoutes {
	return &UserRoutes{
		db: database,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
}

func (u *UserRoutes) LoginUser(c *gin.Context) {
	var user entity.LoginUser
	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`)
	loginRegex := regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)
	if !loginRegex.MatchString(user.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин должен содержать только латиницу и цифры, от 3 до 20 символов"})
		return
	}
	if !passwordRegex.MatchString(user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен содержать только латиницу и цифры и быть не короче 8 символов"})
		return
	}

	storedPassword := ""
	err = u.db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// После успешной аутентификации генерируем токен
	token, err := GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Аутентификация успешна", "token": token})

}

func (u *UserRoutes) RegisterUser(c *gin.Context) {
	var user entity.RegistrationUser
	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`)
	loginRegex := regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)
	if !loginRegex.MatchString(user.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин должен содержать только латиницу и цифры, от 3 до 20 символов"})
		return
	}
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

	if sqliteErr, ok := err.(sqlite3.Error); ok {
		if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Логин уже занят"})
			return
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Произошла ошибка при попытке зарегистрироваться. Попробуйте позже."})
		return
	}

	c.JSON(http.StatusCreated, user.ID)
}

func GenerateToken(username string) (string, error) {
	// Создаем новый токен
	token := jwt.New(jwt.SigningMethodHS256)

	// Задаем клеймы токена
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Токен действителен в течение 24 часов

	// Генерация подписи для токена
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *UserRoutes) ProtectedEndpoint(c *gin.Context) {
	// Получаем токен из заголовка Authorization
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен авторизации"})
		return
	}

	// Проверяем и валидируем токен
	claims, err := ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
		return
	}

	// Токен валиден, продолжаем выполнение запроса
	// Дополнительные действия могут быть выполнены здесь
	c.Set("userID", claims.UserID)
	c.Next()
}

func ValidateToken(tokenString string) (*Claims, error) {
	// Парсим JWT токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи: %v", token.Header["alg"])
		}
		// Возвращаем секретный ключ для верификации подписи
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		// Возвращаем клеймы токена
		return &Claims{UserID: username}, nil
	} else {
		return nil, errors.New("неверный токен")
	}
}

func (u *UserRoutes) ProtectedProductEndpoint(c *gin.Context) {
	// Вызываем функцию проверки токена
	claims, err := ValidateToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
		return
	}

	// Токен валиден, продолжаем выполнение запроса
	// Дополнительные действия могут быть выполнены здесь
	c.Set("userID", claims.UserID)
	c.Next()
}

func (u *UserRoutes) ProtectedOrderEndpoint(c *gin.Context) {
	// Вызываем функцию проверки токена
	claims, err := ValidateToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
		return
	}

	// Токен валиден, продолжаем выполнение запроса
	// Дополнительные действия могут быть выполнены здесь
	c.Set("userID", claims.UserID)
	c.Next()
}

func (u *UserRoutes) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен авторизации"})
			c.Abort()
			return
		}

		// Проверяем и валидируем токен
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			return
		}

		// Токен валиден, продолжаем выполнение запроса
		// Дополнительные действия могут быть выполнены здесь
		c.Set("userID", claims.UserID)
	}
}
