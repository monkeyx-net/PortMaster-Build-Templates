package main

import (
	"fmt"
	"math/rand/v2"
)

// Abs returns the absolute value.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// RandInt is like rand.IntN but returns zero for non-positive integers.
func RandInt(n int) int {
	if n <= 0 {
		return 0
	}
	x := rand.IntN(n)
	return x
}

// Indefinite returns the given text prefixed by an indefinite article
// (uppercased if desired).
func Indefinite(s string, upper bool) (text string) {
	if len(s) > 0 {
		switch s[0] {
		case 'a', 'e', 'i', 'o', 'u':
			if upper {
				text = "An " + s
			} else {
				text = "an " + s
			}
		default:
			if upper {
				text = "A " + s
			} else {
				text = "a " + s
			}
		}
	}
	return text
}

// Countable adds an "s" depending on quantity n.
func Countable(s string, n int) string {
	if len(s) == 0 || s[len(s)-1] == 's' || n == 1 {
		return s
	}
	return s + "s"
}

// CountNoun adds a counter and an "s" depending on quantity n.
func CountNoun(s string, n int) string {
	return fmt.Sprintf("%d %s", n, Countable(s, n))
}
