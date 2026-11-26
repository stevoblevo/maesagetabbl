# Tabble games: what is built and why

Artifact (private): "Maesage Tabble Games", https://claude.ai/artifact/D9KLQVehDkUBvCVrEj36bY
The same file is `web/index.html` here.

## The idea
Each game is one massage technique. The score is the receiver settling, never points.

## Built on Depthfield
v5 rebuilds the chrome on Depthfield 2.3.0 instead of imitating it.

| Before | Now | Depthfield rule |
|---|---|---|
| Five tiles under the picture and three numbered discs on it | Five real buttons on the body, each once: Moles on the blade, Hedgehogs on the low back, Golf on the calf, Call at the hand, Flip at the head | show each thing once, where it is used |
| A row of settings across the top | A sheet ("Table"): ceiling, room light, screens, high contrast | settings live in a sheet |
| A hand-written lamp and plucked string | `Depthfield.lamp` and `Depthfield.tone` | ship the kit, do not copy it |
| The state word in a font-variation hack | `Depthfield.stretch`, the page's one stretch line | width can mean something |
| "Inferred" on a scripted state | "Rehearsed, not read"; "inferred" only for the kit's guess about the giver's hands | keep rehearsal and inference apart |
| A truth line on the end card | The strawberry guardian says it | its voice is a reviewer's |
| Dark only | Night, day and high contrast; the art keeps its night in every face | the art carries its own light |

The room light is new play: dim it and the back fades while lit spots stay bright, because they make their own light.

`Depthfield.taste()` reports nothing at desktop and phone width, in night, day and contrast.

Not done: Storeybrook and Hypermetapunk. The family plan says day looks reward stillness, print looks reward impact and tempo, night looks reward a light source. That maps onto the games (Hedgehogs, Moles, Golf) and is written up as an issue, not built.

## The suit (the page draws no body)
Bleve called the first figure funky, the second (an anatomical relief in a warm light) creepy, and the third (a flat grey suit on realistic proportions) still weird. The first two drew a person. The third drew a suit on a person-shaped mannequin, which is a chalk outline with seams. The thing on the page he liked without being asked was the strawberry: a round character with one eye.

So the figure is now a character in the strawberry's family: a plush onesie.

- **Shape.** Still a height field seen from above, because the games need one: Golf rolls a ball down it, hedgehogs slide on it, and cues must land on it. The parts are chunky padded tubes with round ends (`TORSO`, `LEG`, `ARM`, each through `cap()`): a boxy body, fat sleeves and legs, mittens and boots, a big round hood. No shoulder blades, no glutes, no waist.
- **Ears.** The hood has two. They are the receiver's state on the figure itself: up when braced, flopped when settled. Where there is no reading they sit in between.
- **Cloth.** One flat neutral with a soft rim, lighter where the lamp reaches. No skin tone: the cloth is a mix of the page's two inks, so the only colour on the suit is a setting's.
- **Seams are the limits.** Every place the gun will not run is a seam or a panel: the zip down the spine, the collar, the knee gussets. Waistband, gussets, mittens and boots are colour-blocked darker.
- **Cells.** A hex grid of haptic cells covers the suit, none on the rim or in the hood. Dim at rest, brighter under the lamp, swelling with the breath (19 a minute tense, 7 settled), and ringing white where the suit hums. In Golf the cells lie along the slope like mown grass, so the green can be read.
- **Pod.** A small unit at the nape with one light: the only place the suit shows the page's accent. It dims when Stop is pressed.
- **The strawberry mirrors it.** The guardian's eye closes as the receiver settles, so the giver can read the state without looking down at the word.

Rule: do not draw skin or anatomy here. Draw the garment, the table, the cues and the creatures.

The outline is traced once at load into a single path (`SUIT`), so there is no per-frame relief.

## The receiver's state
Inferred, labelled inferred, with a confidence figure that dips after every event. Words in order: braced, caffeinated, still, curious, arriving, calm, puddle. The word's width is the state. "Leans" is the inferred preference; matching it settles the receiver about 40% more, and it drifts every 8–12 s.

In this build the state is a rehearsal. A page inside claude.ai cannot reach a camera, and Call and Flip say so on screen.

## Reading the receiver
A phone lying on the lower back, screen up, tilts a degree or two with each breath. That moves gravity across its accelerometer axes, and that is the whole signal.

The phone's page (`#sense`) only forwards samples, five posts a second. `internal/breath` does the reading:

1. Average the samples down to 10 a second. This is what removes a massage gun's buzz.
2. Keep what is slower than 0.7 s and faster than 10 s: 5 to 42 breaths a minute.
3. Find the one direction the signal moves along and look only at that.
4. Find the period at which the signal best repeats itself. That is the rate. How well it repeats is how sure the reading is.
5. Settle comes from the rate alone, on the scale the game already draws: 19 a minute is braced, 7 is a puddle. It moves only while the reading is at least 35% sure, and is held otherwise.

Shaking that is not breath (hands, a gun, a flinch) is reported separately and lowers "sure". It does not lower settle, because one sensor cannot tell the giver's hands from the receiver's flinch.

On every screen a real reading outranks the rehearsal: the slot says "read from breath", points stop moving the state, and "leans" says "not read". When the phone stops, the screens go back to the rehearsal within six seconds and say "inferred" again. A reading belongs to one sensing session: pressing "Start feeling" again, or six seconds of silence from the phone, throws the old estimate away and the next one warms up from nothing. The server sends each reading with its age, and withdraws it once the phone has been quiet for six seconds, so a screen that joins or reloads later is never shown an old reading as current.

What this is not, yet:
- Checked on a body. Every signal it has seen was made by a test.
- Personal. 19 and 7 are the same for everyone; a baseline per person would be more honest.
- A camera. This reads breath, not a face. It is the cheapest true thing to read first.

## Weather
The state is drawn a second way, so nobody has to read the slot: as weather over the receiver's own back, in Depthfield's four.

| Settled | Weather | What falls |
|---|---|---|
| under 34% | red rain | drops landing, seen from above |
| 34–76% | peach fall | petals with shadows on the table |
| 76–93% | mint clearing | motes rising |
| 93% and up | night garden | fireflies |

The lamp that lights the back takes the weather's colour, the page's one accent follows it, and a chip in the slot names it. With no reading (Call, Flip with a real person, the table's own screen) it stays peach and the chip says "no reading".

The slot also draws the last eight seconds of breath as a trace.

## Flip and the screens
The person on the table opens the page on a phone and chooses On the table. They pick a setting and touch where they want it; "Trick" places a spot that is lit and flat. Their screen owns Stop and the ceiling.

The wire is presence only, sent whole every time:

| Who | Publishes |
|---|---|
| Table | `role: "table"`, `cues` (up to 6: id, x, y, set, kind), `stopped`, `ceiling` |
| Player | `role: "play"`, `hits` (the last 12 answered ids) |

Everything read from another screen is clamped and re-checked: a spot off the body or on a no-go zone is dropped. Inside claude.ai the carrier is the platform's room; on your own network it is `cmd/tabble`.

With nobody on the table, a rehearsal asks instead and keeps asking near two sore spots.

A third role, `watch`, is for a TV (`#watch`). While someone watches, the player also publishes `view`: a small picture of the game about twelve times a second (screen, game, tension, the hand, spots, hedgehogs, ball, the current call, the last four floating words). The watcher redraws it with its own code. Names and markup are never taken from the wire; the hint line is set as text. A stop on the player's screen shows on the watcher.

## Open tuning questions
- Moles window is 3.2 s falling to 2.1 s. Study 03 used 800 ms invitation and 1200 ms response; too tight for a mouse, maybe right for a real gun.
- Golf slopes are gentle on purpose, and a ball that stops close to the cup counts.
- A perfect run reaches puddle. A normal one lands at calm.
