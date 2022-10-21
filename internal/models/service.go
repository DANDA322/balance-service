package models

type Service struct {
	Title  string  `db:"title"`
	Amount float64 `db:"amount"`
}
