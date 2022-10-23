package csv

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"unicode/utf8"
)

type WriterCSV struct {
}

func (c *WriterCSV) WriteReport(w http.ResponseWriter, data map[string]float64) error {
	writer := csv.NewWriter(w)
	delimiter, _ := utf8.DecodeRuneInString(";")
	writer.Comma = delimiter

	row := []string{"ServiceTitle", "Amount"}
	err := writer.Write(row)
	if err != nil {
		return fmt.Errorf("cannot write to CSV file: %w", err)
	}

	for key, value := range data {
		row[0] = key
		row[1] = fmt.Sprintf("%v", value)
		err = writer.Write(row)
		if err != nil {
			return fmt.Errorf("cannot write to CSV file: %w", err)
		}
	}
	writer.Flush()

	return nil
}
