## Notes
Thanks to the [Frozen-Bubble team](http://www.frozen-bubble.org/) for the original game and to
[Huy Chau](https://github.com/dchau360) for the C++/SDL2 rewrite.

[frozen-bubble-sdl2](https://github.com/dchau360/frozen-bubble-sdl2)

## Setup screen (Love2D launcher)

A setup screen opens before the game with **Play** selected, so one press of A starts the game.
Settings are saved in `conf/` and are kept for next time.

| Option | Description |
|--|--|
| Network mode | **Offline / local**: single player and local 2-player. **Join a server**: pick a saved server. **Host a server**: starts `fb-server` on this device before the game |
| Nickname | Up to 10 characters (letters, digits, `-`, `_`). Used in network games |
| Server (join) | Manage saved servers: A use, X edit, Y delete, **+ Add server** |
| Server port / Max connected players / Server name / LAN discovery (host) | Options passed to `fb-server` |
| Local players / controllers | Assign a controller (1-5) or keyboard to each player and choose the fire button |
| Save and exit | Save settings and return to the frontend without playing |

### Playing over the network

* **Joining:** choose *Join a server*, add the host's IP and port (default 1511), then Play.
  In game select **Net Game > Manual entry**: the server is already filled in and **Connect**
  is selected, so press A twice.
* **Hosting:** choose *Host a server*. The setup screen shows this device's IP; other players
  join that address. In game select **LAN Game** and pick `127.0.0.1` (default port), or
  **Net Game > Manual entry > Connect** (any port).
* Devices on the same network can also find each other with **LAN Game** when LAN discovery is on.
* Public servers from the community server list are shown under **Net Game**.

The server log is written to `server.log` in the port folder.

### Local multiplayer

Controller 1 is the built-in pad. Extra USB/Bluetooth controllers are numbered in the order they
connect. Upstream supports 2 local players; 3-5 local players is still marked as experimental.

## Controls

| Button | Action |
|--|--|
| D-pad Left/Right, Left Analog | Aim |
| A (or the fire button chosen in setup) | Fire / Select |
| D-pad Down | Centre the launcher |
| B | Back / Exit game |
| Start | Pause |
| Select + Start | Quit |

## Compiling

Github Actions automatic build using `recipes/ports/frozenbubble/build.sh`.
`frozenbubble-portmaster.patch` adds:

* `FB_DATA_DIR` environment variable to locate the bundled assets
* `[Network] Host` / `Port` in `settings.ini` to pre-fill the Net Game manual entry
* `FB_SERVER_BIN` / `./fb-server` so the in-game "Host a server" option finds the bundled server
