package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/Concord/sdk/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Board is what a player sees and plays on: aim with the arrows, drop with
// Enter (or a column's number).
type Board struct {
	seat          *table.Seat
	col           int // the column being aimed at
	width, height int
	err           string

	dropAt    time.Time // when the last disc was dropped, for its fall
	callout   string    // BLOCK!, FORK!, for a moment after a move
	calloutAt time.Time
}

func newBoard(s *table.Seat) *Board {
	return &Board{seat: s, col: s.Game().(*Game).B.W / 2}
}

func (b *Board) game() *Game   { return b.seat.Game().(*Game) }
func (b *Board) Init() tea.Cmd { return nil }

func (b *Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = msg.Width, msg.Height
	case table.ChangedMsg:
		b.err = ""
		g := b.game()
		b.col = min(b.col, g.B.W-1)
		if msg.Move != "" {
			b.dropAt = time.Now()
			b.callout, b.calloutAt = calloutFor(g), time.Now()
		}
	case tea.KeyMsg:
		switch key := msg.String(); key {
		case "left", "h":
			b.col = max(0, b.col-1)
		case "right", "l":
			b.col = min(b.game().B.W-1, b.col+1)
		case "enter", " ", "down", "j":
			b.drop(b.col)
		case "q":
			return b, tea.Quit
		default:
			if n, err := strconv.Atoi(key); err == nil && n >= 1 && n <= b.game().B.W {
				b.col = n - 1
				b.drop(b.col)
			}
		}
	}
	return b, nil
}

func (b *Board) drop(col int) {
	b.err = ""
	if !b.seat.MyTurn() {
		b.err = "not your turn"
		return
	}
	if err := b.seat.Play(strconv.Itoa(col + 1)); err != nil {
		b.err = err.Error()
	}
}

// ── The arcade (table.ArcadeBoard and friends) ─────────────────────────────

func (b *Board) look() look {
	g := b.game()
	l := plainLook(set(b.seat.Equipped(kindDiscs)), frameFor(b.seat.Equipped(kindFrame)))
	l.labels, l.last = true, g.B.Last
	if b.seat.MyTurn() {
		l.cursor, l.mover = b.col, g.B.ToMove
	}
	if line := g.Line(); line != nil {
		l.line = map[int]bool{}
		for _, cell := range line {
			l.line[cell] = true
		}
		l.flash = b.seat.Frame()%4 >= 2
	}
	if b.falling() && g.B.Last >= 0 {
		rows := g.B.H - g.B.Last/g.B.W
		l.fall = &falling{cell: g.B.Last, p: float64(time.Since(b.dropAt)) / float64(time.Duration(rows)*fallPerRow)}
	}
	return l
}

// falling reports whether the last disc is still on its way down.
func (b *Board) falling() bool {
	g := b.game()
	if b.dropAt.IsZero() || g.B.Last < 0 || b.seat.Effects() != "" {
		return false
	}
	rows := g.B.H - g.B.Last/g.B.W
	return time.Since(b.dropAt) < time.Duration(rows)*fallPerRow
}

// Draw draws the board as big as the space allows (table.ArcadeBoard).
func (b *Board) Draw(c *arcade.Canvas, x, y, w, h int) {
	drawBoard(c, x, y, w, h, b.game().B, b.look())
	if b.callout != "" && b.seat.Effects() == "" && time.Since(b.calloutAt) < calloutLasts && !b.falling() {
		drawCallout(c, b.callout, x+w/2, y+h/2-2)
	}
}

// DrawSeat draws a side's disc, big.
func (b *Board) DrawSeat(c *arcade.Canvas, seat, x, y, w, h int) {
	s := set(b.seat.Equipped(kindDiscs))
	sl := slots[1]
	c.FineSprite(2*(x+(w-sl.cw)/2), 2*y, sl.disc, map[rune]string{'#': s.colour(seat, false)})
}

// DrawPanel shows who could win next move, and where (table.Panel).
func (b *Board) DrawPanel(c *arcade.Canvas, x, y, w, h int) {
	g := b.game()
	s := set(b.seat.Equipped(kindDiscs))
	if g.B.Over() {
		c.Text(x, y, "GAME OVER", "pink", "", true)
		if w := g.B.Winner(); w >= 0 {
			c.Text(x, y+1, strings.ToUpper(b.seat.SeatName(w))+" GOT FOUR", s.colour(w, false), "", true)
		} else {
			c.Text(x, y+1, "THE BOARD IS FULL", "dim", "", false)
		}
		return
	}
	c.Text(x, y, "WINS NEXT", "pink", "", true)
	for side := 0; side < 2; side++ {
		name := strings.ToUpper(b.seat.SeatName(side))
		cols := g.B.Wins(side)
		line := "NONE"
		if len(cols) > 0 {
			var nums []string
			for _, col := range cols {
				nums = append(nums, strconv.Itoa(col+1))
			}
			line = "COL " + strings.Join(nums, " ")
		}
		c.Text(x, y+1+side, fmt.Sprintf("%-6s %s", name, line), s.colour(side, false), "", len(cols) > 0)
	}
	if h >= 6 {
		left := g.B.W*g.B.H - g.B.Played
		c.Text(x, y+4, fmt.Sprintf("%d SLOTS LEFT", left), "dim", "", false)
	}
}

// Status is what went wrong, for the status row (in red).
func (b *Board) Status() string { return b.err }

// Hint is what to do next, for the status row (table.Hinter).
func (b *Board) Hint() string {
	if !b.seat.MyTurn() {
		return ""
	}
	if len(b.game().B.Wins(b.game().B.ToMove)) > 0 {
		return "you can win this move!"
	}
	return fmt.Sprintf("←→ aim · Enter drops · or 1-%d", b.game().B.W)
}

// Animating is true while a disc falls or a callout shows (table.Animator).
func (b *Board) Animating() bool {
	if b.seat.Effects() != "" {
		return false
	}
	return b.falling() || (b.callout != "" && time.Since(b.calloutAt) < calloutLasts)
}

// ── The plain layout (small panes, standalone) ─────────────────────────────

func (b *Board) View() string {
	g := b.game()
	first := lipgloss.NewStyle().Foreground(b.seat.Color("red", lipgloss.Color("9"))).Bold(true)
	second := lipgloss.NewStyle().Foreground(b.seat.Color("yellow", lipgloss.Color("11"))).Bold(true)
	dim := lipgloss.NewStyle().Foreground(b.seat.Color("comment", lipgloss.Color("8")))
	line := map[int]bool{}
	for _, cell := range g.Line() {
		line[cell] = true
	}
	var out []string
	aim := strings.Repeat("  ", b.col) + " ▼"
	if !b.seat.MyTurn() {
		aim = ""
	}
	out = append(out, aim)
	for r := g.B.H - 1; r >= 0; r-- {
		var row strings.Builder
		row.WriteString(dim.Render("│"))
		for c := 0; c < g.B.W; c++ {
			cell := r*g.B.W + c
			switch g.B.Cells[cell] {
			case 1:
				row.WriteString(" " + first.Underline(line[cell]).Render("●"))
			case 2:
				row.WriteString(" " + second.Underline(line[cell]).Render("●"))
			default:
				row.WriteString(dim.Render(" ·"))
			}
		}
		row.WriteString(dim.Render(" │"))
		out = append(out, row.String())
	}
	nums := " "
	for c := 1; c <= g.B.W; c++ {
		nums += " " + strconv.Itoa(c)
	}
	out = append(out, dim.Render(nums))
	status := b.err
	if status == "" && b.seat.MyTurn() {
		status = fmt.Sprintf("your move: ←→ and Enter, or 1-%d", g.B.W)
	}
	out = append(out, status)
	return strings.Join(out, "\n")
}
