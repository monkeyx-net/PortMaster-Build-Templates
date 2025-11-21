Shamanic Mountain Guardian
==========================

Shamogu is a coffee-break roguelike game with a focus on tactical movement and
careful timing of totemic spirit invocation and comestible consumption.
Visibility and noise stealth mechanics also play an important role.

*“All life in your mountain is being corrupted. Beasts are losing their minds
and becoming aggressive. These disturbing events started to happen after a
dungeon portal suddenly appeared at the top of the mountain. As the guardian,
you decide to explore that dungeon in order to find and destroy the source of
corruption… You hope luck will be on your side!”*

Website
-------

![Shamogu introduction screen](https://anaseto.codeberg.page/games/shamogu/intro-screen-tiles.png)

You can visit the [project's development
website](https://anaseto.codeberg.page/games/shamogu) for more information and
tips about the game. The game can be played on the browser [on
itch.io](https://anaseto.itch.io/shamogu), too. You may also want to read the
[Shamogu: design
ramblings](https://anaseto.codeberg.page/games/shamogu/design.html) article.

Install from Sources
--------------------

First, you need to install the [go compiler](https://golang.org/) (Go 1.23 or
later), so that the `go` command is available.

Shamogu uses the [gruid](https://codeberg.org/anaseto/gruid) library for
grid-based user interfaces, which offers three different rendering drivers:
terminal, graphical SDL2, and browser. You don’t need to install it manually.

### Terminal (ASCII)

You can build a native ASCII version from source by using the following
command:

	go build

You should now have a working `shamogu` executable in the current directory.
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

This will install the [go-sdl2](https://github.com/veandco/go-sdl2/sdl) Go
bindings for SDL2. You need to install
[SDL2](https://libsdl.org/download-2.0.php) first.

You should now have a working `shamogu` executable in the current directory.
Use the `-help` option for command-line usage.

### Browser (Tiles or ASCII)

You can also build a WebAssembly version with:

    GOOS=js GOARCH=wasm go build -o play-wasm/shamogu.wasm

You can then play by serving the `play-wasm` directory containing the wasm file
via http. The directory should also contain the `wasm_exec.js` file from your
Go distribution, as explained in [Go Wiki:
WebAssembly](https://go.dev/wiki/WebAssembly).

Documentation
-------------

See the man page shamogu(6) in the `docs/` folder for more information on
command line options and use of the replay file. For example:

    shamogu -r _

launches an auto-replay of your last game.
