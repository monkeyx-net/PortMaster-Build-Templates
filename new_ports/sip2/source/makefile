
LDLIBS = -lSDL2 -lSDL2_mixer -lm -lGL
CXXFLAGS = -g -Wall -std=c++2a
SRC = spaceinvaders.cpp pixiretro.cpp main.cpp
INC = spaceinvaders.h pixiretro.h

si : $(SRC) $(INC)
	$(CXX) $(CXXFLAGS) -o $@ $(SRC) $(LDLIBS)

.PHONY: clean
clean:
	rm si *.o
