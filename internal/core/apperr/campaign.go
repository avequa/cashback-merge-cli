// Package apperr содержит типизированные ошибки домена
package apperr

import "fmt"

// ErrInvalidCampaign - нарушено правило конфигурации кампании
type ErrInvalidCampaign struct {
	Field  string // поле конфигурации
	Reason string // что не так
}

func (e *ErrInvalidCampaign) Error() string {
	return fmt.Sprintf("invalid campaign: %s %s", e.Field, e.Reason)
}
