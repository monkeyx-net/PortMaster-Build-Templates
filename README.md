# PortMaster-Build-Templates
Holding repo for meta information and assets needed to build/update ports in PortMaster

<!-- PORT-STATUS-START -->
## Port Status

_Last checked: 2026-04-20_

| Title | Port | Status | Version | Date Updated | PM Date Updated |
|-------|------|--------|---------|--------------|-----------------|
| -- | animalcrossing | 🔵 UPDATE | commit:ce5fd36 | 2026-04-20 | -- |
| Defendguin | defendguin | 🟢 OK | commit:8143007 | 2026-01-29 | 2026-02-14 |
| -- | edgar | 🟢 OK | 1.38 | 2026-03-27 | -- |
| -- | epiphany | 🟢 OK | commit:c9ba7b4 | 2026-02-24 | -- |
| Harmonist | harmonist | 🟢 OK | v1.0.2 | 2026-02-27 | 2026-01-07 |
| Manic Miner | manicminer | 🟢 OK | 1.1.1 | 2026-03-01 | 2025-11-04 |
| Shamogu | shamogu | 🟢 OK | v1.5.0 | 2026-04-16 | 2025-12-29 |
| Space Invaders Part 2 | sip2 | 🟢 OK | commit:0c589d9 | 2026-03-02 | 2026-03-17 |
| Ship of Harkinian | soh | 🟢 OK | 9.2.3 | 2026-04-16 | 2026-03-16 |
| Ship of Harkinian 2 | soh2 | 🟢 OK | 4.0.2 | 2026-03-10 | 2026-03-16 |
| SpaghettiKart | spaghettikart | 🟢 OK | 1.0.0 | 2026-03-14 | 2025-09-08 |
| Stunt Car Racer | stuntcarracer | 🟢 OK | continuous | 2026-02-25 | 2026-01-25 |

_Checked: 12 &nbsp; OK: 11 &nbsp; Updates: 1 &nbsp; Errors: 0 &nbsp; Skipped: 0_
<!-- PORT-STATUS-END -->

## A slightly different approach

Ports are built using Github Actions to produce binaries for aarch64/arm64 and x86_64.Using Github actions allows leveraging Github runners to compile and produce the ports. The github actions are being used to creat the binaries and meta data for PortMaster. 

One advantage of this approach is that ports can be produced on a much wider variety of machines without having to leverage virtual machines or docker images to create binaries of aarch64 on an x86_64 device or vice versa on aarch64 device.

All builds are fully logged which actually helps troubleshooting the build process.

## Do you need to be a github actions guru to use this approach?

Quick answer: No!

```bash
    - name: Build ${{ needs.build-arm64.outputs.game_name }} # This variable translates as Shamogu
      run: |
        go build -tags sdl
```

Note the run: process as it runs the exactly the compilation command you would normally use.

Corresponding build log 

```
go: downloading go1.23.0 (linux/arm64)
go: downloading codeberg.org/anaseto/gruid-sdl v0.6.0
go: downloading codeberg.org/anaseto/gruid v0.25.0
go: downloading github.com/veandco/go-sdl2 v0.4.40
go: downloading golang.org/x/image v0.29.0
```

Also needed to add a golang repo but this step would have been needed no matter what build process was chosen. These steps will documented as part of the build process.

## Pre Requisites

- A github account
- Basic knowledge of using:- 
  - git
  - bash
  - application compilation
  - a device to test the port on.

## Build Process

The actions created can be viewed at this [base repo.](https://github.com/monkeyx-net/PortMaster-Build-Templates)

Ubuntu 20.04 is starting to show its age a bit especially in the world of cross compilation. In an ideal world we would be able to an Ubuntu 22.04 aarch64 runner to create the binaries. As 22.04 has full support for aarch64 complilation. Due to limitations with aarch64 support on 20.04. We utilise a 22.04 x86_64 image to run a 20.04 docker image to build binaries that are compatible with aarch64 PortMaser supported devices.


### Step 1

Create a port.json file. We will use the first game used for this process as a working example.

### port.json for Shamogu

```json
{
    "version": 4,
    "name": "shamogu.zip",
    "items": [
        "Shamogu.sh",
        "shamogu"
    ],
  "items_opt": [],
   "attr": {
    "title": "Shamogu",
    "porter": [
      "monkeyx-net"
    ],
    "desc": "Shamogu is a coffee-break roguelike game with a focus on tactical movement and careful timing of totemic spirit invocation and comestible consumption. Visibility and noise stealth mechanics also play an important role.",
    "desc_md": null,
    "inst": "ready to run",
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
      "x86_64"
    ],
    "min_glibc": ""
    }
}
```

[We have a handy port.json generator here](http://portmaster.games/port-json.html).


## Step 2

Create images for testing? Maybe leave this as not needed for first run test of a port :)

## Step 3

Create a fork/copy of the project. A fork might better as you can then feedback this build process to the project. 


## ToDo List

- [x] Create aarch64 and x86_64 binaries using a github actions
- [ ] Automate the creation of meta data and bash startup files (About 80% done)
- [ ] Cache actions to speed up testing on multiple builds
- [ ] Document how to use the process
- [ ] Use common apt-get statements between aarch64 and x86_64.
- [ ] Document how the development process works for aarch64 ports
- [ ] Document how the development process works for x86_64 ports

