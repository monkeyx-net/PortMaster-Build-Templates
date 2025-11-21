#!/usr/bin/env perl
# Usage: perl make-fonts.pl [new|all]
# Defaults to only updating new tiles.

use strict;
use warnings;
use v5.18;
use utf8;
use open qw(:std :utf8);

use Imager;

my $font = Imager::Font->new(
    file => '/usr/local/share/fonts/adobe-source-code-pro/SourceCodePro-Regular.ttf',
    size => 23,
    aa => 0,
    color => "white",
);

my $fontbold = Imager::Font->new(
    file => '/usr/local/share/fonts/adobe-source-code-pro/SourceCodePro-Bold.ttf',
    size => 23,
    aa => 0,
    color => "white",
);

my %names = (
    " " => "space",
    "!" => "totem",
    "#" => "wall",
    "%" => "food",
    "&" => "menhir",
    "'" => "quote",
    "(" => "lparen",
    ")" => "rparen",
    "*" => "asterisc",
    "," => "comma",
    "-" => "hyphen",
    "." => "ground",
    "/" => "slash",
    ":" => "colon",
    ";" => "semicolon",
    "<" => "less",
    "?" => "interrogation",
    "@" => "player",
    "[" => "lbracket",
    "\\" => "backslash",
    "]" => "rbracket",
    "^" => "rubble",
    "{" => "lbrace",
    "|" => "pipe",
    "}" => "rbrace",
    "~" => "tilde",
    "¤" => "frontier",
    "§" => "cloud",
    "«" => "ldiag",
    "»" => "rdiag",
    "×" => "times",
    "Φ" => "magic",
    "—" => "longhyphen",
    "•" => "tick",
    "…" => "dots",
    "←" => "larrow",
    "↑" => "uarrow",
    "→" => "rarrow",
    "↓" => "darrow",
    "√" => "hit",
    "∞" => "kill",
    "─" => "hbar",
    "│" => "vbar",
    "┌" => "boxnw",
    "┐" => "boxne",
    "└" => "boxsw",
    "┘" => "boxse",
    "┤" => "boxe",
    "◊" => "glass",
    "☼" => "orb",
    "♣" => "tree",
    "♪" => "sound",
    "♫" => "footsteps",
    '"' => "quotes",
    '+' => "plus",
    '=' => "rune",
    '>' => "portal",
    '’' => "apostrophe",
    '“' => "lquotes",
    '”' => "rquotes",
);

my @letters = (
    qw( a b c d e f g h i j k l m n o p q r s t u v w x y z
        A B C D E F G H I J K L M N O P Q R S T U V W X Y Z
        1 2 3 4 5 6 7 8 9 0),
    keys %names
);

my $cmd = shift @ARGV // "new";

if ($cmd ne "new" and $cmd ne "all") {
    die "$0: unknown command argument: $cmd";
}

my $p = "../tiles";
for my $l (@letters) {
    my $name = $names{$l} // $l;
    if ($cmd eq "new" and -f "$p/letter-$name.png") {
        next;
    }
    my $img = Imager->new(xsize => 16, ysize => 24);
    $img->box(filled => 1, color => 'black');
    if ($l eq "#" or $l eq "◊") {
        $img->string(font => $fontbold, text => $l, x => 1, y => 18);
    } else {
        $img->string(font => $font, text => $l, x => 1, y => 18);
    }
    $img->write(file => "$p/letter-$name.png")
        or die "Cannot save letter-$name", $img->errstr;
}
