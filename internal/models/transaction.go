package models

type Transaction struct {
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment"`
}

type TransferTransaction struct {
	Target  int     `json:"target"`
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment"`
}

type ReserveTransaction struct {
	ServiceID int     `json:"service_id"`
	OrderID   int     `json:"order_id"`
	Amount    float64 `json:"amount"`
}
