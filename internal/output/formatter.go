package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

type CSVRecord interface {
	CSVHeaders() []string
	CSVRow() []string
}

func FormatJSON[T any](w io.Writer, records []T) error {
	if records == nil {
		records = []T{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

func FormatCSV[T CSVRecord](w io.Writer, records []T) error {
	if len(records) == 0 {
		return nil
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(records[0].CSVHeaders()); err != nil {
		return err
	}
	for _, r := range records {
		if err := cw.Write(r.CSVRow()); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func Write[T CSVRecord](w io.Writer, format string, records []T) error {
	switch format {
	case "json":
		return FormatJSON(w, records)
	case "csv":
		return FormatCSV(w, records)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func intToStr(n int) string {
	return strconv.Itoa(n)
}
