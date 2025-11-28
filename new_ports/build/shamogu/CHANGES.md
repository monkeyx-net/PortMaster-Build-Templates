# v? ?

This is a quite significant release with two new advanced primary spirits,
two new expansions featuring many surprises and seven new advanced secondary
spirits, as well as various others novelties and improvements, including some
important Lignification adjustments. The design of most new and reworked
content was discussed in issues #6 and #7 of Shamogu's codeberg repository.

New content:

* New advanced new game settings and menu, accessed by pressing TAB in the
  default new game menu. They allow to select various game mods and extra
  advanced primary spirits. A new configuration option allows to make the game
  start by default with the advanced new game menu.
* New advanced primary spirit: **Spinning Crocodile**. It has a *tail-slap*
  ability that hits and unbalances all adjacent foes and reverses
  the player's current direction. The crocodile's attacks are melee attacks
  that drag foes backwards, unbalancing them. The crocodile cannot rotate
  easily: it needs a turn to reverse current direction, and while it can move
  laterally as usual, it cannot attack laterally without first moving or
  rotating.  Attack and defense bonuses compensate somewhat for the crocodile's
  mobility restrictions, but it's still considered an *advanced* primary spirit
  (only accessible through the advanced new game settings). See design
  discussion in issue #6.
* New advanced primary spirit: **Vampiric Bat**. It has a *vampirism*
  ability that allows to steal life by damaging foes: the Vampirism status
  effect is short-lived, so one has to chose foes wisely! The bat's attacks
  follow a two-phased attack pattern: when ranged, it works as a confusing
  rampaging attack. In melee, a plain attack is followed by a quick 2-tile
  retreat.
* New expansion mod: **Corrupted Dungeon**. The orb's influence has scattered
  twisted surprises through the dungeon! The mod is recommended to any player
  that wants more unpredictability and variability  during their adventures.
  See design discussion in issue #6.
* New expansion mod: **Advanced Spirits**. It introduces 7 new advanced
  secondary spirits with peculiar strong points but also serious and quirky
  drawbacks! See design discussion in issue #7.
* New challenge mod: **Small Inventory**. Inventory can only hold 3
  comestibles. Choose them with care!
* New challenge mod: **No Recharges**. Spirit ability charges are doubled but
  don’t recharge when going to the next map level. Use them wisely!
* New experimental challenge mod: **Healing Combat**. Healing happens through
  combat.  When a monster dies, you may heal for 1 HP, with higher chance at
  low HP.  However, comestibles don’t provide healing anymore. Vampiric Bat
  players start with an extra “vampirism” charge and get another one at map
  level 5, instead of healing on monster death. Less frequent healing but
  better tactical control over it! This experimental mod is playable but still
  lacks polish with respect to the default experience.

The **Advanced Spirits** expansion mod provides the following advanced
secondary spirits:

* **Dazzling Zebra**. It has a “disorient” ability that makes foes move in the
  opposite direction with respect to your current direction. The Zebra has a
  special dazzling passive effect on foes: it makes foes on one side mistakenly
  hit instead any foe just behind you on the other side. However, being so
  dazzling also means Zebra is easy to notice instantly, so beware of sudden
  attacks while exploring!
* **Gardening Lion**. It has a “garden” ability that allows to grow foliage
  around you for several turns. The Lion roars loudly on sighting foes for the
  first time, scaring them.
* **Gawalt Monkey**. It features wall-jumping (inspired from Harmonist) and a
  “shadows” ability that allows to move around unnoticed and go through
  translucent walls. Stealth and mobility are the Gawalt's way, and combat is
  more difficult, because attack damage is weakened. The Gawalt can sense
  menhirs and hide in them.
* **Gluttonous Bear**. It has a powerful “snack” ability that teleports you to
  the nearest comestible and makes you eat it. However, the Bear needs
  otherwise to eat comestibles in pairs!
* **Runic Chicken**. It has a “lay rune” ability that allows to lay random
  runes in places. The Chicken doesn't trigger traps, but cackles loudly on
  sighting totems and portals, as it of course also does when laying runes.
* **Staring Owl**. It has a “death stare” ability that allows to instantly kill
  a single foe but leaves you afraid due to its backlash. The Owl flies a bit
  over foliage, increasing visibility over it, but with the Owl comes the
  Night, reducing maximum view range.
* **Stomping Elephant**. It has a “stomp” ability that destroys nearby walls
  and makes you Berserk for a short while. The Elephant is immune to Berserk
  after-effects and can remain unnoticed in dead-ends, but it is afraid of rats
  and cannot turn easily when facing walls!

Reworked and improved content:

* Rework Lignification's balance following a discussion in issue #7.
  + Reduce Lignification duration from lignification fruit from 12 to 10 to
    make it a bit less risky. Now it's a +2 duration with respect to Berserk
    from the berserking flower, the same difference as the one between
    lignification and berserk runes.
  + Getting hit while lignified and afraid now makes you Berserk, like in
    cornered situations, unless you have “woody legs”. You still need to take
    an actual hit. Misses don't count.
  + Fire vulnerability from Lignification has been removed. Due to the high
    moisture levels coming from the ground, actors without “woody legs” now
    even get delayed burning from Lignification. Not being able to move is
    already a big enough vulnerability to Fire. Fire is still dangerous for
    lignified actors, as getting on fire means 3 total damage over 3 turns, but
    it should be more survivable now.
  + To compensate for the Fire-related buffs and preserve Fire+Lignification
    offensive tactics, HP floor for lignified monsters is now only 1, unlike
    for the player. That was in some ways an oversight, anyway, as an HP floor
    of 3 was quite bizarre for monsters, given how 3 HP is actually a lot for
    most of them, and sometimes higher than the maximum. Thematically, we may
    think that corrupted monsters don't benefit as much from nature's healing.
* Eating clarity leaves and foggy-skin onion while Clarity or Foggy-Skin is
  active now extends the beneficial status by the remaining duration, instead
  of losing remaining turns. Eating a berserking flower or lignification fruit
  renews the duration and temporary HP bonus, but doesn't extend the duration
  beyond the normal initial one (Lignification is already long enough anyway).
  Also, teleport mushroom now gives some healing when you're lignified instead
  of doing nothing, and firebreath pepper now extends Fire and Confusion by the
  remaining duration, so that using it on foes that already have those statuses
  still has some useful effect.
* Wind Fox's recoil has been reworked: guaranteed on hits in melee, 2/3 chance
  on hits at distance 2, and never beyond. Pushing gale's damage in melee is
  now 1-2, so often 2 damage instead of 1. The idea behind those changes is to
  make monsters more likely to reach the player, but help the player a bit once
  it is in melee, to make Wind Fox less extreme in making easy situations
  easier and hard situations harder.
* Monster fear is now somewhat biased toward going in the opposite direction
  from the player, instead of fleeing laterally, so that non-Hydra players get
  more offensive use from Fear.

UI, graphics, QoL:

* SDL2 version: new command-line options `-w` and `-h` to set window width and
  height scaling factors (see issue #8). Results may look ugly for non-integer
  factors. At the very least, using multiples of 0.25 seems better (some other
  values result in visual glitches).
* Improved “extra warnings” configuration option. It now also warns when some
  particular statuses like Berserk or Lignification are about to expire.  It
  doesn't warn anymore for Poison after Berserk, given you got warned on the
  previous turn.
* Added detailed descriptions for various menu entries.

Fixes:

* A translucent wall explosion now releases a poison cloud, without replacing
  it with a normal cloud. The cloud replacing wasn't very interesting and was
  an oversight.
* Fix bug that could make an afraid monster take a (non-attacking) step toward
  you.
* Fix bug where you could get Imbalance despite being lignified (see #6).
* Fix some edge-cases in earth menhir that prevented revealing some walls that
  should've been revealed.
* Fix small animation glitch with Clarity, where the previous last-seen
  position of a monster would be briefly shown with the wrong color (see #6).
* Fix off-by-one mistake for duration in description of lignification rune.
* Minor update in frog's catching description: it always unbalances when
  ranged, but description said "may".
* Fix translucent-wall related cases of auto-explore saying you couldn't
  explore the whole map, when you actually had explored all the terrain
  connected to passable tiles.
* Fix potential crash when trying to load an incompatible save (see #12).

# v1.1.0 2025-10-27

This release is a polishing one with some gameplay improvements and various
fixes.

Reworked and improved content:

* Improve portal placement so that it's less predictable: vaults are now more
  likely to contain a portal when they're far from the player's starting
  position, but the minimum distance is now lower. In particular, it's now not
  always as clear whether the portal will be on the left or on the right, and
  if there are two, they may appear on opposite sides.
* Improve pushing behavior in edge-cases when attacking a monster that has a
  monster behind: pushing can now happen if the monster behind dies, and if the
  monster in front dies, you can at least move to its position.
* Rework pushing for rampaging boars: unlike dragons, they now perform pushing
  only when charging from beyond a 2-tile range, but always do so unless they
  miss. Also, charging from beyond a 4-tile range results in double Imbalance
  duration (4 instead of 2). For consistency and simplicity, frog's catching
  unbalancing effect has been adjusted to match the boar's more closely: more
  deterministic chance, lower Imbalance duration for short ranges (but
  including 2-tile range), and double Imbalance duration beyond a 4-tile range.
  Note that earth dragons and Rampaging Boar players with Dig still perform
  pushing in melee as before. See discussion in issue #4 for some context
  around this change.
* Disable Cat swapping while lignified even with woody legs, as it's more a
  teleport-like kind of movement (see #4).
* Lignification now unbalances adjacent actors on expiration, too, due to
  withering roots messing with nearby ground.
* The lich is now immune to Confusion, and the undead knight is immune to Daze.
* Thunder Porcupine's “daze resistance” now allows spirit invocation while
  dazed, instead of reducing the duration (see discussion in issue #4).
* Adjust poison cloud duration from 6-9 to 7-11, making it have same average
  duration as fire clouds but with slightly smaller variability.
* Adjust duration of Berserk and Lignification from runic traps to 5 and 8
  respectively (see issue #4).

UI, graphics, QoL:

* Change color of monster's out of view location memory to light grey, so that
  it's clearer that it doesn't necessarily represent their current location
  (similar change suggested on r/roguelikes by u/femto42). However, monsters
  sensed with Clarity are shown with their true color.

Fixes:

* Don't leak mindstate information of previously seen out of view monsters.
* Fix a minor regression in a couple of animations (incorrect y-axis position).
* Afraid monsters can no longer ignore lignification when escaping (see issue
  #2).
* Prevent information leak: monster last-seen marker disappeared if it died out
  of view (see issue #2).
* Fix case of nadre ghost after explosion due to obstructing cloud (the player
  sees the explosion but because of the cloud you didn't see the death), and
  make clarity leaves properly update information about dead monsters (instead
  of showing in some cases a monster with 0 HP); see issue #5.
* Fix case of ghost footsteps (see issue #5).
* Fix confused butterfly self-blinking (see issue #4).
* Prevent possible redundant confusion-related message when a confused charging
  monster dies of poison while doing so.
* Nadres can no longer end up with 2/1 HP after lignification (see issue #4).
* When lignifying a poisoned monster, make induced confusion not hurt you for
  having used “lignify” if you weren't already confused (see issue #3).
* Catching a poisoned monster from afar no longer inflicts extra poison damage
  on it (as it's forced movement, like for pushing).

# v1.0.0 2025-09-18

New content:

* New secondary spirit: **Fire Salamander**. It features a *fire retreat*
  ability that makes the player retreat up to two tiles while leaving a fire
  cloud on the previous position, as long as there's room for the retreat. The
  new spirit also confers +1 attack and a chance of burning foes on hit (like
  phoenixes do) and the upgraded version confers fire resistance (delayed
  burning).  As a drawback, it inflicts bad hearing on the player, making it
  difficult to hear creep and wing-flapping noises. If the Fire Salamander is
  combined with the Barking Hound spirit, hearing becomes normal again (neither
  good nor bad). Fire Salamander is hence a spirit that offers very good
  offense capabilities and some mobility, but it can impair stealth tactics.
* New unique monster: **walking mushroom**. It's less sturdy than a walking
  tree but releases lignifying spores on sight from afar, affecting other
  monsters too. Also, its melee hits are confusing.
* New unique monster: **noisy imp**. It performs music just for the player,
  attracting nearby monsters. It avoids fighting similarly to afraid monsters.
* New special monster: **warping wraith**. Its hits teleport the player away.
  It may be sent by the dungeon to investigate and protect the portal after you
  use a warping menhir.
* New portal guardian monster: **blazing golem**. A sturdy magical golem that
  explodes on death. May occasionally guard non-portal vaults, too.
* New totem guardian monster: **totem wasp**. Rampaging and venomous wasps that
  like to use totems as their nest.
* New kind of static item: **runic trap**. They represent *visible* traps that
  are triggered when stepped on by the player or a monster. Wandering monsters
  avoid traps, but hunting ones do not. Effects depend on the kind of rune.
  Placement is designed so that there's always at least one path that avoids
  traps.
* New static **empty totem** item on map levels without a regular totem. Its
  purpose is to make it so the player doesn't feel the need to explore all the
  map corners on those maps (feature suggested by @mlochbaum in issue #1).
  Normal totems now become empty totems once you've picked their spirit.
* New Clarity status effect, provided by “clarity leaves”. It protects against
  Confusion, Daze, and unvoluntary Berserk. It can now be used as a way to exit
  Berserk early with extra HP (healing applied after bonus is removed). Also,
  Clarity makes you sense monsters now for several turns, but in a shorter
  range (2 times view range, instead of 3).
* New kind of player's knowledge about terrain: “unknown passable terrain”.
  Hearing footsteps and Clarity provide clues about terrain passability, so
  optimal play could encourage boringly memorizing (or even writing down) that
  information. Revealing the terrain and anything it contains would have solved
  the issue, but it would have been unintuitive from an immersion perspective.
  The new “unknown passable terrain”, shown as uncolored footsteps, solves the
  issue without having to leak any extra information.

Reworked and improved content:

* Make translucent walls liberate a poison cloud when destroyed. Thematically,
  they are walls that contain a special poisonous gas that makes them
  translucent and, on destruction, that gas is expelled, resulting in a poison
  cloud and ordinary rubble.
* Rework menhirs. The “mapping menhir” is now called “earth menhir” and only reveals
  the frontier formed by map walls adjacent to interior ones; its noisy echoing
  sound also destroys most nearby walls. Other menhirs remain similar, but
  offer some partial information about the map too: “poison menhirs” reveal
  translucent walls (which now contain poison), “fire menhirs” partially reveal
  foliage (that can be burnt), and “warping menhirs” (previously called
  “teleport menhirs”) reveal portals.
* When the player or a monster is both *afraid and cornered* with no possible
  escape, they now become berserk for a short while!
* A few small improvements in footstep hearing: more deterministic and better
  UI (use a different color for each monster noise class).
* Remove the Nausea status effect, because it was player-only and didn't have
  interesting interactions. It was mainly useful as a nerf for teleport: now
  teleport results in either Daze or Imbalance (chosen randomly between both,
  for some surprise and adaptability). The “disgusting rat” has been renamed
  into a “hungry rat” with a new “hunt by smell” trait that makes rats able to
  track you by smell from out of view and, as an extra trait, become berserk if
  they see you eat a comestible.
* Reworked Lignification and Foggy-Skin (see discussion in issue #1).
  + Lignification now only grants +2 defense (and an extra +1 against ranged
    attacks), but caps attack damage at 1.  This cap affects Hydra attacks, but
    not the extra effective damage bonuses from Dig or Berserk. And while
    monsters that ignore-defense will bypass the +2 defense bonus, the damage
    cap applies to them (so they will have a high hit chance, but will do 1
    damage). Monster walking trees have their defense reduced from 4 to 2, but
    get a “permanently lignified” trait that grants them the attack damage cap
    of Lignification permanently. Overall, the change implies less misses and
    more predictability against lignified foes, but somewhat lower average
    damage. Also, HP now never falls below 3 on bonus expiration, making
    Lignification partially preserve HP bonus in semi-critical situations and
    work as a +2 HP or +1 HP heal when under 3 HP.  The Imbalance debuff
    duration was adjusted from 6 to 5, too.
  + Foggy-Skin has been reworked to play better with Lignification: the healing
    now applies after removing Lignification's bonus and it's been made into +3
    HP (ensuring 4 HP after Lignification). To compensate, Foggy-Skin doesn't
    provide anymore a +1 defense bonus (which did not feel good even if
    actually not bad for such a long duration), and it doesn't protect against
    defense-ignoring attacks anymore.
* Remove all eating restrictions. For example, a “berserking flower” can now be
  eaten even if already Berserk: it will end Berserk early and apply it again.
  That means you can use it to safely prolong Berserk when only one turn is
  left, but you still get the Poison debuff.  Eating a “lignification fruit”
  while you have Foggy-Skin or eating a “berserking flower” while you have
  Clarity works, but it results in half Lignification or Berserk duration
  respectively.
* Make “ambrosia berries” either confuse you (existing behavior) or make you
  afraid. The idea, like for the “teleport mushroom”, is to provide some kind
  of surprise and require adaptability (a bit like identification systems do in
  some situations, but in a more restricted way). A fun interaction with the
  new afraid and cornered mechanic means it can be used to become berserk (with
  50% chance).
* Make “firebreath pepper” produce a nadre-like explosion when reaching a wall,
  providing a limited means of wall destruction for non-Boar players. Make
  explosions produce dust or smoke clouds. Also, make the breath's smell
  inflict a very brief (2 turns) Confusion effect, increasing the possible
  tactical usages.  The Daze removing effect has been moved into the “clarity
  leaves”.
* Unlike plain clouds of fog, dust or smoke, dangerous clouds of fire and
  poison are now only difficult to see through, like foliage (suggested by
  @mlochbaum in issue #1). This should make menhir usage obstruct vision less,
  as well as allow for things like jumping over a fire tile with the Frog or
  Gazelle spirits. It also allows ranged attacks to reach just behind a fire
  or poison cloud (which is both good and bad for different reasons).
* New rules for clouds replacing existing clouds of different types. If the
  type is the same, the duration is still the maximum of the old and new ones,
  like before. Fire clouds now take priority over other clouds: the heat
  expands the air and dissipates other kinds of clouds. This makes Foggy-Skin
  usable for protection during a foliage fire but without removing the fires
  (which has a visibility impact too). For other kinds of clouds, like poison
  ones or normal smoke/dust/fog, there are no special priorities: the new one
  replaces the old one.
* Adjust slightly down the duration of fire clouds to 6-12 turns and poison
  clouds to 6-9 turns. Adjust duration of Poison inflicted by clouds down to 3,
  and slightly reduce confusion duration from 5 to 4 to compensate. Also reduce
  the Skunk's “noxious smell” confusion duration from 7 to 5, because it was a
  bit too strong.
* Temporal Cat's space-distorting attacks now *swap clouds* too. While getting
  swapped into a cloud of fire or poison by foes was (very occasionally) fun,
  Temporal Cat players not being able to safely attack monsters on dangerous
  clouds was more annoying than fun. And in both cases, it could result in
  unexpected dangerous situations that would've required explicit examination
  to be avoided: when a monster is on a tile with a cloud, only the monster is
  drawn, and while that limitation could be solved for the tiles version with
  image layers, introducing that kind of visual clutter doesn't fit the
  simplicity of the game.
* Temporal Cat now has one less “stop time” charge, but the duration is
  increased from 4 to 5 (see discussion in #1). Also, to compensate having only
  1 charge initially and to make the upgrade feel less strong with respect to
  other spirits, the “dazing attack” effect is now active since the
  non-upgraded spirit.
* Wind Fox's “pushing gale” now dissipates clouds, providing the first cloud
  dissipating source which, due to “recoil”, can be quite useful for Wind Fox
  players. Also, as a special interaction, “pushing gale” propagates fire
  clouds on foliage tiles instead of dissipating them. Also, as discussed in
  issue #1 about Wind Fox balance, make “pushing gale” always do 1 damage
  (instead of 1-2 damage). And make “recoil” produce some wind noise in melee
  and generally only happen within a 4-tile range.
* Wind Fox's “recoil” is now active even with the non-upgraded Wind Fox
  (pushing is active on the non-upgraded Rampaging Boar, so having recoil only
  on upgraded Wind Fox seemed a bit strange).
* Four-Headed Hydra's attack bonus for four-directional attacks has been
  increased by one: that means fighting 2, 3 or 4 foes gives a +2, +3, or +4
  attack bonus now. Fighting a single foe remains the same as before, without
  bonus. The bonus was a bit small before, as fighting 2 monsters is the most
  common bonus case, either because there are only two, or because the third
  already died so, attack-wise, a boar fighting two foes separately was the
  same as a Hydra fighting two of them in the open, but with an initial charge
  bonus, and Hydra became weaker again if one of the foes died before the
  other.
* Rampaging Boar players don't move into poison or fire clouds after pushing
  anymore. While sometimes funny, it had similar issues as swapping into
  unexpected clouds. Monster boars and dragons still move into dangerous clouds
  after pushing, because they're enraged and corrupted, so not as smart!
* Rampaging Boar's “dig” ability now also guarantees pushing will happen,
  including the unbalancing effect for the upgraded spirit. Also, the +1 Attack
  base bonus and “unbalancing” upgrade bonus have been swapped to provide the
  more qualitative attack effect first (like for other spirits).
* Jumping Frog now has one extra jump charge initially (so 3 initially and then
  still 4 when upgraded), but the jump's daze duration on monsters is decreased
  by one from 5 to 4. The +1 Defense base bonus and “unbalancing” upgrade bonus
  have been swapped to provide the more qualitative attack effect first (like
  for other spirits).
* Venomous Viper's “poison cloud” ability now creates poison clouds *up to two
  free tiles away* (before a monster or obstacle). This creates more varied
  possibilities for tactical movement, while still restricting safe movements
  in a similar but less extreme way, and providing similar defense from ranged
  monsters.
* Thunder Porcupine's lightning has its daze duration increased from 3 to 4,
  now on the same level as the Jumping Frog's “dazing jump”.
* Walking Tree's “fire vulnerability” and Lignification increase Fire duration
  by 2 instead of 3.
* Sprinting Gazelle now leaves dust behind only on floor tiles (in particular
  not on foliage ones). Also, Sprint's duration has been adjusted down from 7
  to 6.
* Using Sprinting Gazelle's “sprint” or Jumping Frog's “jump” abilities while
  imbalanced may now sometimes lead to a dazing fall.
* The “sprint” and “focus” abilities now cancel each other. It's never really
  useful nor thematic to have both at the same time, so hydra players have now
  an instant sprint cancellation option (at the cost of a spirit charge).
* While sprinting, the Jumping Frog's “jump” ability now unbalances foes
  without hurting them, but still dazes them, effectively combining the effects
  of the two kinds of jumps in a sensible way.
* HP of earth dragons has been decreased from 6 to 5, but along with vipers,
  nadres and hydras, they now have “scales” that provide extra +1 defense
  against ranged attacks.
* Add simple boredom clock: if you spend too much time on a single level (more
  than necessary with normal play), the dungeon core will detect your presence
  and mark you until you go through a portal.

Monster behavior and distribution improvements:

* Improve a bit monster distribution. Mostly, bias toward early-mid-late
  monsters depending on map level has been increased, making the various phases
  more distinct. Early and Mid game difficulty remains similar. Late game has
  now usually less monsters than before, but a bit more of dangerous ones, so
  it has become more stealth-friendly while still being challenging enough,
  except in case of a special “swarm” level, favoring greater numbers over
  individual monster dangerousness.
* Improve wandering monster destination selection. It's still quite random with
  some bias toward close locations, but there's some extra bias toward
  special vaults, so that interesting places are visited more often by
  monsters.
* Improvements in monster group formation. In particular, less groups of three
  monsters than before and more groups of two, so somewhat better coverage of
  the map in mid and late game, but smaller likeliness of large monster groups.
* Make the order in which monsters act within a turn after the player properly
  random. It already felt that way, but a very attentive player could sometimes
  guess that some monster seemed to always act before some other, potentially
  leading to gamey predictions.

Map generation:

* Mixed map generation algorithms in some map levels and some other smaller
  improvements in map generation.
* Malfunctioning portals are now more common (around 1-3 per game, with 2 being
  the most common). Some minor adjustments in menhir distribution (in
  particular, the same type can't appear twice on a given map level anymore).
* Improved placement of non-vault comestibles, now more often present in
  spots next to walls than before (suggested by @mlochbaum).
* Add a few new special vaults.

UI, graphics, QoL:

* Add concise information about current tile in the status bar which tells
  whether there's fog, an item, or a special kind of terrain under the player.
* Animation improvements and fixes. Monster deaths are now always animated when
  in view, independently of the cause. Hits between confused monsters are now
  animated too, making it clearer what's happening. Status bar and logs are now
  progressively updated during chained animations, instead of only at the end
  of the sequence.  Other minor enhancements.
* Improve various tiles (footsteps, translucent walls, food…).
* Wizard mode improvements: intended for development purposes, but now cheat
  usage is recorded in a more fine-grained way in statistics, so any frustrated
  casual players might consider using it as an *adventure mode*. Enabling
  Wizard Mode *once* with `W` only enables non-permadeath mode (each
  resurrection by the wizard will be recorded in the timeline). Extra cheats
  and visibility modes are enabled by pressing `W` again one or more times.
* Refine drawing of tiles with both a cloud and an item: clouds are now drawn
  over items, but if the cloud is not dangerous, the color of the item is used
  (which never conflicts with the color of dangerous clouds).
* A few improvements in game statistic's dump: ASCII map now shows monster
  noise too, the number of last messages was increased a bit.
* Various help improvements.
* In native builds, rename game statistic's dump and logs files to `dump.txt`
  and `logs.txt`, so that it's clearer that those files are text, unlike the
  other ones.

Fixes:

* Nadre explosions now properly blow up nearby walls too, like in Boohu and
  Harmonist.
* Make the player never appear on a vault with a totem: not really a bug, but
  it wasn't very fun.
* Fix field of view sometimes not being updated just after cloud creation but
  later at the end of a turn, which could allow monsters to attack once from a
  distance before the update.
* Fix spelling and formatting issues in a few log messages. Avoid case of
  duplicated “Saw *monster*” entries in timeline.
* Fix healing amount in description of foggy-skin onion.
* Fix application of Imbalance when jumping with the Sprinting Gazelle's sprint.
* Fix vipers biting themselves too even when not confused (issue #1).
* When killing a nadre, make the explosion burn you just once even if you
  didn't move in that turn (see issue #1).
* Some confused monster actions that happened outside of view incorrectly
  appeared in the game logs.
* Improve/fix the page-down and scrolling instructions when details of all
  entities in a single examined tile don't fit.

# v1.0.0-beta.1 2025-07-22

First public release of Shamogu.
