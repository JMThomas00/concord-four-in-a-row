// Package engine is Four in a Row: drop discs into columns, and four of
// yours in a line (across, up, or diagonally) wins. Boards come in a few
// sizes; the classic one is 7 columns by 6 rows.
package engine

import "fmt"

// Board is a position. Cells are row-major from the bottom row up: cell
// r*W+c is column c, row r (row 0 at the bottom). A cell holds 0 for empty,
// 1 for the first player's disc and 2 for the second's.
type Board struct {
	W, H    int
	Cells   []int8
	Heights []int // discs in each column
	ToMove  int   // 0 or 1
	Played  int   // discs dropped
	Last    int   // the cell of the last disc, or -1
}

// New is an empty board.
func New(w, h int) *Board {
	return &Board{W: w, H: h, Cells: make([]int8, w*h), Heights: make([]int, w), Last: -1}
}

// Clone copies the board.
func (b *Board) Clone() *Board {
	c := *b
	c.Cells = append([]int8(nil), b.Cells...)
	c.Heights = append([]int(nil), b.Heights...)
	return &c
}

// At is the disc in column c, row r (0 empty, 1 or 2), or -1 off the board.
func (b *Board) At(c, r int) int8 {
	if c < 0 || c >= b.W || r < 0 || r >= b.H {
		return -1
	}
	return b.Cells[r*b.W+c]
}

// Open reports whether column c has room.
func (b *Board) Open(c int) bool { return c >= 0 && c < b.W && b.Heights[c] < b.H }

// Drop plays the side to move's disc into column c.
func (b *Board) Drop(c int) error {
	if c < 0 || c >= b.W {
		return fmt.Errorf("there's no column %d", c+1)
	}
	if !b.Open(c) {
		return fmt.Errorf("column %d is full", c+1)
	}
	if b.Over() {
		return fmt.Errorf("the game is over")
	}
	b.drop(c)
	return nil
}

func (b *Board) drop(c int) {
	cell := b.Heights[c]*b.W + c
	b.Cells[cell] = int8(b.ToMove + 1)
	b.Heights[c]++
	b.ToMove = 1 - b.ToMove
	b.Played++
	b.Last = cell
}

// undo takes the top disc out of column c (search only: Last isn't kept).
func (b *Board) undo(c int) {
	b.Heights[c]--
	b.Cells[b.Heights[c]*b.W+c] = 0
	b.ToMove = 1 - b.ToMove
	b.Played--
}

// directions to look for a line: across, up, and both diagonals.
var dirs = [4][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}

// LineAt is the line of four or more through cell, or nil.
func (b *Board) LineAt(cell int) []int {
	if cell < 0 || b.Cells[cell] == 0 {
		return nil
	}
	p, c0, r0 := b.Cells[cell], cell%b.W, cell/b.W
	for _, d := range dirs {
		line := []int{cell}
		for _, s := range []int{1, -1} {
			for c, r := c0+s*d[0], r0+s*d[1]; b.At(c, r) == p; c, r = c+s*d[0], r+s*d[1] {
				line = append(line, r*b.W+c)
			}
		}
		if len(line) >= 4 {
			return line
		}
	}
	return nil
}

// Winner is the player with four in a row (0 or 1), or -1.
func (b *Board) Winner() int {
	if b.Last < 0 {
		return -1
	}
	if b.LineAt(b.Last) != nil {
		return int(b.Cells[b.Last]) - 1
	}
	return -1
}

// Full reports whether every column is full (a draw, unless that disc won).
func (b *Board) Full() bool { return b.Played == b.W*b.H }

// Over reports whether the game has ended.
func (b *Board) Over() bool { return b.Winner() >= 0 || b.Full() }

// Wins lists the columns where player would make four with their next disc.
func (b *Board) Wins(player int) []int {
	var out []int
	for c := 0; c < b.W; c++ {
		if !b.Open(c) {
			continue
		}
		cell := b.Heights[c]*b.W + c
		b.Cells[cell] = int8(player + 1)
		if b.LineAt(cell) != nil {
			out = append(out, c)
		}
		b.Cells[cell] = 0
	}
	return out
}

// Diagonal reports whether line runs diagonally.
func (b *Board) Diagonal(line []int) bool {
	if len(line) < 2 {
		return false
	}
	dc := line[1]%b.W - line[0]%b.W
	dr := line[1]/b.W - line[0]/b.W
	return dc != 0 && dr != 0
}
