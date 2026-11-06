// Package domain - содержит описание доменных сущностей: игровой доски, кампании, кэшбека, состояния партии
package domain

import (
	"cashbackmerge/internal/core/apperr"
	"math/bits"
	"strconv"
)

// Cashback - уровень кэшбэка плитки в процентах
type Cashback struct {
	percent uint
}

// NewCashback создаёт значение кэшбэка
func NewCashback(percent uint) (Cashback, error) {
	// считает, сколько единиц есть в двоичной записи числа
	if bits.OnesCount(percent) != 1 {
		return Cashback{}, &apperr.ErrInvalidCashback[uint]{
			Input: percent,
		}
	}

	return Cashback{percent: percent}, nil
}

// ParseCashback парсит строку с целым числом процентов: "16" в 16%
func ParseCashback(s string) (Cashback, error) {
	val, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return Cashback{}, &apperr.ErrInvalidCashback[string]{Input: s}
	}

	cashback, err := NewCashback(uint(val))
	if err != nil {
		return Cashback{}, &apperr.ErrInvalidCashback[string]{Input: s}
	}

	return cashback, nil
}

// String возвращает значение в виде "16%"
func (c Cashback) String() string {
	return strconv.FormatUint(uint64(c.percent), 10) + "%"
}
