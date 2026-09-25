// This file exposes compile-time configuration constants or variables that can
// be edited by hand before building the game. The defaults should be good
// enough for most players, but if you play Shamogu a lot and are unsatisfied
// with some aspects of it, you might want to have a look at the options below.
//
// The variables (var) may also be temporarily changed for a given play session
// through command-line options (in native versions only). For example:
//
//     shamogu -O NoAnim
//
// Any extra options should be comma separated. Options can be preceded by an
// exclamation mark "!" to negate them (in case you set "true" as default
// before building the game).

package main

// NoAnim disables animations.
var NoAnim = false

// NoChaos changes chaos megabat to a screeching megabat with no chaos bite,
// and removes rare comestibles from the game, except conditional comestibles
// when Totem Conditions is enabled.
var NoChaos = false

// GluttonyRework makes the Gluttonous Bear from Advanced Spirits eat in pairs
// through a choice-of-two menu, without relying on a Gluttony status.
const GluttonyRework = false

// ResetMods makes it so starting a new game resets mod selection, instead of
// restoring the saved selection from the previous session.
const ResetMods = false

// VampiricHC implicitly sets Healing Combat when the Vampiric Bat is chosen,
// irrespectively of mod selection.
const VampiricHC = false
