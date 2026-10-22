# Four in a Row

Four in a Row for the terminal and for
[Concord](https://github.com/JMThomas00/Concord) channels, from the same
program. In Concord it's a game on the pier, **THE BOARDWALK**: a title screen
where the computer plays itself, round discs that fall into place, a winning
four that lights up, **BLOCK!** and **FORK!** called out, chiptune sounds, a
Hall of Fame, and **Sand Dollars** to spend at **THE GIFT SHOP** on disc sets
and frames.

## Play it on your own computer

**Download:** from the [Releases](https://github.com/JMThomas00/concord-four-in-a-row/releases)
page, get the zip for your system (`concord-four-in-a-row_windows_amd64.zip`,
`concord-four-in-a-row_darwin_arm64.zip` for Apple silicon,
`concord-four-in-a-row_linux_amd64.zip`, ...), unzip it, and run the program
inside from a terminal:

```sh
./concord-four-in-a-row          # Windows: .\concord-four-in-a-row.exe
```

On macOS, if it's blocked as being from an unidentified developer, run
`xattr -d com.apple.quarantine concord-four-in-a-row` once. **Or with Go
installed:** `go install github.com/JMThomas00/concord-four-in-a-row@latest`,
then run `concord-four-in-a-row`.

It opens on the arcade, as in Concord: 1 PLAYER VS CPU at three levels,
2 PLAYERS on one keyboard, or a NETWORK GAME. For a network game, one player hosts and is
shown their address and a 6-character code; the other chooses join and types
both. Your record, unlockables and hall of fame are kept on this computer, and the
sounds play from the `client` folder that comes in the zip: keep it next to
the program. (Installed with `go install`, only the menu sounds play.)

## Play it on a Concord server

You need to be the server owner, or have the **Manage Plugins** permission.

1. In Concord, open **Server Settings → Plugins** and press **I** (install).
2. Type `JMThomas00/concord-four-in-a-row` and press Enter. Concord downloads
   the latest release for the server's own system, verifies it, and starts
   it: no restart, no files to edit.
3. Open **Server Settings → Channels**, create a channel, and choose
   **Four in a Row Table** as its type. Its options:
   - **Seating**: *seats* (one board; sit down from the menu, everyone else
     watches), *challenge* (a lobby where members challenge each other), or
     *private* (your own games with opponents you pick).
   - **Allow spectators**, **Computer opponent**, and **Computer strength**
     (easy, normal or hard).
4. Select the channel and press **Tab** (or click it) so your keys go to the
   game. Press **Enter** on the title screen, then pick from the menu.

## The arcade

Everyone who opens the channel starts on the title screen. The menu:

- **1 PLAYER VS CPU** `◂ NORMAL ▸`: a game of your own against the computer;
  leave and come back to carry on.
- **TAKE A SEAT** (seats channels), **2 PLAYERS** and **WATCH** (challenge
  channels), or **NEW GAME** and **YOUR GAMES** (private channels).
- **THE GIFT SHOP**: your discs and frame. 12 disc sets (classic red and
  yellow, boardwalk, saltwater taffy, seashells ... and pirate gold, arcade
  tokens, grapes) and 7 frames (classic blue, midnight, lagoon, cherry,
  tangerine, bubblegum, driftwood). Everyone sees the game in their own.
- **HALL OF FAME**, **HOW TO PLAY** and **OPTIONS** (your sound and effects).

**Every new game starts on a NEW GAME screen** to choose the board: the
classic 7×6, 8×7 or 9×7. The board grows with the window, so a bigger pane
gets bigger discs. The panel on the left shows **WINS NEXT**: any column where
either player could make four with their next disc. You earn a **Sand Dollar**
for each new achievement, every three wins in a row and every ten games. A
pane smaller than 64×24 gets the plain board instead.

To update later: select it in **Server Settings → Plugins**, press **U**, then
Enter. A failed update rolls back by itself.

## Rules

Players take turns dropping a disc into a column; it falls to the lowest empty
slot. Four of your discs in a line wins: across, up and down, or diagonally. A
full board with no four is a draw.

**Keys:** ←/→ aim and Enter drops, or press a column's number to drop there.

## License

MIT. See [LICENSE](LICENSE).
