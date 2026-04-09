package domain

import (
	"fmt"
	"strings"
)

func requireNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewDomainError(ErrInvalidArgument, fmt.Sprintf("%s is required", field))
	}
	return nil
}

func requireMaxRunes(field, value string, max int) error {
	if len([]rune(strings.TrimSpace(value))) > max {
		return NewDomainError(ErrInvalidArgument, fmt.Sprintf("%s exceeds %d characters", field, max))
	}
	return nil
}
