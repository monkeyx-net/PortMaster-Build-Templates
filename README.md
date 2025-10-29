# PortMaster-Build-Templates
Holding repo for meta information and assets needed to build/update ports in PortMaster



### New Port Structure:

Ports are now contained within the `port` top level directory, each port has its own sub-directory named after the port itself. Each port must adhere to the `portname` rules stated above. Each port must have a `port.json`, `screenshot.{jpg,png}`, `README.md`, a port script and a port directory. It may optionally include a `cover.{jpg,png}`.

The script should have capital letters (like `Port Name.sh`) and must end in `.sh`, the port directory should be the same as the containing directory. Some legacy ports have different names, new ports won't be accepted unless they follow the new convention.

Scripts and port directories must be unique across the whole project, checks will be run to ensure this is right.

A port directory might look like the following:

```
- portname/
  - port.json
  - README.md
  - screenshot.jpg
  - gameinfo.xml
  - cover.jpg (Optional)
  - Port Name.sh
  - portname/
    - licenses/LICENSE Files
    - <portfiles here>
```

The above file structure would create a `portname.zip` file.

#### README.md

This adds additional info for the port on the wiki, [we have a handy README.md generator here](http://portmaster.games/port-markdown.html).

```markdown
## Notes

{ADDITIONAL NOTES HERE}

Thanks to {THANKS INFO HERE} for the source code. Also thanks to {PORTER NAME} for the packaging for portmaster.

```

#### port.json

This is used by portmaster, this should include all the pertinent info for the port, [we have a handy port.json generator here](http://portmaster.games/port-json.html).

Example from 2048.

```json
{
    "version": 4,
    "name": "2048.zip",
    "items": [
        "2048.sh",
        "2048"
    ],
  "items_opt": [],
   "attr": {
    "title": "2048",
    "porter": [
      "BOT"
    ],
    "desc": "The 2048 puzzle game",
    "desc_md": null,
    "inst": "ready to run,
    "inst_md": null,
    "genres": [
      "other"
    ],
    "image": null,
    "rtr": true,
    "exp": false,
    "runtime": [],
    "store": [],
    "availability": "full",
    "reqs": [],
    "arch": [
      "aarch64",
      "armf"
    ],
    "min_glibc": ""
    }
}
```
