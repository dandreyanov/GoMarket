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

	// Определение канала для получения результата выполнения запроса к базе данных
	resultChan := make(chan string, 1)

	// Запуск горутины для выполнения запроса к базе данных
	go func(username string) {
		var storedPassword string
		err := u.db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				resultChan <- "Неверный логин или пароль"
			} else {
				resultChan <- "Ошибка при запросе в базу данных"
			}
			return
		}
		resultChan <- storedPassword
	}(user.Username)

	// Ожидание результата выполнения запроса к базе данных
	storedPassword := <-resultChan

	// Проверка результата выполнения запроса
	if storedPassword == "Неверный логин или пароль" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": storedPassword})
		return
	} else if storedPassword == "Ошибка при запросе в базу данных" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": storedPassword})
		return
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// После успешной аутентификации генерируем токен
	token, err := GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при генерации токена"})
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
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !loginRegex.MatchString(user.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин должен содержать только латиницу и цифры, от 3 до 20 символов"})
		return
	}
	if !passwordRegex.MatchString(user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен содержать только латиницу и цифры и быть не короче 8 символов"})
		return
	}
	if !emailRegex.MatchString(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Введен невалидный email"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при хешировании пароля"})
		return
	}
	user.Password = strings.TrimRight(string(hashedPassword), "\n")
	user.ID = uuid.New().String()

	// Определение канала для получения результата выполнения запроса к базе данных
	resultChan := make(chan error, 1)

	// Запуск горутины для выполнения запроса к базе данных
	go func() {
		_, err := u.db.Exec("INSERT INTO users (id, username, password, email) VALUES ($1, $2, $3, $4)", user.ID, user.Username, user.Password, user.Email)
		resultChan <- err
	}()

	// Ожидание результата выполнения запроса к базе данных
	err = <-resultChan

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

func (u *UserRoutes) ProtectedOrderEndpoint(c *gin.Context) {
	// Вызываем функцию проверки токена в горутине
	go func(token string) {
		claims, err := ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			return
		}

		// Токен валиден, продолжаем выполнение запроса
		// Дополнительные действия могут быть выполнены здесь
		c.Set("userID", claims.UserID)
		c.Next()
	}(c.GetHeader("Authorization"))
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

		// Вызываем функцию проверки токена в горутине
		go func(token string) {
			claims, err := ValidateToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
				c.Abort()
				return
			}

			// Токен валиден, продолжаем выполнение запроса
			// Дополнительные действия могут быть выполнены здесь
			c.Set("userID", claims.UserID)
		}(tokenString)
	}
}

func ValidateToken(tokenString string) (*Claims, error) {
	// Создаем новый токен
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
