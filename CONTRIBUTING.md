# Contributing

Anyone is welcome. This is a small project with strong opinions, so here they are up front.

## How work moves

1. **Issues first.** One issue per decision or build step. Comment there before building something large.
2. **Branch and pull request.** Nothing goes straight to `main`.
3. **Review.** An automated Codex review runs on pull requests here. Fix what it finds, or say on the thread why it is wrong.
4. **We version.** A change to the page bumps `VERSION` in `web/index.html` (it shows on the lobby), adds an entry to `CHANGELOG.md`, and is tagged on `main` once merged. Nothing is deleted to make room: old versions stay reachable by tag.
5. **Say what you checked and what you did not.** Every pull request ends with both lists. "Not checked on a person" is an honest line and most of ours carry it.

## Run it

```
go run ./cmd/tabble
```

Open `http://localhost:8787`. Open it again on a phone at `#table` to be the one on the table. `go vet ./... && go test ./...` before you push.

## Languages

Server and tooling code is Go, Rust or Erlang. The page is JavaScript because that is what a browser runs. Please do not add Node or Python tooling to `main`.

## Rules the page keeps

These are not style preferences. A change that breaks one will be asked to change.

- **The receiver owns Stop.** Stop is latched. A stop from the table's screen is lifted only there.
- **Say what is known.** A rehearsed state says "rehearsed", a guess says "inferred", a reading says what it read and how sure it is. Where nothing is reading the receiver, the page says so and shows no figure.
- **Never points.** The score is the receiver settling.
- **The gun has limits.** It cuts out over the spine, neck, head and the backs of the knees, and the receiver's ceiling caps it.
- **No body.** The receiver is a hood above a quilt. Do not draw skin, anatomy or an outline of a person; draw the quilt, the table, the cues and the creatures. Five figures were tried before this (see `docs/tabble-games.md`).
- **Colour is a setting.** Each setting also has a glyph, so colour never carries meaning alone.
- **Nothing from another screen is trusted.** Presence is clamped and re-checked; no markup comes off the wire.

## Style

The page is built on Depthfield. Use the kit's tokens and components before writing new CSS, and open the page with `?df-taste` at desktop and phone width before asking for review.

## Licence

Apache 2.0. By opening a pull request you agree your contribution is under it. The copy of Depthfield in `web/vendor/` is a separate project and is not covered; do not edit it here.
