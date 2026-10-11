// Package game plugs the Four in a Row engine into the Concord table kit:
// the rules adapter, the computer player, and the board people play on.
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/concord-four-in-a-row/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// OptionSize is the NEW GAME screen's board: "7x6" (the classic), "8x7" or
// "9x7".
const OptionSize = "size"

// Rules is Four in a Row for the table kit.
var Rules = table.Rules{
	Name:      "Four in a Row",
	SeatNames: []string{"First", "Second"},
	New:       func(options map[string]string) table.Game { return New(options) },
	NewBoard:  func(s *table.Seat) tea.Model { return newBoard(s) },
	AI: func(g table.Game, level int) string {
		return strconv.Itoa(engine.Best(g.(*Game).B, level) + 1)
	},
	Sound:  sound,
	Awards: awards,
	Arcade: arcadeLook,
}

// size reads OptionSize (the classic 7x6 if it's missing or odd).
func size(options map[string]string) (w, h int) {
	switch options[OptionSize] {
	case "8x7":
		return 8, 7
	case "9x7":
		return 9, 7
	}
	return 7, 6
}

// Game adapts an engine.Board to table.Game. A move is a column number,
// "1" on the left.
type Game struct {
	B *engine.Board
	// The last move: its column (-1 before the first), and whether it
	// blocked a four or set up two at once.
	LastCol         int
	Blocked, Forked bool
	// Each side's blocks and forks so far, for achievements. Saved games
	// are replays, so these are worked out again on loading.
	Blocks, Forks [2]int
}

func New(options map[string]string) *Game {
	return &Game{B: engine.New(size(options)), LastCol: -1}
}

func (g *Game) Turn() int {
	if g.B.Over() {
		return -1
	}
	return g.B.ToMove
}

func (g *Game) Play(move string) error {
	col, err := strconv.Atoi(strings.TrimSpace(move))
	if err != nil {
		return fmt.Errorf("a move is a column number, 1 to %d", g.B.W)
	}
	col--
	mover := g.B.ToMove
	blocks := contains(g.B.Wins(1-mover), col)
	if err := g.B.Drop(col); err != nil {
		return err
	}
	g.LastCol = col
	g.Blocked = blocks && !g.B.Over()
	g.Forked = !g.B.Over() && len(g.B.Wins(mover)) >= 2
	if g.Blocked {
		g.Blocks[mover]++
	}
	if g.Forked {
		g.Forks[mover]++
	}
	return nil
}

func contains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func (g *Game) Outcome() table.Outcome {
	if w := g.B.Winner(); w >= 0 {
		return table.Outcome{Over: true, Winner: w, Reason: "four in a row"}
	}
	if g.B.Full() {
		return table.Outcome{Over: true, Winner: -1, Reason: "the board is full"}
	}
	return table.Outcome{}
}

// Line is the winning four (or more), or nil.
func (g *Game) Line() []int {
	if g.B.Winner() < 0 {
		return nil
	}
	return g.B.LineAt(g.B.Last)
}

// sound is what everyone watching hears after a move.
func sound(tg table.Game, move string) string {
	g := tg.(*Game)
	switch o := g.Outcome(); {
	case o.Over && o.Winner >= 0:
		return arcade.SoundFinish
	case o.Over:
		return "sounds/draw.wav"
	case g.Blocked || g.Forked:
		return "sounds/block.wav"
	}
	return "sounds/drop.wav"
}

// awards are the game's own achievements (declared in plugin.toml), beyond
// the table kit's standard ones.
func awards(tg table.Game, seat int, o table.Outcome) []string {
	g := tg.(*Game)
	if !o.Over || o.Winner != seat {
		return nil
	}
	var out []string
	if g.B.Diagonal(g.Line()) {
		out = append(out, "diagonal")
	}
	if (g.B.Played+1)/2 == 4 {
		out = append(out, "four_straight") // won with your first four discs
	}
	if g.B.Full() {
		out = append(out, "last_disc")
	}
	if g.Forks[seat] > 0 {
		out = append(out, "fork")
	}
	if g.Blocks[seat] >= 3 {
		out = append(out, "sea_wall")
	}
	return out
}
