package main

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price,omitempty"`
	Category *Category `json:"category"`
}

func GetSubtotal(products []Product) float64 {
	total := 0.0
	for _, product := range products {
		total += product.Price
	}
	return total
}
