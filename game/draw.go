package game

import (
	"strconv"
	"time"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/concord-four-in-a-row/engine"
)

// Drawing on the arcade canvas: THE BOARDWALK. Discs are round, drawn in
// fine pixels (quadrant characters) so they stay round at every size; the
// frame is the board's colour and an empty slot is a dark hole. A dropped
// disc falls into place.

// slot is a hole's size in cells, and its disc in fine pixels (two across
// and two down to a cell).
type slot struct {
	cw, ch int
	disc   []string
}

var slots = []slot{
	{8, 4, round(16, 8, 6, 3)},
	{6, 3, round(12, 6, 4, 2)},
	// Small holes are a touch wider than tall, with a row of frame under
	// each, so they don't run together into stripes.
	{5, 2, []string{"..######..", ".########.", "..######..", ".........."}},
	{3, 1, nil}, // the smallest: a ● on the frame

}

// round is a disc rx by ry fine pixels in radius, centred in fw x fh. A
// fine pixel is twice as tall as it is wide, so rx = 2 * ry is round.
func round(fw, fh int, rx, ry float64) []string {
	rows := make([]string, fh)
	for y := 0; y < fh; y++ {
		row := make([]byte, fw)
		for x := 0; x < fw; x++ {
			dx := (float64(x) + .5 - float64(fw)/2) / rx
			dy := (float64(y) + .5 - float64(fh)/2) / ry
			row[x] = '.'
			if dx*dx+dy*dy <= 1.05 {
				row[x] = '#'
			}
		}
		rows[y] = string(row)
	}
	return rows
}

// look is how to draw one board.
type look struct {
	set    *discSet
	frame  *frame
	cursor int // the column being aimed at (a disc shown above it), or -1
	mover  int // whose disc is shown above the cursor
	last   int // the last disc's cell, or -1
	line   map[int]bool
	flash  bool // the winning line lights up now
	fall   *falling
	ghost  bool // the discs are a locked silhouette
	grey   bool // the frame is locked: grey
	labels bool // column numbers under the board
}

// falling is a disc on its way down: into cell, p of the way (0 to 1).
type falling struct {
	cell int
	p    float64
}

func plainLook(s *discSet, f *frame) look {
	return look{set: s, frame: f, cursor: -1, last: -1}
}

// slotFor is the biggest slot that fits the board (and its rows above and
// below) in w x h.
func slotFor(b *engine.Board, w, h int) slot {
	for _, s := range slots {
		if b.W*s.cw+2 <= w && b.H*s.ch+2 <= h {
			return s
		}
	}
	return slots[len(slots)-1]
}

// drawBoard draws position b centred in w x h cells at (x, y): a row for
// the disc being aimed, the board, and the column numbers.
func drawBoard(c *arcade.Canvas, x, y, w, h int, b *engine.Board, l look) {
	s := slotFor(b, w, h)
	bw, bh := b.W*s.cw+2, b.H*s.ch+2
	x += max(0, (w-bw)/2)
	y += max(0, (h-bh)/2)
	top := y + 1 // the board's first row; the aiming row is above it
	frameRole := l.frame.role
	if l.grey {
		frameRole = "line"
	}
	c.Shade(x, top, bw, b.H*s.ch, frameRole)
	for r := 0; r < b.H; r++ {
		for col := 0; col < b.W; col++ {
			cell := r*b.W + col
			role := "bg" // an empty hole
			if p := b.Cells[cell]; p != 0 && (l.fall == nil || l.fall.cell != cell) {
				role = l.set.colour(int(p)-1, l.ghost)
				if l.line[cell] && l.flash {
					role = "fg"
				}
			}
			sx, sy := x+1+col*s.cw, top+(b.H-1-r)*s.ch
			if s.disc == nil {
				c.Text(sx+s.cw/2, sy, "●", role, frameRole, false)
				continue
			}
			c.FineSprite(2*sx, 2*sy, s.disc, map[rune]string{'#': role})
			if cell == l.last && s.ch >= 2 && role != "bg" && l.fall == nil {
				c.FinePx(2*sx+len(s.disc[0])/2-1, 2*sy+len(s.disc)/2-1, frameRole) // a dot on the last disc
			}
		}
	}
	if l.fall != nil && s.disc != nil { // the falling disc, from above the board to its slot
		col, r := l.fall.cell%b.W, l.fall.cell/b.W
		endY := 2 * (top + (b.H-1-r)*s.ch)
		startY := 2 * (top - s.ch)
		fy := startY + int(float64(endY-startY)*min(1, l.fall.p))
		sx := x + 1 + col*s.cw
		role := l.set.colour(int(b.Cells[l.fall.cell])-1, l.ghost)
		for j, row := range s.disc {
			if fy+j >= 2*top { // only inside the board
				c.FineSprite(2*sx, fy+j, []string{row}, map[rune]string{'#': role})
			}
		}
	}
	if l.cursor >= 0 && l.cursor < b.W {
		cx := x + 1 + l.cursor*s.cw + s.cw/2
		c.Text(cx, y, "▼", l.set.colour(l.mover, false), "", true)
	}
	if l.labels {
		for col := 0; col < b.W; col++ {
			c.Text(x+1+col*s.cw+s.cw/2, top+b.H*s.ch, strconv.Itoa(col+1), "dim", "", false)
		}
	}
}

// samplePosition is a game in progress, for previews.
func samplePosition() *engine.Board {
	b := engine.New(7, 6)
	for _, c := range []int{3, 3, 2, 4, 4, 2, 5, 1, 2, 5} {
		_ = b.Drop(c)
	}
	return b
}

// preview draws an unlockable for the arcade: a disc set as a game on its
// own frame, or the two discs side by side when space is short; a frame as
// a game in it. Only the locked thing is greyed out.
func preview(c *arcade.Canvas, id string, sel map[string]string, x, y, w, h int, ghost bool) {
	s, f := set(sel[kindDiscs]), frameFor(sel[kindFrame])
	isFrame := len(id) > len(framePrefix) && id[:len(framePrefix)] == framePrefix
	if isFrame {
		f = frameFor(id)
	} else {
		s = set(id)
	}
	if h >= 8 {
		l := plainLook(s, f)
		l.ghost, l.grey = ghost && !isFrame, ghost && isFrame
		drawBoard(c, x, y, w, h, samplePosition(), l)
		return
	}
	// Two big discs, one per side, set in a piece of the frame.
	sl := slots[1]
	if h < 4 || w < 2*sl.cw+4 {
		sl = slots[2]
	}
	bw, bh := 2*sl.cw+2, sl.ch+1
	x0 := x + max(0, (w-bw)/2)
	y0 := y + max(0, (h-bh)/2)
	role := f.role
	if ghost && isFrame {
		role = "line"
	}
	c.Shade(x0, y0, bw, bh, role)
	for side := 0; side < 2; side++ {
		c.FineSprite(2*(x0+1+side*sl.cw), 2*y0+1, sl.disc, map[rune]string{'#': s.colour(side, ghost && !isFrame)})
	}
}

// setupPreview is the NEW GAME screen's picture: an empty board of the
// chosen size.
func setupPreview(c *arcade.Canvas, options map[string]string, x, y, w, h int) {
	l := plainLook(&discSets[0], &frames[0])
	l.labels = true
	drawBoard(c, x, y, w, h, New(options).B, l)
}

// ── Callouts ───────────────────────────────────────────────────────────────

const (
	calloutLasts = 1400 * time.Millisecond
	fallPerRow   = 70 * time.Millisecond // how fast a dropped disc falls
)

// calloutFor is what to call out after the last move, or "".
func calloutFor(g *Game) string {
	switch {
	case g.Forked:
		return "FORK!"
	case g.Blocked:
		return "BLOCK!"
	}
	return ""
}

// drawCallout writes text in the logo font, centred on (cx, row), on a
// plain backing so it reads over the board.
func drawCallout(c *arcade.Canvas, text string, cx, row int) {
	w := arcade.LogoWidth(text, 1)
	x := cx - w/2
	for j := row - 1; j < row+arcade.LogoHeight(1)+1; j++ { // clear what's under it, pixels too
		c.Fill(x-1, j, w+3, " ", "fg", "bg")
	}
	c.Logo(text, x, 2*row, 1, []string{"yellow", "yellow", "yellow", "orange", "orange", "orange", "orange"})
}

// ── Attract mode ───────────────────────────────────────────────────────────

// The computer plays itself: two seeded opening moves, then the normal
// player (which always picks the same column in the same position), a disc
// every demoStep ticks, a pause on the end, then the next game in the next
// set and frame. Positions are worked out as first needed and kept, so a
// frame is the same picture whoever asks.
const (
	demoStep  = 3
	demoPause = 12
)

var demo struct {
	n     int
	games []*engine.Board
	done  bool
}

func demoGame(n, ply int) ([]*engine.Board, bool) {
	if demo.games == nil || demo.n != n {
		demo.n, demo.games, demo.done = n, []*engine.Board{engine.New(7, 6)}, false
	}
	for !demo.done && len(demo.games) <= ply {
		b := demo.games[len(demo.games)-1].Clone()
		col := engine.Best(b, 2)
		if b.Played < 2 {
			col = (n*3 + b.Played*5) % 7
		}
		_ = b.Drop(col)
		demo.games = append(demo.games, b)
		demo.done = b.Over()
	}
	return demo.games, demo.done
}

func attract(c *arcade.Canvas, x, y, w, h, frame int) {
	step := frame / demoStep
	n := int(time.Now().Unix() / 3600)
	games, done := demoGame(n, step)
	for done && step >= len(games)+demoPause && n < 1<<30 {
		step -= len(games) + demoPause
		n++
		games, done = demoGame(n, step)
	}
	b := games[min(step, len(games)-1)]
	s := &discSets[n%len(discSets)]
	f := &frames[n%len(frames)]
	l := plainLook(s, f)
	l.last = b.Last
	if b.Winner() >= 0 {
		l.line = map[int]bool{}
		for _, cell := range b.LineAt(b.Last) {
			l.line[cell] = true
		}
		l.flash = frame%4 < 2
	}
	bw := 44
	drawBoard(c, x, y, bw, h, b, l)
	px := x + bw + 2
	c.Text(px, y+1, "NOW SHOWING", "pink", "", true)
	c.Text(px, y+3, s.Name, s.Tier.Role(), "", true)
	c.Text(px, y+4, "in "+f.Name, "dim", "", false)
	if s.Tier != arcade.Starter {
		c.Text(px, y+6, s.Tier.Stars()+" "+s.Tier.Name(), s.Tier.Role(), "", false)
	}
	c.Text(px, y+8, "DISC "+strconv.Itoa(b.Played), "dim", "", false)
	switch {
	case b.Winner() >= 0:
		c.Text(px, y+9, "★ FOUR!", "yellow", "", true)
	case b.Full():
		c.Text(px, y+9, "★ FULL!", "yellow", "", true)
	}
}
