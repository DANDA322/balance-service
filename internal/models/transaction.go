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
