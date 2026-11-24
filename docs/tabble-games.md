# Tabble games: what is built and why

Artifact (private): "Maesage Tabble Games", https://claude.ai/artifact/D9KLQVehDkUBvCVrEj36bY
The same file is `web/index.html` here.

## The idea
Each game is one massage technique. The score is the receiver settling, never points.

## The body
The receiver is a height field, seen from above, prone. Every part is a tube along the body: a centre line, a half-width and a height, each a smooth curve through a few measured points (`TORSO`, `LEG`, `ARM` in the page). Parts meet like clay. Shoulder blades, a spine groove and the two glute mounds are added on top.

It is drawn twice: a relief lit by a lamp that follows the hand, and fine contour lines over it. It breathes: 19 breaths a minute tense, 7 settled. The suit's hum warps the contours and sends white ripples out.

## The receiver's state
Inferred, labelled inferred, with a confidence figure that dips after every event. Words in order: braced, caffeinated, still, curious, arriving, calm, puddle. The word's width is the state. "Leans" is the inferred preference; matching it settles the receiver about 40% more, and it drifts every 8–12 s.

In this build the state is a rehearsal. A page inside claude.ai cannot reach a camera, and Call and Flip say so on screen.

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
