//go:build ignore

// gen writes the Four in a Row sounds into client/sounds: its own (a disc
// clattering into its slot, a block, a draw) and
// the Concord Arcade kit. Run from the repo root: go run tools/gen.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/JMThomas00/Concord/sdk/arcade"
)

const rate = 22050

func main() {
	check(arcade.WriteSoundKit("client"))
	sound("drop.wav", .32, func(t float64) float64 { // a plastic disc rattling down into its slot
		v := 0.0
		for i, f := range []float64{2300, 2100, 1900, 1700} {
			if s := t - float64(i)*.045; s >= 0 {
				v += .3 * math.Sin(2*math.Pi*f*s) * math.Exp(-s*160)
			}
		}
		if s := t - .2; s >= 0 { // and lands
			v += .45*math.Sin(2*math.Pi*700*s)*math.Exp(-s*70) + .25*noise(s)*math.Exp(-s*120)
		}
		return v
	})
	sound("block.wav", .35, func(t float64) float64 { // the landing, then a bright two-note blip
		v := .45*math.Sin(2*math.Pi*650*t)*math.Exp(-t*70) + .25*noise(t)*math.Exp(-t*120)
		for i, f := range []float64{880, 1318.5} {
			if s := t - .08 - float64(i)*.08; s >= 0 {
				v += .25 * math.Tanh(3*math.Sin(2*math.Pi*f*s)) * math.Exp(-s*18)
			}
		}
		return v
	})
	sound("draw.wav", .6, func(t float64) float64 { // two gentle notes, the second lower
		v := 0.0
		for i, f := range []float64{587.33, 493.88} {
			if s := t - float64(i)*.16; s >= 0 {
				v += .25 * math.Sin(2*math.Pi*f*s) * math.Exp(-s*6)
			}
		}
		return v
	})
	fmt.Println("Four in a Row sounds and the arcade kit written to client/sounds")
}

// noise is a repeatable hiss (the same every run).
func noise(t float64) float64 {
	x := math.Sin(t*12345.678) * 43758.5453
	return 2*(x-math.Floor(x)) - 1
}

func sound(name string, dur float64, f func(t float64) float64) {
	n := int(dur * rate)
	data := make([]byte, 2*n)
	for i := 0; i < n; i++ {
		t := float64(i) / rate
		fade := math.Min(1, (dur-t)/.01)
		v := max(-1, min(1, f(t)*fade))
		binary.LittleEndian.PutUint16(data[2*i:], uint16(int16(v*32000)))
	}
	var h bytes.Buffer
	h.WriteString("RIFF")
	binary.Write(&h, binary.LittleEndian, uint32(36+len(data)))
	h.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(rate), uint32(rate * 2), uint16(2), uint16(16)} {
		binary.Write(&h, binary.LittleEndian, v)
	}
	h.WriteString("data")
	binary.Write(&h, binary.LittleEndian, uint32(len(data)))
	h.Write(data)
	path := filepath.Join("client", "sounds", name)
	check(os.MkdirAll(filepath.Dir(path), 0o755))
	check(os.WriteFile(path, h.Bytes(), 0o644))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
