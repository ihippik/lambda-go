package builder

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

type (
	ErrorLine struct {
		Error       string      `json:"error"`
		ErrorDetail ErrorDetail `json:"errorDetail"`
	}

	ErrorDetail struct {
		Message string `json:"message"`
	}
)

func checkErr(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	var errLine ErrorLine

	if err := json.Unmarshal([]byte(lastLine), &errLine); err != nil {
		return err
	}

	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
