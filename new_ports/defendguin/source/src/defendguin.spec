# RPM .spec file for Defendguin
#
# by Bill Kendrick
# bill@newbreedsoftware.com
# http://www.newbreedsoftware.com/defendguin/

# June 6, 2000 - January 29, 2006

Name: defendguin
Summary: A bidirectionally scrolling space game based on Defender.
Version: 0.0.11
Release: 1
Copyright: GPL
Group: Games/Arcade
Source: ftp://ftp.billsgames.com/unix/x/defendguin/defendguin-0.0.11.tar.gz
URL: http://www.newbreedsoftware.com/defendguin/
Packager: Bill Kendrick <bill@newbreedsoftware.com>

AutoReqProv: no
Requires: SDL >= 1.1
Requires: SDL_mixer >= 1.0

%description
A certain monopoly-owning bad guy has been cloned hundreds of times by
an unknown alien race.  They are now attacking earth, kidnapping little
penguinoids and converting them into mutants.  Helping them on their way
are some other nasty alien ships, of which there are plenty.

%prep
%setup

%build
make

%install
make install

%files
%doc AUTHORS.txt
%doc CHANGES.txt
%doc COPYING.txt
%doc INSTALL.txt
%doc README.txt
%doc TODO.txt
%doc defendguin.lsm
/usr/local/share/defendguin/
/usr/local/bin/defendguin
/usr/local/man/man6/defendguin.6
