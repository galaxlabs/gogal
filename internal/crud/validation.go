package crud

import "fmt"

func ValidatePagination(limit int, offset int) error {
	if limit < 0 || offset < 0 {
		return fmt.Errorf("limit and offset must be non-negative")
	}
	if limit > 500 {
		return fmt.Errorf("limit is too high")
	}
	return nil
}
