package game

import (
	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/Concord/sdk/table"
)

// The arcade front door's personality: THE BOARDWALK, a game on the pier
// between the taffy stand and the claw machines. Discs fall into place,
// a winning four lights up, BLOCK! and FORK! get called, and Sand Dollars
// buy disc sets and frames at THE GIFT SHOP. The board size is chosen on the
// NEW GAME screen for each game, and the table grows with the pane.
var arcadeLook = &table.Arcade{
	Title:   "FOUR IN A ROW",
	Tagline: "A GAME ON THE PIER",
	HowTo: []string{
		"Take turns dropping a disc into a column; it falls to the lowest",
		"empty slot. Four of yours in a line wins: across, up and down, or",
		"diagonally.",
		"Aim with ←→ and drop with Enter, or press a column's number.",
		"Stop a four before it happens: that's a BLOCK! Leave two ways to",
		"win at once and they can only block one: that's a FORK!",
		"A full board with no four is a draw.",
	},
	Keys: []arcade.Key{
		{Key: "←→", Does: "aim"},
		{Key: "Enter", Does: "drop"},
		{Key: "1-9", Does: "drop in a column"},
	},
	Sounds:      true,
	Reward:      "SAND DOLLAR",
	Collection:  "THE GIFT SHOP",
	Unlockables: unlockables(),
	Kinds:       []table.Kind{{ID: kindDiscs, Label: "DISCS"}, {ID: kindFrame, Label: "FRAME"}},
	Preview:     preview,
	Attract:     attract,
	Result:      result,
	Options: []table.Option{
		{Key: OptionSize, Label: "BOARD", Values: []string{"7x6", "8x7", "9x7"},
			Names: []string{"7 x 6", "8 x 7", "9 x 7"}, Default: "7x6",
			Describe: func(v string) string {
				switch v {
				case "8x7":
					return "BIGGER: MORE ROOM TO PLAN"
				case "9x7":
					return "THE BIGGEST: LONG GAMES"
				}
				return "THE CLASSIC"
			}},
	},
	SetupPreview: setupPreview,
	Grow:         true,
}

func result(tg table.Game, o table.Outcome, seats []string) string {
	g := tg.(*Game)
	switch {
	case o.Winner < 0:
		return "FULL BOARD!"
	case o.Reason != "four in a row":
		return "YOU WIN!" // a resignation: the kit's own line says who
	case g.B.Diagonal(g.Line()):
		return "DIAGONAL!"
	}
	return "FOUR!"
}
