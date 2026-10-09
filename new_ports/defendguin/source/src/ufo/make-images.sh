#!/bin/bash

# Renders image of UFO ships for Defendguin
# originally for un-developed UFO game in 2000;
# utilized for Defendguin, 2019 & 2023

# by Bill Kendrick
# bill@newbreedsoftware.com
# http://www.newbreedsoftware.com/bill/

# May 14, 2000 - June 28, 2023


VIEW_ARGS="+D +W32 +H32"
BOSSVIEW_ARGS="+D +W128 +H128"
RENDER_ARGS="+A0.001"
SPIN_ARGS="+KFI0 +KFF11 +KI0 +KF330"


rm -f tmp.pov
echo "#declare TILT=10;" > tmp.pov
cat ufo.pov >> tmp.pov
povray $VIEW_ARGS $RENDER_ARGS $SPIN_ARGS +FP +O../../data/images/ufo/ufo_lft.ppm +Itmp.pov

rm -f tmp.pov
echo "#declare TILT=0;" > tmp.pov
cat ufo.pov >> tmp.pov
povray $VIEW_ARGS $RENDER_ARGS $SPIN_ARGS +FP +O../../data/images/ufo/ufo_ctr.ppm +Itmp.pov

rm -f tmp.pov
echo "#declare TILT=-10;" > tmp.pov
cat ufo.pov >> tmp.pov
povray $VIEW_ARGS $RENDER_ARGS $SPIN_ARGS +FP +O../../data/images/ufo/ufo_rgt.ppm +Itmp.pov

rm -f tmp.pov
echo "#declare TILT=0;" > tmp.pov
cat ufo.pov >> tmp.pov
povray $BOSSVIEW_ARGS $RENDER_ARGS $SPIN_ARGS +FP +O../../data/images/boss/boss.ppm +Itmp.pov

rm -f tmp.pov

echo "Now convert from PPM to BMP!"

