# Maesage

One lays on the table and the other… plays.

Maesage is a haptic suit and a partner touch game. The suit lights a spot and hums under it; the player answers with hands or a massage gun; the gun tunes itself by hovering. This repo holds the playable browser studies, the server that links two screens, and the plans.

## Play it

```
go run ./cmd/tabble
```

Open `http://<this machine>:8787` on two screens on the same network. On the phone that goes under the face cradle, choose **On the table**. On the other screen, choose **05 Flip**. The person on the table picks a setting and touches where they want it; the spot lights on the player's screen. Stop and the ceiling belong to the table's screen.

One screen works too: stops 01 to 04 need nothing else.

| Stop | Game | Teaches | How |
|---|---|---|---|
| 01 | Moles | percussion | Hover a lit spot and the gun tunes itself. Hold to percuss. |
| 02 | Hedgehogs | kneading | Knots with legs. Slow circles curl one up. Rush and it bristles. |
| 03 | Golf | long glides | One glide a stroke. The cup is the knot. |
| 04 | Call | real hands | No suit, no gun. The screen calls a spot and a technique every 20 s. |
| 05 | Flip | listening | The one on the table places the spots. You answer. |

## Rules the games keep

- Colour is the setting: hush 12 Hz, wave 18–32 Hz, wake 40 Hz, knead 28 Hz. Each has a glyph too.
- Lit and rippling is real. Lit and flat is a feint on the player. Rippling with no light is the suit teasing the receiver.
- The gun cuts out over the spine, neck, head and the backs of the knees.
- The receiver's ceiling caps the gun. Stop is latched, and a stop from the table's screen is lifted only there.
- State is inferred and labelled inferred. Where there is no camera, the page says so.

## Layout

| Path | What |
|---|---|
| `web/index.html` | The whole game: one file, no libraries. Also published as a private claude.ai artifact, where it uses the platform's room instead of this server. |
| `cmd/tabble` | Go. Serves the page and relays presence between screens over one event stream. Standard library only. |
| `docs/` | Design notes: the games, the suit's oil, cooling and power. |

## Languages

Server and tooling code is Go, Rust or Erlang. The page is JavaScript because that is what a browser runs; nothing else here is Node or Python.

## Check it

```
go vet ./... && go test ./...
```

The tests cover the room: presence travels between screens, a leaving screen is forgotten, and bad presence is refused. The games themselves were play-tested in a headless browser, not yet by people on a table.

## Where plans live

Issues. One issue per decision or build step; comment there.
