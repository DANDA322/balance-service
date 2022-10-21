package csv

import (
	"encoding/csv"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"unicode/utf8"
)

type WriterCSV struct {
}

func (c *WriterCSV) GetReport(data map[string]float64) (*os.File, error) {
	file, err := os.Create("report.csv")
	if err != nil {
		return nil, fmt.Errorf("cannot open CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Error("err closing file: %v", err)
		}
	}()

	writer := csv.NewWriter(file)
	delimiter, _ := utf8.DecodeRuneInString(";")
	writer.Comma = delimiter

	row := []string{"ServiceTitle", "Amount"}
	err = writer.Write(row)
	if err != nil {
		return nil, fmt.Errorf("cannot write to CSV file: %w", err)
	}

	for key, value := range data {
		row[0] = key
		row[1] = fmt.Sprintf("%v", value)
		err = writer.Write(row)
		if err != nil {
			return nil, fmt.Errorf("cannot write to CSV file: %w", err)
		}
	}
	writer.Flush()

	return file, nil
}
