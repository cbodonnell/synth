// demo plays Golden Brown in an Ebiten window and lights its thirteen-beat phrase as it
// goes, to show a picture keeping time with the music.
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/player"
	"github.com/cbodonnell/synth/songs"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const width, height = 640, 240

var (
	unlit = color.RGBA{0x3a, 0x2c, 0x14, 0xff}
	lit   = color.RGBA{0xd9, 0xa4, 0x41, 0xff}
)

var sfxKeys = []ebiten.Key{
	ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
	ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
}

type game struct {
	mixer  *player.Mixer
	song   synth.Song
	sounds []synth.Song
	beat   synth.Position
	synced bool
}

func (g *game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyM):
		g.mixer.ToggleMute()
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if g.mixer.Current() == "" {
			g.mixer.PlayMusic(g.song)
		} else {
			g.mixer.StopMusic()
		}
	}
	for i, k := range sfxKeys {
		if i < len(g.sounds) && inpututil.IsKeyJustPressed(k) {
			g.mixer.Play(g.sounds[i])
		}
	}
	g.beat, g.synced = g.mixer.Beat()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x14, 0x10, 0x0a, 0xff})

	bars := g.song.Meter.Bars
	step := float32(width-80) / float32(g.song.Meter.Cycle()+len(bars)-1)
	x := float32(40) + step/2
	barInCycle := g.beat.Bar % len(bars)
	for b, beats := range bars {
		for k := range beats {
			c, r := unlit, float32(10)
			if g.synced && b == barInCycle && k == g.beat.Beat {
				c = lit
				r += 8 * float32(1-g.beat.Phase)
			}
			vector.FillCircle(screen, x, height/2, r, c, true)
			x += step
		}
		x += step
	}

	status := "space: play/stop   m: mute   1-9: sounds"
	if g.synced {
		status = fmt.Sprintf("bar %d beat %d   %s", g.beat.Bar+1, g.beat.Beat+1, status)
	}
	ebitenutil.DebugPrintAt(screen, g.song.Name, 16, 16)
	ebitenutil.DebugPrintAt(screen, status, 16, height-28)
}

func (g *game) Layout(int, int) (int, int) { return width, height }

func main() {
	p := player.New()
	g := &game{
		mixer:  player.NewMixer(p),
		song:   songs.GoldenBrown(),
		sounds: songs.Sounds(),
	}
	p.Prepare(append([]synth.Song{g.song}, g.sounds...)...)
	g.mixer.PlayMusic(g.song)

	ebiten.SetWindowSize(width*2, height*2)
	ebiten.SetWindowTitle("synth")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
