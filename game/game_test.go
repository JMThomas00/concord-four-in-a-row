package game

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/Concord/sdk/plugin"
	"github.com/JMThomas00/Concord/sdk/plugintest"
	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/Concord/sdk/wire"
	"github.com/google/uuid"
)

func startServer(t *testing.T) (*plugintest.Server, uuid.UUID) {
	t.Helper()
	srv := plugintest.NewServer(t)
	cfg := srv.Config() // before the cleanup below, so the plugin stops before its data folder goes
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { plugin.Run(ctx, cfg, table.New(Rules).Handler()); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	srv.WaitReady()
	ch := uuid.New()
	srv.Channel(wire.Channel{ID: ch, Name: "four", PluginConfig: map[string]string{table.SettingSeating: table.ModeSeats}})
	return srv, ch
}

func play(t *testing.T, g *Game, moves ...string) {
	t.Helper()
	for _, m := range moves {
		if err := g.Play(m); err != nil {
			t.Fatalf("%s: %v", m, err)
		}
	}
}

func TestGameAdapter(t *testing.T) {
	g := New(nil)
	if g.B.W != 7 || g.B.H != 6 {
		t.Fatalf("default board %dx%d", g.B.W, g.B.H)
	}
	if b := New(map[string]string{OptionSize: "9x7"}).B; b.W != 9 || b.H != 7 {
		t.Fatalf("9x7 board is %dx%d", b.W, b.H)
	}
	for _, bad := range []string{"0", "8", "x"} {
		if err := g.Play(bad); err == nil {
			t.Fatalf("%q was accepted", bad)
		}
	}
	play(t, g, "1", "7", "2", "7", "3")
	if sound(g, "3") != "sounds/drop.wav" {
		t.Fatal("a plain drop's sound")
	}
	// Second player blocks at 4.
	play(t, g, "4")
	if !g.Blocked || calloutFor(g) != "BLOCK!" || g.Blocks[1] != 1 || sound(g, "4") != "sounds/block.wav" {
		t.Fatalf("blocked %v, callout %q", g.Blocked, calloutFor(g))
	}
	if g.Turn() != 0 || g.Outcome().Over {
		t.Fatalf("turn %d, outcome %+v", g.Turn(), g.Outcome())
	}
}

func TestForkAndAwards(t *testing.T) {
	g := New(nil)
	// First player: 2, 3 on the bottom; then 4 makes three with both ends open.
	play(t, g, "2", "2", "3", "3", "4")
	if !g.Forked || calloutFor(g) != "FORK!" {
		t.Fatalf("forked %v, callout %q", g.Forked, calloutFor(g))
	}
	play(t, g, "1", "5") // the other end: four
	o := g.Outcome()
	if !o.Over || o.Winner != 0 || result(g, o, Rules.SeatNames) != "FOUR!" || sound(g, "5") != arcade.SoundFinish {
		t.Fatalf("outcome %+v, result %q", o, result(g, o, Rules.SeatNames))
	}
	got := awards(g, 0, o)
	for _, want := range []string{"fork", "four_straight"} {
		if !slices.Contains(got, want) {
			t.Errorf("awards %v: missing %s", got, want)
		}
	}
	if awards(g, 1, o) != nil {
		t.Error("the loser earned something")
	}
	// A diagonal.
	g = New(nil)
	play(t, g, "1", "2", "2", "3", "3", "4", "3", "4", "4", "7", "4")
	if o := g.Outcome(); result(g, o, Rules.SeatNames) != "DIAGONAL!" || !slices.Contains(awards(g, 0, o), "diagonal") {
		t.Fatalf("diagonal: %q %v", result(g, o, Rules.SeatNames), awards(g, 0, o))
	}
}

func TestComputer(t *testing.T) {
	for level := 1; level <= 3; level++ {
		g := New(nil)
		for i := 0; i < 8 && !g.Outcome().Over; i++ {
			mv := Rules.AI(g, level)
			if err := g.Play(mv); err != nil {
				t.Fatalf("level %d suggested %q: %v", level, mv, err)
			}
		}
	}
}

// Every size, set and frame draws inside its space, and attract mode is
// the same picture for the same frame.
func TestDrawing(t *testing.T) {
	sgr := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	const ax, ay = 3, 2
	inside := func(name string, w, h int, c *arcade.Canvas) {
		t.Helper()
		for y, raw := range strings.Split(c.String(), "\n") {
			in := y >= ay && y < ay+h
			for x, r := range []rune(sgr.ReplaceAllString(raw, "")) {
				if r != ' ' && (!in || x < ax || x >= ax+w) {
					t.Fatalf("%s at %dx%d: %q at %d,%d, outside its space", name, w, h, r, x, y)
				}
			}
			if !in && strings.Contains(raw, "\x1b[") {
				t.Fatalf("%s at %dx%d: colour on row %d, outside its space", name, w, h, y)
			}
		}
	}
	for _, sz := range []string{"7x6", "8x7", "9x7"} {
		for _, area := range [][2]int{{82, 34}, {60, 22}, {42, 18}, {40, 12}, {30, 13}} {
			c := arcade.New(100, 45, arcade.NewPalette(nil))
			g := New(map[string]string{OptionSize: sz})
			play(t, g, "1", "2", "3")
			l := plainLook(&discSets[0], &frames[0])
			l.labels, l.cursor, l.last = true, 2, g.B.Last
			l.fall = &falling{cell: g.B.Last, p: .5}
			drawBoard(c, ax, ay, area[0], area[1], g.B, l)
			inside(sz, area[0], area[1], c)
		}
	}
	sel := map[string]string{kindDiscs: "neon", kindFrame: framePrefix + "cherry"}
	for _, u := range unlockables() {
		for _, area := range [][2]int{{40, 17}, {21, 6}, {20, 6}} {
			c := arcade.New(100, 45, arcade.NewPalette(nil))
			preview(c, u.ID, sel, ax, ay, area[0], area[1], u.Tier != arcade.Starter)
			inside(u.ID, area[0], area[1], c)
		}
	}
	for _, sz := range []string{"7x6", "8x7", "9x7"} {
		c := arcade.New(100, 45, arcade.NewPalette(nil))
		setupPreview(c, map[string]string{OptionSize: sz}, ax, ay, 30, 13)
		inside("setup "+sz, 30, 13, c)
	}
	draw := func(frame int) string {
		c := arcade.New(100, 45, arcade.NewPalette(nil))
		attract(c, ax, ay, 76, 14, frame)
		inside("attract", 76, 14, c)
		return c.String()
	}
	for f := 0; f < 300; f += 7 {
		if draw(f) != draw(f) {
			t.Fatalf("attract mode changes within frame %d", f)
		}
	}
	if draw(0) == draw(30) {
		t.Fatal("attract mode doesn't move")
	}
}

// A game against the computer in a big pane: the NEW GAME screen's board
// size, then the table with the threats panel.
func TestArcadeGame(t *testing.T) {
	srv, ch := startServer(t)
	v := srv.Enter(ch, "alice", 120, 40)
	srv.FrameContaining(v, "PRESS ENTER")
	srv.Key(v, "enter")
	srv.FrameContaining(v, "THE GIFT SHOP")
	srv.Key(v, "enter") // 1 PLAYER VS CPU
	frame := srv.FrameContaining(v, "◂ 7 x 6 ▸")
	if !strings.Contains(frame, "THE CLASSIC") {
		t.Fatalf("NEW GAME:\n%s", frame)
	}
	srv.Key(v, "up")
	srv.Key(v, "right")
	srv.FrameContaining(v, "◂ 8 x 7 ▸")
	srv.Key(v, "enter")
	frame = srv.FrameContaining(v, "VS CPU · NORMAL · 8 x 7")
	for _, want := range []string{"WINS NEXT", "56 SLOTS LEFT", "▼"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("the table is missing %q:\n%s", want, frame)
		}
	}
	srv.Key(v, "4") // a drop; the computer answers
	srv.FrameContaining(v, "54 SLOTS LEFT")
}

// Two people at a small pane: the plain layout.
func TestPlainTable(t *testing.T) {
	srv, ch := startServer(t)
	a, b := srv.Enter(ch, "alice", 60, 20), srv.Enter(ch, "bob", 60, 20)
	for _, v := range []*plugintest.Viewer{a, b} {
		srv.FrameContaining(v, "Enter to play")
		srv.Key(v, "enter")
	}
	srv.FrameContaining(a, "M: sit down")
	for _, v := range []*plugintest.Viewer{a, b} {
		srv.Key(v, "m")
		srv.Key(v, "enter")
	}
	srv.FrameContaining(a, "your move")
	srv.Key(a, "4")
	srv.FrameContaining(b, "your move")
	srv.Key(a, "4")
	srv.FrameContaining(a, "not your turn")
	time.Sleep(10 * time.Millisecond)
}
