# Songs

The example songs: what each is and what it shows. Sounds are in
`hits.go` (one note per patch) and `platformer.go` (a game's effects).

- `goldenbrown.go`, the Golden Brown intro: harpsichord sound design and irregular metre
  (`Meter.Bars = {3, 3, 3, 4}`).
- `amberhours.go`, "Amber Hours": an original in Golden Brown's spirit with
  Interstellar's pipe organ, a D minor harpsichord waltz in `{3, 3, 3, 4}`: recorder,
  bass, brushed kit, a ticking clock, and organ pedal, ostinato, and chords opened by an
  `AutoFilter` swell box, in sections A A B B A A after a two-pass intro.
- `playasunrise.go`, "Playa Sunrise": the same `{3, 3, 3, 4}` lilt taken to a desert
  festival at dawn, E mixolydian and lightly swung: handpan keeping Golden Brown's
  interlock, tanpura and didgeridoo drones, djembes and shaker, sitar and bansuri
  sharing a hook with meend bends, a phased pad, wind, and chimes. Its form is Amber
  Hours'.
- `lavalamp.go`, "Lava Lamp Sun": a full band (kit on a group bus, Höfner-style bass,
  double-tracked phased rhythm guitars, wah-fuzz lead with bends, Leslie organ, Juno
  arp). The reference for mixing; balance with `go run ./cmd/synth -id <id> -stems`,
  which reports each part's level while it's sounding.
- `brighterdays.go`, "Brighter Days": Avicii-style progressive house. The reference for
  arrangement dynamics: `Sidechain` keyed by the kick tracks so pumping follows the
  kick, `AutoFilter` automating a build, a riser, and a build quieter than its drop.
- `smokesignal.go`, "Smoke Signal": 2015 festival trap. Half-time drums with ratcheted
  hat rolls, a gliding distorted 808, brass, and a `TapeStop` on the parts that should
  spin down before the drop (an insert per part, so the air horns survive the silence).
- `woodlands.go`, "The Woodlands": the Firefly festival in the woods outside Dover,
  stomp-clap indie folk in G (acoustic guitar, banjo rolls, upright bass, whistled
  verse, a gang singalong with mandolin tremolo and glockenspiel, a shouted "Hey!")
  until nightfall, when crickets, fireflies, an arp, and a pad take a bridge that
  builds back into the last chorus.
- `interstellar.go`, the Interstellar main theme's organ intro: A minor in 6/8, an
  ostinato ticking against an E pedal while the chords move beneath it, the swell box
  breathing open and shut again by the loop's seam, a pedalled grand rolling the chords
  and singing a tune an octave above.
- `longlight.go`, "Long Light", flywheel's theme (ID `theme`): downtempo in G at 100, a
  chorused pad and arp pumping off the kick, a picked bass, a clean guitar in
  dotted-eighth echoes. Ported from code note for note; its noise is a different take.
- `daybreak.go`, "Daybreak": Long Light's sister at sunrise on the beach, downtempo in
  D at 108 with none of its instruments (`patches/downtempo.go`): a wowing tape pad,
  FM glass plucks opening through an `AutoFilter`, a picked bass, a gliding monosynth
  lead, vibraphone, surf and birds. B drops the drums and builds back on a roll.
- `goldenhour.go`, "Golden Hour": the third downtempo sibling, the main stage at
  sundown in A major at 110 (`patches/downtempo.go` again, all new voices): a reedy FM
  lead singing a 3-3-2 hook, a kalimba ticking that rhythm and doubling the hook at the
  end, a clean guitar into dotted-eighth echoes, a swell pad, a round synth bass, a
  tight kit, and a distant crowd under the intro and the breakdown.
- `elegy.go`, "Elegy": solo grand piano in D-flat at 63, written as one pianist's
  hands (tune, right-hand inner voices, pedalled left hand, low octaves). One
  eight-bar tune over a bass walking down the scale, played four times, each fuller,
  around a middle section in Bb minor: a run into a deceptive cadence on A major (bVI)
  with A0 in the bass is the peak, and the coda is the opening an octave up, soft.
  It plays once, no loop.
