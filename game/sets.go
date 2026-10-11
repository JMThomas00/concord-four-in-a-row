package game

import (
	"strings"

	"github.com/JMThomas00/Concord/sdk/arcade"
)

// The unlockables, bought with Sand Dollars at THE GIFT SHOP: disc sets and
// frames. Each player sees the game in their own discs and frame.

const (
	kindDiscs = "discs"
	kindFrame = "frame"
	// Frame ids are prefixed in the unlockables list so they can't clash
	// with a disc set's.
	framePrefix = "frame:"
)

// discSet is each side's disc colour.
type discSet struct {
	arcade.Unlockable
	colours [2]string
}

func discs(id, name string, tier arcade.Tier, blurb, first, second string) discSet {
	return discSet{arcade.Unlockable{ID: id, Name: name, Tier: tier, Kind: kindDiscs, Blurb: blurb}, [2]string{first, second}}
}

var discSets = []discSet{
	discs("classic", "CLASSIC", arcade.Starter, "Red goes first.", "red", "yellow"),
	discs("boardwalk", "BOARDWALK", arcade.Starter, "Sea and sun.", "cyan", "orange"),
	discs("taffy", "SALTWATER TAFFY", arcade.Common, "Pulled fresh on the pier.", "pink", "cyan"),
	discs("shells", "SEASHELLS", arcade.Common, "Found at low tide.", "fg", "orange"),
	discs("lime", "KEY LIME", arcade.Common, "A slice of pie.", "green", "yellow"),
	discs("beachball", "BEACH BALL", arcade.Common, "Don't let it blow away.", "red", "cyan"),
	discs("neon", "NEON SIGN", arcade.Rare, "Open late.", "pink", "green"),
	discs("sunset", "SUNSET", arcade.Rare, "From the end of the pier.", "orange", "pink"),
	discs("snowcone", "SNOW CONES", arcade.Rare, "Cherry or blue raspberry.", "red", "fg"),
	discs("gold", "PIRATE GOLD", arcade.Legendary, "Doubloons from the deep.", "yellow", "fg"),
	discs("tokens", "ARCADE TOKENS", arcade.Legendary, "Saved from the claw machine.", "hub", "orange"),
	discs("grapes", "GRAPES", arcade.Legendary, "From the vineyard.", "purple", "green"),
}

// frame is the board's colour.
type frame struct {
	arcade.Unlockable
	role string
}

func frameOf(id, name string, tier arcade.Tier, blurb, role string) frame {
	return frame{arcade.Unlockable{ID: id, Name: name, Tier: tier, Kind: kindFrame, Blurb: blurb}, role}
}

var frames = []frame{
	frameOf("blue", "CLASSIC BLUE", arcade.Starter, "The one from the toy shop.", "cyanD"),
	frameOf("midnight", "MIDNIGHT", arcade.Starter, "After the lights come on.", "line"),
	frameOf("lagoon", "LAGOON", arcade.Common, "Clear and green.", "greenD"),
	frameOf("cherry", "CHERRY", arcade.Common, "Candy-apple red.", "redD"),
	frameOf("tangerine", "TANGERINE", arcade.Rare, "Bright as the lifeguard's hut.", "orangeD"),
	frameOf("bubblegum", "BUBBLEGUM", arcade.Rare, "Blown on the boardwalk.", "pinkD"),
	frameOf("driftwood", "DRIFTWOOD", arcade.Legendary, "Washed up, sanded smooth.", "comment"),
}

// unlockables is everything at THE GIFT SHOP, for the arcade.
func unlockables() []arcade.Unlockable {
	var out []arcade.Unlockable
	for _, s := range discSets {
		out = append(out, s.Unlockable)
	}
	for _, f := range frames {
		u := f.Unlockable
		u.ID = framePrefix + u.ID
		out = append(out, u)
	}
	return out
}

// set looks up a disc set (the first starter if id is unknown).
func set(id string) *discSet {
	for i := range discSets {
		if discSets[i].ID == id {
			return &discSets[i]
		}
	}
	return &discSets[0]
}

// frameFor looks up a frame (with or without its prefix).
func frameFor(id string) *frame {
	id = strings.TrimPrefix(id, framePrefix)
	for i := range frames {
		if frames[i].ID == id {
			return &frames[i]
		}
	}
	return &frames[0]
}

// colour is a side's disc colour; ghost greys it out for a locked set.
func (s *discSet) colour(side int, ghost bool) string {
	if ghost {
		return "ghost"
	}
	return s.colours[side]
}
