package models

import "time"

type TransactionsQueryParams struct {
	From       time.Time
	To         time.Time
	Limit      int
	Offset     int
	Sorting    string
	Descending string
}
