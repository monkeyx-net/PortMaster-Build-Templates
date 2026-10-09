/*
  ufo-above.pov

  By Bill Kendrick
  bill@newbreedsoftware.com
  http://www.newbreedsoftware.com/bill/

  May 14, 2000 - May 15, 2000
*/

#declare SPIN = clock;
#declare GLOW = 2;


/* Colors and sun location: */

#declare SUNX = 5;
#declare SUNY = -1;
#declare SUNZ = 0;

#declare BKGD_R = 0;
#declare BKGD_G = 0;
#declare BKGD_B = 0;


/* Background: */

background { color <BKGD_R, BKGD_G, BKGD_B> }


/* Camera: */

camera {
  location <0, -1.25, 0>
  up    <0,1,0>
  right  <1,0,0>
  look_at <0, 0, 0>
}


/* Ambient light source: */

global_settings {
  ambient_light rgb <.5, .5, .5>
}


/* Sun: */

light_source {
  <SUNX, SUNY, SUNZ>
  color rgb <1, 1, 0.8>
}


/* UFO: */

union {

  /* Disc: */

  difference {
    sphere {
      <0, 0, 0>,
      .5
      scale <1, .3, 1>

      pigment { color <1, 1, 1> }
      finish {
        ambient .25
        irid {
          .15
          thickness 1
          turbulence .25
        }
      }
    }

    box {
      <-.5, .1, -.5>,
      <.5, .15, .5>
      pigment { color <1, 0, 1> }

      finish {
        ambient GLOW
        diffuse .5
        reflection .5
        specular 1
        roughness .001

        irid {
          0.35
          thickness 1.5
          turbulence .5
        } 

      }
    }
  }


  /* Cockpit: */

  sphere {
    <0, -.25, 0>,
    .25

    pigment { color <0, 1, 1> }

    finish {
      ambient 0.5
      diffuse .1
      reflection .1
      specular 1
    }

    scale <1, .5, 1>
  }


  /* Colored lights: */

  /* (Front left) */

  sphere {
    <.3, -.0675, -.3>,
    .05

    pigment { color <1, 0, 0> }
    finish {ambient 2 diffuse .6}
  }


  /* (Front right) */

  sphere {
    <.3, -.0675, .3>,
    .05

    pigment { color <0, 0, 1> }
    finish {ambient 2 diffuse .6}
  }

  /* (Back left) */

  sphere {
    <-.3, -.0675, -.3>,
    .05

    pigment { color <0, 1, 0> }
    finish {ambient 2 diffuse .6}
  }


  /* (Back right) */

  sphere {
    <-.3, -.0675, .3>,
    .05

    pigment { color <1, 1, 0> }
    finish {ambient 2 diffuse .6}
  }

  rotate <0, SPIN, 0>
  rotate <0, 0, -TILT>
  rotate <70, 0, 0>
}
