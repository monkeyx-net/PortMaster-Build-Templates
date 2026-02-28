#include "pixiretro.h"
#include "spaceinvaders.h"

pxr::Engine* engine {nullptr};

int main()
{
  engine = new pxr::Engine{};
  engine->initialize(std::move(std::unique_ptr<pxr::Application>{new SpaceInvaders{}}));
  engine->run();
  delete engine;
}
