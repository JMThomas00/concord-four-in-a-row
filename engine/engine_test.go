package engine

import (
	"testing"
	"time"
)

func play(t *testing.T, b *Board, cols ...int) {
	t.Helper()
	for _, c := range cols {
		if err := b.Drop(c); err != nil {
			t.Fatalf("column %d: %v", c+1, err)
		}
	}
}

func TestLines(t *testing.T) {
	for name, cols := range map[string][]int{
		"across":   {0, 0, 1, 1, 2, 2, 3},
		"up":       {0, 1, 0, 1, 0, 1, 0},
		"diagonal": {0, 1, 1, 2, 2, 3, 2, 3, 3, 6, 3},
		"anti":     {3, 2, 2, 1, 1, 0, 1, 0, 0, 6, 0},
	} {
		b := New(7, 6)
		play(t, b, cols...)
		if b.Winner() != 0 || !b.Over() {
			t.Errorf("%s: winner %d", name, b.Winner())
		}
		line := b.LineAt(b.Last)
		if len(line) < 4 {
			t.Errorf("%s: line %v", name, line)
		}
		if diag := name == "diagonal" || name == "anti"; b.Diagonal(line) != diag {
			t.Errorf("%s: diagonal %v", name, b.Diagonal(line))
		}
		if err := b.Drop(5); err == nil {
			t.Errorf("%s: played on after the end", name)
		}
	}
}

func TestFullColumnAndDraw(t *testing.T) {
	b := New(7, 6)
	play(t, b, 0, 0, 0, 0, 0, 0)
	if err := b.Drop(0); err == nil {
		t.Fatal("a seventh disc went into a full column")
	}
	// A full board with no four: rows ABABABA, ABABABA, BABABAB, BABABAB, ...
	b = New(7, 6)
	for r := 0; r < 6; r++ {
		for c := 0; c < 7; c++ {
			b.Cells[r*7+c] = int8(1 + (c+r/2)%2)
		}
	}
	for c := range b.Heights {
		b.Heights[c] = 6
	}
	b.Played, b.Last = 42, 41
	for cell := range b.Cells {
		if line := b.LineAt(cell); line != nil {
			t.Fatalf("the pattern has a four: %v", line)
		}
	}
	if !b.Full() || b.Winner() >= 0 {
		t.Fatalf("full %v, winner %d", b.Full(), b.Winner())
	}
}

func TestComputerWinsAndBlocks(t *testing.T) {
	for level := 1; level <= 3; level++ {
		b := New(7, 6)
		play(t, b, 0, 6, 1, 6, 2) // first player has three across the bottom
		if c := Best(b, level); level >= 2 && c != 3 {
			t.Errorf("level %d didn't block at column 4: %d", level, c+1)
		}
		b = New(7, 6)
		play(t, b, 0, 6, 1, 6, 2, 6) // ...and it's their move: win
		if c := Best(b, level); c != 3 {
			t.Errorf("level %d missed the win: %d", level, c+1)
		}
	}
}

func TestComputerIsQuickEnough(t *testing.T) {
	for _, size := range [][2]int{{7, 6}, {8, 7}, {9, 7}} {
		b := New(size[0], size[1])
		began := time.Now()
		for i := 0; i < 6 && !b.Over(); i++ {
			play(t, b, Best(b, 3))
		}
		if d := time.Since(began); d > 15*time.Second {
			t.Errorf("%dx%d: six hard moves took %v", size[0], size[1], d)
		}
	}
}
