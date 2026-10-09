This project was forked into Git to allow code customisations for to get game running on devices supported by the [PortMaster Project](https://github.com/PortsMaster)


README for Defendguin

by Bill Kendrick
bill@newbreedsoftware.com
http://www.newbreedsoftware.com/defendguin/

Version 0.0.13


June 27, 2023


DESCRIPTION
-----------
  "Defendguin" is based loosely on William's classic 1981 arcade
  game, "Defender" designed by Eugene Jarvis and Larry DeMar.
  (See https://en.wikipedia.org/wiki/Defender_(1981_video_game) and
  https://www.arcade-museum.com/game_detail.php?game_id=7547)


STORY
-----
  Mysterious aliens are now attacking earth, kidnapping little penguinoids
  and converting them into mutants.  Helping them on their way are
  some other nasty alien ships, of which there are plenty...


DOCUMENTATION
-------------
  Important documentation for "Defendguin" is contained in multiple files.
  You can find them in the "docs" folder/directory.  Please look through them:

    AUTHORS.txt    - Credits for who made this game.
    CHANGES.txt    - Changes since the previous versions of "Defendguin."
    COPYING.txt    - The GNU Public License, which "Defendguin" is under.
    INSTALL.txt    - Instructions on requirements, compiling and installing.
    README.txt     - (This file.)  Game story, usage, rules and controls.
    TODO.txt       - A wish-list for this version of "Defendguin."


INSTALLATION
------------
  See INSTALL.txt.


RUNNING THE GAME
----------------
  Just type "defendguin" to get started.

  The program accepts some game options:

    --disable-sound     - If sound support was compiled in, this will
                          disable it for this session of the game.
                          (Also "--nosound" or "-q")

    --fullscreen        - Run in fullscreen mode.
                          (Must be supported by your display configuration.)
                          (Also "-f")


  ... and can also be asked to show some information:

    --help              - Display a help message summarizing game controls,
                          and quit.   (Also "-h")

    --usage             - Display a brief message summarizing command-line
                          options and quit.   (Also "-u")

    --copying           - Display copying information (GPL info) and quit.
                          (Also "-c")


  When you start the game, a screen will appear stating that the game data
  is loading.  When prompted, press any key or a joystick firebutton to
  continue to the title screen.


TITLE SCREEN
------------
  At the title screen, you can choose to begin either a ONE PLAYER, or
  TWO PLAYER game, go to the OPTIONS screen, or EXIT the program.

  Use the [ UP ARROW ] and [ DOWN ARROW ] keys to choose an option, then
  press [ SPACE ], [ ENTER / RETURN ] or a joystick firebutton to
  activate the selected option.

  Note: Once you've played at least one game, the last scores for each
  player are displayed at the bottom left and right corners of the title
  screen.


OPTIONS
-------
  The Options screen allows you to control settings in Defendguin.
  Use the [ UP ARROW ] and [ DOWN ARROW ] keys to select an
  option and [ LEFT ARROW ] and [ RIGHT ARROW ] to change its value.
  Press [ ESCAPE ] when you're done to return to the main menu.

  Note: Your settings will be saved in a file called ".defendguinrc" in
  your home directory.

  The Options screen lets you change the volume of both sound effects
  and music in Defendguin, as well as configure a joystick.
  You can choose which fire button fires lasers, and which fires smart bombs,
  and pick which axes for X and Y controls of your ship's movement.

  You may also choose to disable the horizontal controls (by choosing "-1"
  as the X axis input, and setting the "NO_HAT_X" option to "1"), and
  configure buttons for thrust and reverse controls.

  For example, when using a Sony PlayStation 2 controller via a USB adapter,
  I chose the following settings:

    Fire - 2  (the "X" button)
    Bomb - 1  (the "O" button)
    X axis - 4  (the digital directional pad)
    Y axis - 5  (the digital directional pad)

  And a Nintendo Switch wired (USB) controler, I chose:

    Fire - 1  (the "B" button)
    Bomb - 2  (the "A" button)
    X axis - 0  (the left analog stick)
    Y axis - 1  (the left analog stick)

  The "sensing" section on the right lets you detect which inputs
  on your controller are being detected.  Press buttons, move sticks (axes),
  or push d-pads (hats) to determin what Defendguin sees.

  (Note: Defendguin used the Simple DirectMedia Layer library (libSDL).)


GAME CONTROLS
-------------
  The following controls are available:

    [ LEFT ARROW KEY ]     -- Point your ship left and thrust left.
    [ JOYSTICK AXIS LEFT ]
    [ JOYSTICK HAT LEFT ]

    [ RIGHT ARROW KEY ]    -- Point your ship right and thrust right.
    [ JOYSTICK AXIS RIGHT ]
    [ JOYSTICK HAT RIGHT ]

  -- or, for hard-core classic Defender fans: --

    [ 1 KEY ]               -- Thrust in the direction your ship is pointing.
    [ JOYFIRE THRUST ]

    [ 2 KEY ]               -- Reverse the direction your ship is pointing.
    [ JOYFIRE REVERSE ]


  [ UP ARROW KEY ]       -- Move your ship up.
  [ JOYSTICK AXIS UP ]
  [ JOYSTICK HAT UP ]

  [ DOWN ARROW KEY ]     -- Move your ship down.
  [ JOYSTICK DOWN ]
  [ JOYSTICK HAT DOWN ]

  [ SPACE KEY ]          -- Fire lasers.  (Destroys all enemies in its path)
  [ JOYFIRE B ]

  [ CTRL / SHIFT / ALT ] -- Fire a smart bomb.  (Destroys all visible enemies)
  [ JOYFIRE A ]

  [ TAB / P KEY ]        -- Pause / unpause.

  [ ESCAPE ]             -- Abort the game in progress, return to the title menu.


PLAYING THE GAME
----------------
  When you begin the game, your ship appears over some icy mountains
  which are crawling with penguinoids (the good guys).  A number of alien
  ships are flying around above the ground.

  MOVEMENT
  --------
    When you reverse direction your ship will slide to the opposite side of
    the screen, so that you can see more ahead of your ship.  As you fly
    around, your ship remains on one side of the screen, and everything else
    passes by around you.

    If you move up or down, your ship simply raises or lowers.

  RADAR
  -----
    At the very top of the screen in the center is a radar display.  You can
    see the "entire" planet, including what you see on the main part of the
    screen, in the center, outlined in green.

    Miniature versions of your ship, the penguinoids, and the aliens are
    also visible on the radar.

    Note: The planet 'wraps around', if you go far enough in one direction
    you'll arrive back to where you started.

  ALIENS
  ------
    UFO'S
    -----
      The little alien ships are the ones which abduct penguinoids.
      For the most part, they mind their own business.  They occasionally
      shoot at you, and are especially agressive when trying to catch a
      penguinoid.

      When they start abducting a penguinoid, you'll hear a special sound.
      You'll also notice a yellow and black alert box appear over the
      penguinoid on the radar screen.

      The UFO takes a moment to beam up the penguinoid, and then begins
      flying upwards towards space.  If they reach the top of the screen,
      the screen will flash, you'll hear a pop, and they will have converted
      the penguinoid into a horrible mutant penguinoid alien!

      If you manage to shoot down the UFO before it gets to the top, the
      penguinoid will start falling.  If it falls too far before reaching the
      ground, it will die!  (The screen will flash and you'll hear a "splat!"
      sound.)  If it doesn't have to fall too far, it will be safe.

      Your best bet is to shoot the UFO, and then fly towards the penguinoid
      to catch them.  When you catch them, they will hold on to your ship
      as you fly around.  Fly down to the mountains and he'll jump off your
      ship and thank you.  (You get more points for doing it this way.)

      Be careful!  You can shoot penguinoids and kill them, too -- both in
      mid-air AND on the ground!

    BOMBERS
    -------
      These rainbow-colored alien ships fly gracefully up and down,
      and occasionally spit out mines.  Other than that, just don't run
      into them and shoot them when you get the chance.

      Note: If they go very high in the sky or low to the ground,
      they'll "wrap around", appearing on the opposite side of the screen.

    MINES
    -----
      Bombers drop these spikey grey balls.  They can be in the sky or at
      ground level, so be careful!  If you get too close, they'll start
      flashing.  Note: you can't shoot them!

      After floating for a few moments, they disappear with a 'pop.'

    PODS
    ----
      Pods are purplish, rotating, berry-shaped aliens which fly about
      randomly.  If you shoot them, they release a bunch of swarmers,
      and you'll hear a special sound.  They don't proactively attack.

    SWARMERS
    --------
      When you shoot a pod, these little redish balls appear.  If you're
      nearby, they'll start hunting you down, shooting agressively.
      Try to attack these from afar, otherwise they'll gang up on you.

    BAITERS
    -------
      If you take too long to complete a level, Baiters begin appearing.
      (They yell out at you: "You will be absorbed!")  They relentlessly
      persue you, so watch your radar and blast them as quickly as you
      can!


  WEAPONS
  -------
    LASERS
    ------
      To fire lasers, press the [ SPACE ] key or your chosen joystick
      firebutton.  (See "Options", above.)

      Lasers streak across the screen, destroying any aliens (except mines!),
      OR PENGUINOIDS!, in their path.

      Lasers are unlimited, and you can shoot many at a time by firing
      repeatedly.

      Note: There is currently no built-in autofire option.
      See TODO.txt for plans.

    SMART BOMBS
    -----------
      When you begin the game, you have three smart bombs.  To use one,
      press [ CTRL ], [ SHIFT ] or [ ALT ], or your chosen joystick
      firebutton (see "Options", above).

      The screen flashes and all aliens and alien bullets on the screen
      will explode.

      Your ship, as well as penguinoids, are not adversely affected.


  PLANETARY DESTRUCTION
  ---------------------
    If all penguinoids are gone (transformed into mutants or shot by
    your friendly fire), the entire planet will explode.  For the next
    few levels, you'll be flying in space, and all enemies that appear
    in each level will be mutants.

    Every 5th level (5, 10, 15, 20), the planet will return, repopulated
    with penguinoids.


  ADVANCING LEVELS
  ----------------
    Each level is split into 3 sections:

      * Beginning
        UFO's and other aliens appear.

      * 30 seconds later (or after the first wave is destroyed)
        More UFO's appear.

      * 15 seconds later (and every 30 seconds thereafter)
        A baiter appears.  This means you're taking too long to complete
        the level!

    There are 20 levels in the game.  The last level is special, with
    an evil boss you must destroy to complete the game.

    The level you are currently on is displayed above your score at
    the top of the screen.


  ONE-UPS AND EXTRA SMART-BOMBS
  -----------------------------
    Every time you surpass 10,000 points, one additional ship and one
    additional smartbomb are awarded.  A special sound effect will play.


  PAUSING
  -------
    If you press [ TAB ] or [ P ], the game will pause.
    The game screen will look something like a screensaver.

    To return to the game, press any key or joystick firebutton.


  GAME OVER
  ---------
    Once the game is over, the words "GAME OVER" appear on the screen.
    Simply press [ ESCAPE ] to return to the main menu.


THE END!
--------
  Enjoy!  Your feedback is welcome via e-mail:

    bill@newbreedsoftware.com

