package entity

type Product struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty" binding:"required"`
	Description string `json:"description,omitempty"`
	Price       uint8  `json:"price,omitempty"`
	Quantity    uint8  `json:"quantity,omitempty"`
}

type ExtendedProduct struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
}

type Order struct {
	ID        string `json:"id,omitempty"`
	UserID    string `json:"userId,omitempty"`
	ProductID string `json:"productId,omitempty"`
	Price     uint8  `json:"price,omitempty"`
	Quantity  uint8  `json:"quantity,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type ExtendedOrder struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

type RegistrationUser struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
}

type LoginUser struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
