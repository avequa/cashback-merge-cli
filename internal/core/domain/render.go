// Package domain - содержит описание доменных сущностей: игровой доски, кампании, кэшбека, состояния партии
package domain

import (
	"fmt"
	"strings"
)

// String возвращает поле в текстовом виде: size строк по size клеток
func (b Board) String() string {
	var sb strings.Builder

	for r := range b.size {
		for c := range b.size {
			cell := "."
			if v := b.cells[r][c]; v != (Cashback{}) {
				cell = v.String()
			}

			fmt.Fprintf(&sb, "%6s", cell)
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
