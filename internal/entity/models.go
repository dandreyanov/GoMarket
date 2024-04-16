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
