Harmonist: Dayoriah Clan Infiltration
-------------------------------------

Harmonist is a light-hearted stealth coffee-break roguelike game with light and
noise mechanisms, making use of various terrain types and cones of view for
monsters.

*Your friend Shaedra got captured by nasty people from the Dayoriah Clan while
she was trying to retrieve a powerful magara artifact that was stolen from the
great magara-specialist Marevor Helith.*

*As a gawalt monkey, you don't understand much why people complicate so much
their lives caring about artifacts and the like, but one thing is clear: you
have to rescue your friend, somewhere to be found in this Underground area
controlled by the Dayoriah Clan.  If what you heard the guards say is true,
Shaedra's imprisoned on the eighth floor.*

*You are small and have good night vision, so you hope the infiltration
will go smoothly...*

Website
-------

![Harmonist introduction screen](https://anaseto.codeberg.page/games/harmonist/intro-screen-tiles.png)

You can visit the [project's development
website](https://anaseto.codeberg.page/games/harmonist) for more information and
tips about the game. The game can be played on the browser [on
itch.io](https://anaseto.itch.io/harmonist), too.

Install from Sources
--------------------

First, you need to install the [go compiler](https://golang.org/) (Go 1.24 or
later), so that the `go` command is available.

Harmonist uses the [gruid](https://codeberg.org/anaseto/gruid) library for
grid-based user interfaces, which offers three different rendering drivers:
terminal, graphical SDL2, and browser. You don't need to install it manually.

### Terminal (ASCII)

You can build a native ASCII version from source by using the following
command:

	go build

You should now have a working `harmonist` executable in the current directory.
Use the `-help` option for command-line usage.

This version uses the [tcell](https://github.com/gdamore/tcell) terminal
library. If your terminal supports true color, the `-t` option will provide the
same [selenized color palette](https://github.com/jan-warchol/selenized/) as
the tiles build. Otherwise, configuring your terminal to use the same colors is
recommended, but other color schemes may look nice too. As a last resort,
option `-x` may produce good enough colors as well.

### SDL2 (Tiles or ASCII)

You can build a graphical version depending on SDL2 by using the following
command:

	go build -tags sdl

This will install the [gruid-sdl](https://codeberg.org/anaseto/gruid-sdl) Go
bindings for SDL2. You need to install
[SDL2](https://wiki.libsdl.org/SDL2/Installation) first (notice you need SDL2,
not SDL3).

You should now have a working `harmonist` executable in the current directory.
Use the `-help` option for command-line usage. For example, `-F` launches the
game in fullscreen.

### Browser (Tiles or ASCII)

You can also build a WebAssembly version with:

    GOOS=js GOARCH=wasm go build -o play-wasm/harmonist.wasm

You can then play by serving the `play-wasm` directory containing the wasm file
via http. The directory should also contain the `wasm_exec.js` file from your
Go distribution, as explained in [Go Wiki:
WebAssembly](https://go.dev/wiki/WebAssembly).

Documentation
-------------

See the man page harmonist(6) for more information on command line options and use
of the replay file. For example:

    harmonist -r _

launches an auto-replay of your last game.
