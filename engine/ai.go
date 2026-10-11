package engine

import (
	"math/rand/v2"
	"sort"
)

// Best picks a column for the side to move. level runs from 1 (easy: looks
// two discs ahead and sometimes misses) to 3 (hard: eight ahead). It never
// misses an immediate win, and from level 2 it always blocks one.
func Best(b *Board, level int) int {
	depth := map[int]int{1: 2, 2: 5, 3: 8}[level]
	if depth == 0 {
		depth = 5
	}
	if b.W*b.H > 48 && depth > 7 { // the bigger boards: a little shallower
		depth = 7
	}
	me := b.ToMove
	if w := b.Wins(me); len(w) > 0 {
		return w[0]
	}
	if level >= 2 {
		if threats := b.Wins(1 - me); len(threats) > 0 {
			return threats[0]
		}
	}
	s := b.Clone()
	best, bestScore := -1, -1<<30
	for _, c := range order(s.W) {
		if !s.Open(c) {
			continue
		}
		last := s.Last
		s.drop(c)
		score := -negamax(s, depth-1, -1<<30, 1<<30)
		s.undo(c)
		s.Last = last
		if level == 1 {
			score += rand.IntN(60) - 30 // easy: a little careless
		}
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}

// order is the columns, centre first: better moves first prune more.
func order(w int) []int {
	out := make([]int, w)
	for i := range out {
		out[i] = i
	}
	off := func(c int) int { return abs(2*c - (w - 1)) }
	sort.SliceStable(out, func(a, b int) bool { return off(out[a]) < off(out[b]) })
	return out
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

const win = 100000

// negamax scores the position for the side to move.
func negamax(b *Board, depth, alpha, beta int) int {
	if b.Last >= 0 && b.LineAt(b.Last) != nil {
		return -(win + depth) // the last disc won: bad for us, worse the sooner
	}
	if b.Played == b.W*b.H {
		return 0
	}
	if depth == 0 {
		return evaluate(b, b.ToMove)
	}
	best := -1 << 30
	for _, c := range order(b.W) {
		if !b.Open(c) {
			continue
		}
		last := b.Last
		b.drop(c)
		score := -negamax(b, depth-1, -beta, -alpha)
		b.undo(c)
		b.Last = last
		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break
		}
	}
	return best
}

// evaluate counts every window of four: open threes and twos for each
// side, and discs in the centre column.
func evaluate(b *Board, me int) int {
	score := 0
	mine, theirs := int8(me+1), int8(2-me)
	for r := 0; r < b.H; r++ {
		if b.At(b.W/2, r) == mine {
			score += 3
		}
	}
	for r := 0; r < b.H; r++ {
		for c := 0; c < b.W; c++ {
			for _, d := range dirs {
				if b.At(c+3*d[0], r+3*d[1]) < 0 {
					continue
				}
				m, t := 0, 0
				for k := 0; k < 4; k++ {
					switch b.At(c+k*d[0], r+k*d[1]) {
					case mine:
						m++
					case theirs:
						t++
					}
				}
				switch {
				case t == 0 && m == 3:
					score += 50
				case t == 0 && m == 2:
					score += 5
				case m == 0 && t == 3:
					score -= 60
				case m == 0 && t == 2:
					score -= 5
				}
			}
		}
	}
	return score
}
