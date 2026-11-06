// Package domain - содержит описание доменных сущностей: игровой доски, кампании, кэшбека, состояния партии
package domain

// State - состояние партии
type State struct {
	Board      Board
	MovesLimit int
	MovesUsed  int
	Score      int
	Finished   bool
}
