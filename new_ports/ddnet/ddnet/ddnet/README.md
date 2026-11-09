# DDraceNetwork

## Notes

DDraceNetwork (DDNet) is a cooperative online platformer based on Teeworlds. Online play needs Wi-Fi. Without a connection, start a local server from Play > Create Server.

There is no on-screen keyboard: set your name by editing `player_name` in `conf/settings_ddnet.cfg` (created after the first run). The port uses the OpenGL ES 3.0 renderer.

Thanks to the [DDNet team](https://github.com/ddnet/ddnet) for creating this game.

## Controls

| Button | Action |
|--|--|
| Dpad Left/Right | Move |
| Dpad Up | Jump |
| Left Analog | Aim / Mouse |
| Right Analog | Aim / Mouse |
| A | Jump |
| B | Hook |
| X | Hammer |
| Y | Pistol |
| L1 | Hook |
| L2 | Shotgun |
| R1 | Fire / Click |
| R2 | Grenade |
| L3 | Kill (respawn) |
| Select | Scoreboard |
| Start | Menu |

## Compiling

GitHub Actions [build](https://github.com/monkeyx-net/PortMaster-Build-Templates): DDNet built with CMake for OpenGL ES only, with static curl, SQLite and WavPack.
