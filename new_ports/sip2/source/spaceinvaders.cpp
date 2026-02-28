#include "pixiretro.h"
#include "spaceinvaders.h"

//===============================================================================================//
// ##>SPACE INVADERS                                                                             //
//===============================================================================================//

bool SpaceInvaders::initialize(Engine* engine, int32_t windowWidth, int32_t windowHeight)
{
  Application::initialize(engine, windowWidth, windowHeight);

  _worldScale = 1;
  while((baseWorldSize._x * _worldScale) < windowWidth && (baseWorldSize._y * _worldScale) < windowHeight)
    ++_worldScale; 

  --_worldScale;
  if(_worldScale == 0)
    _worldScale = 1;

  _worldSize = baseWorldSize * _worldScale;

  Application::onWindowResize(windowWidth, windowHeight);

  Assets::Manifest_t manifest{};
  for(int32_t i = BMK_CANNON0; i < BMK_COUNT; ++i)
    manifest.push_back({i, _bitmapNames[i], _worldScale}); 

  pxr::assets->loadBitmaps(manifest);

  manifest.clear();
  manifest.push_back({fontKey, fontName, _worldScale});
  pxr::assets->loadFonts(manifest);

  Mixer::Manifest_t mixmanifest{};
  for(int32_t i = SK_EXPLOSION; i < SK_COUNT; ++i)
    mixmanifest.push_back({i, _soundNames[i]});

  loadHiScores();
  updateHudHiScore();

  pxr::mixer->loadSoundsWAV(mixmanifest);

  _isHudVisible = false;
  _hud.initialize(&(pxr::assets->getFont(fontKey, _worldScale)), flashPeriod, phasePeriod);
  _uidScoreText = _hud.addTextLabel({Vector2i{10, 240} * _worldScale, pxr::colors::magenta, "SCORE"});
  _uidScoreValue = _hud.addIntLabel({Vector2i{10, 230} * _worldScale, pxr::colors::white, &_score, 5});
  _uidHiScoreText = _hud.addTextLabel({Vector2i{85, 240} * _worldScale, pxr::colors::red, "HI-SCORE"});
  _uidHiScoreValue = _hud.addIntLabel({Vector2i{95, 230} * _worldScale, pxr::colors::green, &_hiscore, 5});
  _uidRoundText = _hud.addTextLabel({Vector2i{170, 240} * _worldScale, pxr::colors::yellow, "ROUND"});
  _uidRoundValue = _hud.addIntLabel({Vector2i{170, 230} * _worldScale, pxr::colors::magenta, &_round, 5});
  _uidCreditText = _hud.addTextLabel({Vector2i{130, 6} * _worldScale, pxr::colors::magenta, "CREDIT"});
  _uidCreditValue = _hud.addIntLabel({Vector2i{190, 6} * _worldScale, pxr::colors::cyan, &_credit, 1});
  _uidLivesValue = _hud.addIntLabel({Vector2i{10, 6} * _worldScale, pxr::colors::yellow, &_lives, 1});

  for(int i{0}; i < maxLivesHudCannons; ++i){
    _uidLivesBitmaps[i] = _hud.addBitmapLabel({
      Vector2i{(20 + (16 * i)), 6} * _worldScale, 
      pxr::colors::green, 
      &(assets->getBitmap(SpaceInvaders::BMK_CANNON0, _worldScale))
    });
  }

  resetGameStats();

  std::unique_ptr<ApplicationState> game = std::make_unique<GameState>(this);
  std::unique_ptr<ApplicationState> menu = std::make_unique<MenuState>(this);
  std::unique_ptr<ApplicationState> splash = std::make_unique<SplashState>(this);
  std::unique_ptr<ApplicationState> scoreReg = std::make_unique<HiScoreRegState>(this);
  std::unique_ptr<ApplicationState> scoreBoard = std::make_unique<HiScoreBoardState>(this);
  std::unique_ptr<ApplicationState> sos = std::make_unique<SosState>(this);

  game->initialize(_worldSize, _worldScale);
  menu->initialize(_worldSize, _worldScale);
  splash->initialize(_worldSize, _worldScale);
  scoreReg->initialize(_worldSize, _worldScale);
  scoreBoard->initialize(_worldSize, _worldScale);
  sos->initialize(_worldSize, _worldScale);

  // a bodge! dont look! its really ugly! :) ... your still looking! I warned you, you'll regret it!
  static_cast<SosState*>(sos.get())->_gameState = static_cast<GameState*>(game.get());

  addState(std::move(game));
  addState(std::move(menu));
  addState(std::move(splash));
  addState(std::move(scoreReg));
  addState(std::move(scoreBoard));
  addState(std::move(sos));

  switchState(SplashState::name);

  return true;
}

void SpaceInvaders::onUpdate(double now, float dt)
{
  Application::onUpdate(now, dt);
  _hud.onUpdate(dt);
}

void SpaceInvaders::onDraw(double now, float dt)
{
  Application::onDraw(now ,dt);

  if(_isHudVisible && !isWindowTooSmall())
    _hud.onDraw();
}

void SpaceInvaders::hideTopHud()
{
  _hud.hideTextLabel(_uidScoreText);
  _hud.hideIntLabel(_uidScoreValue);
  _hud.hideTextLabel(_uidHiScoreText);
  _hud.hideIntLabel(_uidHiScoreValue);
  _hud.hideTextLabel(_uidRoundText);
  _hud.hideIntLabel(_uidRoundValue);
}

void SpaceInvaders::showTopHud()
{
  _hud.showTextLabel(_uidScoreText);
  _hud.showIntLabel(_uidScoreValue);
  _hud.showTextLabel(_uidHiScoreText);
  _hud.showIntLabel(_uidHiScoreValue);
  _hud.showTextLabel(_uidRoundText);
  _hud.showIntLabel(_uidRoundValue);
}

void SpaceInvaders::hideLivesHud()
{
  _isLivesHudVisible = false;
  _hud.hideIntLabel(_uidLivesValue);
  for(auto uid : _uidLivesBitmaps)
    _hud.hideBitmapLabel(uid);
}

void SpaceInvaders::showLivesHud()
{
  _isLivesHudVisible = true;
  _hud.showIntLabel(_uidLivesValue);
  for(int life{1}; life < maxPlayerLives; ++life)
    if(life < _lives)
      _hud.showBitmapLabel(_uidLivesBitmaps[life - 1]);
}

void SpaceInvaders::startScoreHudFlash()
{
  _hud.startIntLabelFlash(_uidScoreValue);
}

void SpaceInvaders::stopScoreHudFlash()
{
  _hud.stopIntLabelFlash(_uidScoreValue);
}

void SpaceInvaders::setLives(int32_t lives)
{
  _lives = std::max(0, lives);
  updateLivesHud();
}

void SpaceInvaders::addLives(int32_t lives)
{
  _lives += lives;
  _lives = std::max(0, _lives);
  updateLivesHud();
}

void SpaceInvaders::updateLivesHud()
{
  if(!_isLivesHudVisible)
    return;

  for(int life{1}; life < maxPlayerLives; ++life){
    if(life < _lives)
      _hud.showBitmapLabel(_uidLivesBitmaps[life - 1]);
    else
      _hud.hideBitmapLabel(_uidLivesBitmaps[life - 1]);
  }
}

void SpaceInvaders::resetGameStats()
{
  setLives(maxPlayerLives);
  clearPlayerName();
  _round = 0;
  _credit = 0;
  _score = 0;
}

void SpaceInvaders::loadHiScores()
{
  ScoreData data {}; 
  if(data.load(ScoreData::filename) != 0)
    data.write(ScoreData::filename, false);

  for(int i{0}; i < hiscoreCount; ++i){
    int nameKey = ScoreData::NAME0 + (i * 2);
    int scoreKey = ScoreData::SCORE0 + (i * 2);
    int iname = data.getIntValue(nameKey);
    int score = data.getIntValue(scoreKey);
    _hiscores[i]._name = intToName(iname);
    _hiscores[i]._value = score;
  }

  auto scoreCompare = [](const Score& s0, const Score& s1) -> bool {
    return s0._value < s1._value;
  };

  std::sort(_hiscores.begin(), _hiscores.end(), scoreCompare);
}

void SpaceInvaders::writeHiScores()
{
  ScoreData data {}; 

  for(int i{0}; i < hiscoreCount; ++i){
    int nameKey = ScoreData::NAME0 + (i * 2);
    int scoreKey = ScoreData::SCORE0 + (i * 2);
    int iname = nameToInt(_hiscores[i]._name);
    data.setIntValue(nameKey, iname);
    data.setIntValue(scoreKey, _hiscores[i]._value);
  }

  data.write(ScoreData::filename, false);
}

bool SpaceInvaders::isHiScore(int32_t scoreValue)
{
  return scoreValue > _hiscores[0]._value;  
}

bool SpaceInvaders::isDuplicateHiScore(const Score& score)
{
  for(auto& s : _hiscores)
    if((s._name == score._name) && (s._value == score._value))
      return true;
  return false;
}

bool SpaceInvaders::registerHiScore(const Score& score)
{
  size_t position = findScoreBoardPosition(score._value);
  if(position == -1) return false;
  std::shift_left(_hiscores.begin(), _hiscores.begin() + position + 1, 1); 
  _hiscores[position] = score;
  updateHudHiScore();
  return true;
}

size_t SpaceInvaders::findScoreBoardPosition(int32_t scoreValue)
{
  if(scoreValue < _hiscores.front()._value) 
    return -1;

  if(scoreValue > _hiscores.back()._value)
    return hiscoreCount - 1;

  for(int i{0}; i < hiscoreCount - 1; ++i)
    if(_hiscores[i]._value < scoreValue && scoreValue <= _hiscores[i + 1]._value)
      return i;
}

void SpaceInvaders::updateHudHiScore()
{
  _hiscore = _hiscores.back()._value;
}

//===============================================================================================//
// ##>SPLASH STATE                                                                               //
//===============================================================================================//

SplashState::SplashState(Application* app) :
  ApplicationState{app}
{}

void SplashState::initialize(Vector2i worldSize, int32_t worldScale)
{
  _worldSize = worldSize;
  _worldScale = worldScale;

  _masterClock = 0.f;
  _nextNode = 0;
  _sequence = {{
    {1.f, EVENT_SHOW_SPACE_SIGN},
    {2.f, EVENT_TRIGGER_SPACE_SIGN},
    {4.5f, EVENT_SHOW_INVADERS_SIGN},
    {5.5f, EVENT_TRIGGER_INVADERS_SIGN},
    {7.3f, EVENT_SHOW_PART_II},
    {8.3f, EVENT_SHOW_HUD},
    {11.f, EVENT_END},
  }};

  _blockSize = 3 * _worldScale;
  _blockSpace = 1 * _worldScale;
  _signX = 16 * _worldScale;
  _spaceY = 192 * _worldScale;
  _invadersY = 112 * _worldScale;

  _spaceTriggered = false;
  _spaceVisible = false;
  _spaceSign = std::make_unique<Sign<spaceW, spaceH>>(Sign<spaceW, spaceH>{
    {{
      {1,2,2,2,2,1,1,1,1,1,2,2,2,2,2,2,2,1,1,1,1,1,2,2,2,2,2,1,1,1,1,1,1,2,2,2,2,1,1,1,1,2,2,2,2,2,2,2},
      {2,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2,2,1,1,1,1,2,2,2,2,2,1,1,1,1,1,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2},
      {2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2},
      {2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2},
      {2,2,1,1,1,2,2,2,1,1,2,2,2,1,1,1,2,2,2,1,2,2,2,2,2,2,2,2,2,1,1,2,2,2,1,1,2,2,2,1,2,2,2,1,1,1,1,1},
      {2,2,2,1,1,2,2,2,1,1,2,2,2,1,1,1,2,2,2,1,2,2,2,2,2,2,2,2,2,1,1,2,2,2,1,1,2,2,2,1,2,2,2,1,1,1,1,1},
      {1,2,2,2,1,1,1,1,1,1,2,2,2,1,1,1,2,2,2,1,2,2,2,1,2,1,2,2,2,1,1,2,2,2,1,1,1,1,1,1,2,2,2,1,1,1,1,1},
      {1,2,2,2,2,2,1,1,1,1,2,2,2,2,2,2,2,2,2,1,2,2,2,1,2,1,2,2,2,1,1,2,2,2,1,1,1,1,1,1,2,2,2,2,2,1,1,1},
      {1,1,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,2,2,1,2,2,2,1,1,1,1,1,1,2,2,2,2,2,2,1,1,1},
      {1,1,1,1,2,2,2,2,2,1,1,2,2,2,2,2,2,1,1,1,1,1,2,2,2,2,2,1,1,1,2,2,2,1,1,1,1,1,1,2,2,2,2,2,2,1,1,1},
      {1,1,1,1,1,1,2,2,2,1,1,2,2,2,1,1,1,1,1,1,1,2,2,2,1,2,2,2,1,1,2,2,2,1,1,1,1,1,1,2,2,1,1,1,1,1,1,1},
      {1,1,1,2,2,1,1,2,2,1,1,2,2,2,1,1,1,1,1,1,1,2,2,2,1,2,2,2,1,1,2,2,2,1,2,2,2,1,1,2,2,1,1,1,1,1,1,1},
      {1,1,1,2,2,1,1,2,2,1,1,2,2,2,1,1,1,1,1,1,2,2,2,1,1,1,2,2,2,1,2,2,2,1,2,2,2,1,2,2,2,1,1,1,1,1,1,1},
      {1,1,1,2,2,2,2,2,2,1,1,2,2,2,1,1,1,1,1,1,2,2,2,1,1,1,2,2,2,1,2,2,2,2,2,2,2,1,2,2,2,2,2,2,1,1,1,1},
      {1,1,1,1,2,2,2,2,1,1,1,2,2,2,1,1,1,1,1,1,1,2,2,2,1,2,2,2,1,1,1,2,2,2,2,2,1,1,2,2,2,2,2,2,1,1,1,1},
      {1,1,1,1,1,2,2,1,1,1,1,2,2,2,1,1,1,1,1,1,1,2,2,2,1,2,2,2,1,1,1,1,2,2,2,1,1,1,2,2,2,2,2,2,1,1,1,1},
    }},
    std::make_unique<Bitmap>(pxr::assets->makeBlockBitmap(_blockSize, _blockSize)),
    {_signX, _spaceY},
    pxr::colors::green,
    pxr::colors::cyan,
    0.002f,
    _blockSpace,
    _blockSize
  });

  _invadersTriggered = false;
  _invadersVisible = false;
  _invadersSign = std::make_unique<Sign<invadersW, invadersH>>(Sign<invadersW, invadersH>{
    {{
      {1,2,2,1,2,2,1,1,2,2,1,2,2,1,1,2,2,1,1,2,2,2,1,1,2,2,2,2,1,1,2,2,2,2,2,1,2,2,2,2,1,1,1,2,2,2,1,1},
      {1,2,2,1,2,2,1,1,2,2,1,2,2,1,1,2,2,1,1,2,2,2,1,1,2,2,1,2,2,1,2,2,1,1,1,1,2,2,1,2,2,1,2,2,1,2,2,1},
      {1,2,2,1,2,2,2,1,2,2,1,2,2,1,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,1,1,1,2,2,1,2,2,1,2,2,1,1,1,1},
      {1,2,2,1,2,2,2,2,2,2,1,2,2,1,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,1,1,1,2,2,1,2,2,1,2,2,1,1,1,1},
      {1,2,2,1,2,2,2,2,2,2,1,1,2,2,2,2,1,1,2,2,2,2,2,1,2,2,1,2,2,1,2,2,2,2,2,1,2,2,2,2,1,1,1,2,2,2,1,1},
      {1,2,2,1,2,2,1,2,2,2,1,1,2,2,2,2,1,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,1,1,1,2,2,1,2,2,1,1,1,1,2,2,1},
      {1,2,2,1,2,2,1,1,2,2,1,1,1,2,2,1,1,1,2,2,1,2,2,1,2,2,1,2,2,1,2,2,1,1,1,1,2,2,1,2,2,1,2,2,1,2,2,1},
      {1,2,2,1,2,2,1,1,2,2,1,1,1,2,2,1,1,1,2,2,1,2,2,1,2,2,2,2,1,1,2,2,2,2,2,1,2,2,1,2,2,1,1,2,2,2,1,1},
    }},
    std::make_unique<Bitmap>(pxr::assets->makeBlockBitmap(_blockSize, _blockSize)),
    {_signX, _invadersY},
    pxr::colors::magenta,
    pxr::colors::yellow,
    0.002f,
    _blockSpace,
    _blockSize
  });

  _partiiPosition = Vector2i{80, 48} * _worldScale;
  _partiiColor = pxr::colors::red;
  _partiiVisible = false;

  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  si->hideHud();
}

void SplashState::doEvents()
{
  if(_sequence[_nextNode]._time > _masterClock) 
    return;
  switch(_sequence[_nextNode]._event){
    case EVENT_SHOW_SPACE_SIGN:
      _spaceVisible = true;
      break;
    case EVENT_TRIGGER_SPACE_SIGN:
      _spaceTriggered = true;
      break;
    case EVENT_SHOW_INVADERS_SIGN:
      _invadersVisible = true;
      break;
    case EVENT_TRIGGER_INVADERS_SIGN:
      _invadersTriggered = true;
      break;
    case EVENT_SHOW_PART_II:
      _partiiVisible = true;
      break;
    case EVENT_SHOW_HUD:
      {
      SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
      HUD& hud = si->getHud();
      std::string text {"*Remake by ianmurfinxyz*"};
      _uidAuthor = hud.addTextLabel({Vector2i{32, 24} * _worldScale, pxr::colors::cyan, text});
      si->showHud();
      si->hideLivesHud();
      break;
      }
    case EVENT_END:
      {
      SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
      HUD& hud = si->getHud();
      hud.removeTextLabel(_uidAuthor);
      si->switchState(MenuState::name);
      break;
      }
    default:
      break;
  }
  ++_nextNode;
}

void SplashState::onEnter()
{
  static_cast<SpaceInvaders*>(_app)->hideHud();
  _masterClock = 0.f;
  _nextNode = 0;
  (*_spaceSign).reset();
  (*_invadersSign).reset();
  _spaceTriggered = false;
  _spaceVisible = false;
  _invadersTriggered = false;
  _invadersVisible = false;
  _partiiVisible = false;
}

void SplashState::onUpdate(double now, float dt)
{
  _masterClock += dt;
  doEvents();
  if(_spaceTriggered)
    _spaceSign->updateBlocks(dt);
  if(_invadersTriggered)
    _invadersSign->updateBlocks(dt);
}

void SplashState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);
  if(_spaceVisible)
    _spaceSign->draw();
  if(_invadersVisible)
    _invadersSign->draw();
  if(_partiiVisible){
    pxr::renderer->blitBitmap(_partiiPosition, 
                                pxr::assets->getBitmap(SpaceInvaders::BMK_PARTII, _worldScale), 
                                _partiiColor);
  }
}

//===============================================================================================//
// ##>GAME STATE                                                                                 //
//===============================================================================================//

GameState::BeatBox::BeatBox(std::array<Mixer::Key_t, beatCount> beats, float beatFreq_hz) : 
  _beats{beats},
  _nextBeat{0},
  _beatFreq_hz{beatFreq_hz},
  _beatPeriod_s{1.f / beatFreq_hz},
  _beatClock_s{0.f},
  _isPaused{false}
{}

void GameState::BeatBox::doBeats(float dt)
{
  if(_isPaused) return;
  _beatClock_s += dt;
  if(_beatClock_s > _beatPeriod_s){
    mixer->playSound(_beats[_nextBeat]);
    _nextBeat = pxr::wrap(_nextBeat + 1, 0, beatCount - 1);
    _beatClock_s = 0.f;
  }
}

void GameState::BeatBox::setBeatFreq(float freq_hz)
{
  _beatFreq_hz = freq_hz;
  _beatPeriod_s = 1 / _beatFreq_hz;
}

GameState::GameState(Application* app) : 
  ApplicationState{app}
{}

void GameState::initialize(Vector2i worldSize, int32_t worldScale)
{
  // This function 'hard-codes' all game data in one place so it is easy to find.

  _beatBox = BeatBox{{
      SpaceInvaders::SK_FAST1, 
      SpaceInvaders::SK_FAST2, 
      SpaceInvaders::SK_FAST3, 
      SpaceInvaders::SK_FAST4
    }, 
    2.f
  };

  _worldSize = worldSize;
  _worldScale = worldScale;

  _font = &(pxr::assets->getFont(SpaceInvaders::fontKey, _worldScale));
  _hud = &(static_cast<SpaceInvaders*>(_app)->getHud());

  _colorPalette = {     // index:
    colors::red,        // 0
    colors::green,      // 1
    colors::blue,       // 2
    colors::magenta,    // 3
    colors::cyan,       // 4
    colors::yellow,     // 5
    colors::white       // 6
  };

  //
  // Alien drop math:
  //
  // World size is 256px tall, aliens drop from a max height of 186px to a min of 32px. The
  // min is the invasion row which is the row the cannon sits on.
  //
  // Each drop is 14px, so the row heights are,
  //   186, 172, 158, 144, 130, 116, 102, 88, 74, 60, 46, 32   | row height
  //    0    1    2    3    4    5    6    7   8   9  10  11   | row num
  //                                        invasion row--^
  //
  // note that the top row of aliens starts at 186px (row 0), thus the bottom row starts at
  // 130 (row 4).
  //

  _alienShiftDisplacement = Vector2i{2, 0} * _worldScale;
  _alienDropDisplacement = Vector2i{0, -baseAlienDropDisplacement} * _worldScale;
  _alienSpawnDropDisplacement = Vector2i{0, -baseAlienSpawnDropDisplacement} * _worldScale;
  _alienInvasionHeight = baseAlienInvasionRowHeight * _worldScale;

  _alienXSeperation = baseAlienDropDisplacement * _worldScale;
  _alienYSeperation = baseAlienDropDisplacement * _worldScale;

  _aliensSpawnPosition._x = (_worldSize._x - (gridWidth * _alienXSeperation)) / 2;
  _aliensSpawnPosition._y = ((baseAlienTopRowHeight + (minSpawnDrops * baseAlienSpawnDropDisplacement))
                             - ((gridHeight - 1) * baseAlienDropDisplacement)) * _worldScale; 

  //_worldSize._y - (gridHeight * _alienYSeperation) - 30;

  _lastClassAlive = AlienClassId::SQUID;

  _worldMargin = 5 * _worldScale;

  _worldLeftBorderX = _worldMargin;
  _worldRightBorderX = _worldSize._x - _worldMargin;
  _worldTopBorderY = _worldSize._y - (30.f * _worldScale);

  _alienBoomDuration = 0.1f;
  _alienMorphDuration = 0.2f;

  // Each update tick the game performs a number of 'beats'. The beat rate controls the speed of
  // the game; speed of alien shifts and firing etc. Beats are composed into sets called cycles 
  // where each element of the set (cycle) represents a number of beats to perform in an update
  // tick. For example a cycle = {1, 2, end} means perform 1 beat in update tick N, 2 beats in
  // update tick N+1, and repeat, so 1 beat in N+2 etc. Since the engine guarantees an update
  // rate of 60Hz (only slower if the CPU is too slow), the cycle controls the game speed. Yes
  // game speed is tied to update rate, but the update rate is independent of hardware.

  // Theses cycles will produce exponentially increasing beat rates resulting in exponentially
  // increasing alien speed and firing.
  //
  //                          ticks to      fleet moves    
  //                          move fleet    each sec
  //                          ----------    -----------
  _cycles = {{
    {1,  cycleEnd, 0, 0},  // ticks:55.00   freq:1.09
    {1,  1, 2, cycleEnd},  // ticks:42.00   freq:1.43
    {1,  2, cycleEnd, 0},  // ticks:37.00   freq:1.60
    {2,  cycleEnd, 0, 0},  // ticks:27.50   freq:2.18
    {2,  3, cycleEnd, 0},  // ticks:22.00   freq:2.70
    {5,  cycleEnd, 0, 0},  // ticks:18.33   freq:3.33
    {7,  cycleEnd, 0, 0},  // ticks:13.75   freq:4.35
    {10, cycleEnd, 0, 0},  // ticks:11.00   freq:5.56
    {14, cycleEnd, 0, 0},  // ticks:9.17    freq:6.67
    {19, cycleEnd, 0, 0},  // ticks:7.86    freq:7.69
    {25, cycleEnd, 0, 0},  // ticks:6.88    freq:9.09
    {34, cycleEnd, 0, 0},  // ticks:6.11    freq:9.81
    {46, cycleEnd, 0, 0}   // ticks:5.00    freq:12.00
  }};

  // The alien population that triggers the cycle.
  // note: the last element MUST == 0, else a segfault will happen when we select the next cycle.
  _cycleTransitions = {49, 42, 35, 28, 21, 14, 10, 7, 5, 4, 3, 2, 0};

  _alienClasses = {{
    {8  * _worldScale, 8 * _worldScale, 30, 1, {SpaceInvaders::BMK_SQUID0    , SpaceInvaders::BMK_SQUID1    }},
    {11 * _worldScale, 8 * _worldScale, 20, 4, {SpaceInvaders::BMK_CRAB0     , SpaceInvaders::BMK_CRAB1     }},
    {12 * _worldScale, 8 * _worldScale, 10, 3, {SpaceInvaders::BMK_OCTOPUS0  , SpaceInvaders::BMK_OCTOPUS1  }},
    {8  * _worldScale, 8 * _worldScale, 30, 5, {SpaceInvaders::BMK_CUTTLE0   , SpaceInvaders::BMK_CUTTLE1   }},
    {19 * _worldScale, 8 * _worldScale, 60, 5, {SpaceInvaders::BMK_CUTTLETWIN, SpaceInvaders::BMK_CUTTLETWIN}}
  }};

  _ufoClasses = {{
    {{50, 100, 150}, 300, 16 * _worldScale, 7 * _worldScale, 0, 0.f, false, SpaceInvaders::BMK_SAUCER, SpaceInvaders::BMK_UFOBOOM},
    {{300, 350, 400}, 1000, 15 * _worldScale, 7 * _worldScale, 3, schrodingerPhasePeriodSeconds, true, SpaceInvaders::BMK_SCHRODINGER, SpaceInvaders::BMK_UFOBOOM}
  }};

  _ufoSpawnY = 210.f * _worldScale;
  _ufoSpeed = 40.f * _worldScale;     // to move fully across the world and off screen.
  _ufoBoomScoreDuration = 0.5f;
  _ufoPhaseDuration = 0.8f;

  _formations = {{  // note formations look inverted here as array[0] is the bottom row, but they are not.
    // formation 0
    {{
       {OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS},
       {OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS},
       {CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   },
       {CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   },
       {SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  }
    }},
    // formation 1
    {{
       {SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  ,SQUID  },
       {CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   },
       {CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   ,CRAB   },
       {OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS},
       {OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS,OCTOPUS}
    }}
  }};

  _bombIntervalDeviation = 0.5f; // Maximum 50% deviation from base.
  _bombIntervals = {80, 80, 100, 120, 140, 180, 240, 300, 400, 500, 650, 800, 1100};

  _bombClasses = {{
    {3 * _worldScale, 6 * _worldScale, -80.f * _worldScale, 0, 20, 0, {SpaceInvaders::BMK_CROSS0, SpaceInvaders::BMK_CROSS1, SpaceInvaders::BMK_CROSS2, SpaceInvaders::BMK_CROSS3}},
    {3 * _worldScale, 7 * _worldScale, -120.f * _worldScale, 4, 20, 10, {SpaceInvaders::BMK_ZIGZAG0, SpaceInvaders::BMK_ZIGZAG1, SpaceInvaders::BMK_ZIGZAG2, SpaceInvaders::BMK_ZIGZAG3}},
    {3 * _worldScale, 7 * _worldScale, -100.f * _worldScale, 5, 20, 4, {SpaceInvaders::BMK_ZAGZIG0, SpaceInvaders::BMK_ZAGZIG1, SpaceInvaders::BMK_ZAGZIG2, SpaceInvaders::BMK_ZAGZIG3}}
  }};

  _bombBoomKeys = {{SpaceInvaders::BMK_BOMBBOOMBOTTOM, SpaceInvaders::BMK_BOMBBOOMMIDAIR}};
  _bombBoomWidth = 8 * _worldScale;
  _bombBoomHeight = 8 * _worldScale;
  _bombBoomDuration = 0.4f;

  _cannon._spawnPosition = Vector2f{_worldLeftBorderX, baseAlienInvasionRowHeight * _worldScale};
  _cannon._speed = 50.f * _worldScale;
  _cannon._width = 13 * _worldScale;
  _cannon._height = 8 * _worldScale;
  _cannon._boomFrameDuration = 0.2f;
  _cannon._boomDuration = _cannon._boomFrameDuration * 5;
  _cannon._cannonKey = SpaceInvaders::BMK_CANNON0;
  _cannon._boomKeys = {{SpaceInvaders::BMK_CANNONBOOM0, SpaceInvaders::BMK_CANNONBOOM1, SpaceInvaders::BMK_CANNONBOOM2}};
  _cannon._colorIndex = 0;

  _laser._width = 1 * _worldScale;
  _laser._height = 6 * _worldScale;
  _laser._colorIndex = 6;
  _laser._speed = 300.f * _worldScale;
  _laser._bitmapKey = SpaceInvaders::BMK_LASER0;

  _bunkerColorIndex = 0;
  _bunkerSpawnX = 32 * _worldScale;
  _bunkerSpawnY = 48 * _worldScale;
  _bunkerSpawnGapX = 45 * _worldScale;
  _bunkerSpawnCount = 4;
  _bunkerWidth = 22 * _worldScale;
  _bunkerHeight = 16 * _worldScale;
  _bunkerDeleteThreshold = 20 * _worldScale;

  _levels = {{
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
    {minSpawnDrops, 0, true, true},
  }};

  _levelIndex = -1;
  _isGameOver = false;
}

void GameState::startNextLevel()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);

  int round = si->getRound();

  // TODO if round > levelCount, pick a random level from the top 5 levels or something

  _levelIndex = round;
  if(_levelIndex >= levelCount)
    _levelIndex = levelCount - 1;

  _activeCycle = spawnCycle;
  _activeBeat = cycleStart;
  _beatBox.pause();
  _nextMover = {0, 0};
  _alienMoveDirection = 1;
  _dropsDone = 0;
  _isAliensBooming = false;
  _isAliensMorphing = false;
  _isAliensDropping = true;
  _isAliensSpawning = true;
  _isAliensFrozen = false;
  _isAliensAboveInvasionRow = false;

  // Reset aliens.
  for(int32_t col = 0; col < gridWidth; ++col){
    for(int32_t row = 0; row < gridHeight; ++row){
      Alien& alien = _grid[row][col];

      alien._classId = _formations[0][row][col];

      alien._position._x = _aliensSpawnPosition._x + (col * _alienXSeperation);
      alien._position._y = _aliensSpawnPosition._y + (row * _alienYSeperation);

      alien._row = row;
      alien._col = col;

      alien._frame = false;
      alien._isAlive = true;
    }
  }

  std::fill(_columnPops.begin(), _columnPops.end(), gridHeight);
  std::fill(_rowPops.begin(), _rowPops.end(), gridWidth);
  _alienPopulation = gridSize;

  _isUfoBooming = false;
  _isUfoScoring = false;
  _ufo._isAlive = false;
  _tillUfo = pxr::randUniformSignedInt(tillUfoMin, tillUfoMax);
  _ufoCounter = 0;
  _canUfosSpawn = true;

  // Reset bombs.
  for(auto& bomb : _bombs)
    bomb._isAlive = false;

  _bombCount = 0;

  // Reset bomb booms.
  for(auto& boom : _bombBooms)
    boom._isAlive = false;

  _laser._isAlive = false;
  _shotCounter = 0;

  // Reset player cannon.
  _cannon._isBooming = false;
  _cannon._isAlive = false;

  _bombClock = _bombIntervals[_activeCycle];

  // Create fresh (undamaged) hitbar.
  _hitbar = std::make_unique<Hitbar>(
      pxr::assets->getBitmap(SpaceInvaders::BMK_HITBAR, _worldScale), 
      _worldSize._x,
      1 * _worldScale,
      16 * _worldScale, 
      1
  );

  // Create fresh bunkers.
  _bunkers.clear();
  Vector2f position {_bunkerSpawnX, _bunkerSpawnY};
  for(int i = 0; i < _bunkerSpawnCount; ++i){
    spawnBunker(position, SpaceInvaders::BMK_BUNKER);
    position._x += _bunkerSpawnGapX;
  }

  _isGameOver = false;
  _isVictory = false;
  startRoundIntro();

  si->showHud();
  si->hideTopHud();
  si->showLivesHud();
}

void GameState::updateBeatFreq()
{
  int32_t ticksPerCycle {0};
  int32_t beatsPerCycle {0};
  for(auto beats : _cycles[_activeCycle]){
    if(beats != cycleEnd){
      ++ticksPerCycle;
      beatsPerCycle += beats;
    }
    else
      break;
  }

  // example: If a cycle does 5 beats in 3 ticks, then 5 aliens will move in 3 ticks, so
  // 55 aliens will move in (55/5) * 3 = 33 ticks. If it takes 33 ticks to move the fleet,
  // and we do 60 ticks per second, then the frequency of full fleet movements will be
  // 60 / 33 = 1.81hz.
  //
  // hence note that the fequency of the beating equals the frequency of full fleet movements.
  //
  //                    v-- the fixed frame rate set in the engine.
  float beatFreq_hz = (60.f / ((gridSize / beatsPerCycle) * ticksPerCycle)) * beatFreqScale;
  //                                                  base rate a little intense --^

  _beatBox.setBeatFreq(beatFreq_hz);
}

void GameState::updateActiveCycle()
{
  _activeCycle = 0;
  while(_alienPopulation < _cycleTransitions[_activeCycle]){
    ++_activeCycle;

    // should never happen unless we have set, in _cycleTransitions, for the last cycle
    // to trigger at a population greater than 0.
    assert(_activeCycle < cycleCount);

    updateBeatFreq();
  }
}

void GameState::updateActiveCycleBeat()
{
  ++_activeBeat;
  if(_cycles[_activeCycle][_activeBeat] == cycleEnd)
    _activeBeat = cycleStart;
}

void GameState::endSpawning()
{
  _isAliensSpawning = false;
  _isAliensDropping = false;
  spawnCannon(false);
  _activeCycle = 0;
  updateBeatFreq();
  _beatBox.unpause();

  static_cast<SpaceInvaders*>(_app)->showTopHud();
}

void GameState::spawnCannon(bool takeLife)
{
  if(takeLife){
    SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
    si->addLives(-1);
    if(si->getLives() <= 0){
      startGameOver();
      return;
    }
  }

  _cannon._position = _cannon._spawnPosition;
  _cannon._moveDirection = 0;
  _cannon._isBooming = false;
  _cannon._isAlive = true;
  _isAliensFrozen = false;
  _beatBox.unpause();
}

void GameState::spawnBoom(Vector2i position, BombHit hit, int32_t colorIndex)
{
  auto search = std::find_if_not(_bombBooms.begin(), _bombBooms.end(), isBombBoomAlive);

  // If this asserts then expand the booms array until it does not.
  assert(search != _bombBooms.end());

  (*search)._hit = hit;
  (*search)._colorIndex = colorIndex;
  (*search)._position = position;
  (*search)._boomClock = _bombBoomDuration;
  (*search)._isAlive = true;
}

void GameState::spawnBomb(Vector2f position, BombClassId classId)
{
  // If this condition does occur then increase the max bombs until it doesn't.
  assert(_bombCount != maxBombs);

  // Find a 'dead' bomb instance to use.
  Bomb* bomb {nullptr};
  for(auto& b : _bombs)
    if(!b._isAlive)
      bomb = &b;

  // If this occurs my bomb counts are off.
  assert(bomb != nullptr);

  bomb->_classId = classId;
  bomb->_position = position;
  bomb->_frame = 0;
  bomb->_isAlive = true;
  bomb->_frameClock = _bombClasses[classId]._frameInterval;

  ++_bombCount;
}

void GameState::spawnBunker(Vector2f position, Assets::Key_t bitmapKey)
{
  const Bitmap& bitmap = pxr::assets->getBitmap(bitmapKey, _worldScale);
  _bunkers.emplace_back(std::make_unique<Bunker>(bitmap, position));
}

void GameState::spawnUfo(UfoClassId classId)
{
  _ufo._classId = classId;
  _ufo._isAlive = true;
  _ufoDirection = (randUniformSignedInt(0, 1) == 0) ? 1 : -1;
  _ufo._position._x = (_ufoDirection == 1) ? 0 : _worldSize._x;
  _ufo._position._y = _ufoSpawnY;
  _ufo._phase = true;
  _ufoSfxChannel = mixer->playSound(SpaceInvaders::SK_UFO_HIGH_PITCH, 1000);
}

void GameState::morphAlien(Alien& alien)
  // predicate: alien->_col != gridWidth - 1
{
  static_cast<SpaceInvaders*>(_app)->addScore(_alienClasses[alien._classId]._scoreValue);

  alien._classId = CUTTLETWIN;
  _alienMorpher = &alien;
  _alienMorphClock = _alienMorphDuration;
  _isAliensFrozen = true;
  _isAliensMorphing = true;

  Alien& neighbour = _grid[alien._row][alien._col + 1];
  if(neighbour._isAlive){
    neighbour._isAlive = false;
    --(_columnPops[neighbour._col]);
    --(_rowPops[neighbour._row]);
    --_alienPopulation;
  }

  mixer->playSound(SpaceInvaders::SK_INVADER_MORPHED);
}

void GameState::boomCannon()
{
  _cannon._moveDirection = 0;
  _cannon._boomClock = _cannon._boomDuration;
  _cannon._boomFrame = 0;
  _cannon._boomFrameClock = _cannon._boomFrameDuration;
  _cannon._isBooming = true;
  _cannon._isAlive = false;

  _isAliensFrozen = true;

  _beatBox.pause();

  mixer->playSound(SpaceInvaders::SK_EXPLOSION);
}

void GameState::boomBomb(Bomb& bomb, bool makeBoom, Vector2i boomPosition, BombHit hit)
{
  --_bombCount;
  bomb._isAlive = false;

  if(makeBoom)
    spawnBoom(boomPosition, hit, _bombClasses[bomb._classId]._colorIndex);
}

void GameState::boomUfo()
{
  _ufo._isAlive = false;
  _isUfoBooming = true;
  _ufoBoomScoreClock = 0.f;

  const UfoClass& uc = _ufoClasses[_ufo._classId];

  if((_ufoCounter == 1 && _shotCounter == 23) ||
     (_ufoCounter > 1  && (_shotCounter - 23) % 15 == 0))
    _ufoLastScoreGiven = uc._specialScoreValue;
  else
    _ufoLastScoreGiven = uc._randScoreValues[randUniformSignedInt(0, UfoClass::randScoreValueCount - 1)];

  static_cast<SpaceInvaders*>(_app)->addScore(_ufoLastScoreGiven);

  mixer->stopChannel(_ufoSfxChannel);
  _ufoSfxChannel = mixer->playSound(SpaceInvaders::SK_UFO_LOW_PITCH);
}

void GameState::boomAlien(Alien& alien)
{
  static_cast<SpaceInvaders*>(_app)->addScore(_alienClasses[alien._classId]._scoreValue);

  alien._isAlive = false;
  _alienBoomer = &alien;
  _alienBoomClock = _alienBoomDuration;
  _isAliensFrozen = true;
  _isAliensBooming = true;
  --(_columnPops[alien._col]);
  --(_rowPops[alien._row]);
  --_alienPopulation;
  if(_alienPopulation <= 0)
    _lastClassAlive = alien._classId;
  if(_alienPopulation <= 8)
    _canUfosSpawn = false;
  updateActiveCycle();

  mixer->playSound(SpaceInvaders::SK_INVADER_KILLED);
}

void GameState::boomLaser(bool makeBoom, BombHit hit)
{
  _laser._isAlive = false;

  if(makeBoom){
    Vector2i position {};
    position._x = _laser._position._x - ((_bombBoomWidth - _laser._width) / 2);
    position._y = _laser._position._y;

    spawnBoom(position, hit, _laser._colorIndex);
  }
}

void GameState::boomBunker(Bunker& bunker, Vector2i pixelHit)
{
  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap& aBitmap {bunker._bitmap};
  const Bitmap& bBitmap {pxr::assets->getBitmap(SpaceInvaders::BMK_BOMBBOOMMIDAIR, _worldScale)};

  aPosition._x = bunker._position._x;
  aPosition._y = bunker._position._y;

  bPosition._x = bunker._position._x + pixelHit._x - (_bombBoomWidth / 2);
  bPosition._y = bunker._position._y + pixelHit._y - (_bombBoomHeight / 2);

  const Collision& c = testCollision(aPosition, aBitmap, bPosition, bBitmap, true);

  // If this asserts then the mask used to blit off damage is not intersecting any pixels
  // on the bitmap, the result is that no change is made. This creates the situation in which
  // parts of the bunker cannot be destroyed. Thus if this happens redesign the damage mask.
  assert(c._isCollision);

  for(auto& pixel : c._aPixels)
    bunker._bitmap.setBit(pixel._y, pixel._x, 0, false);

  bunker._bitmap.regenerateBytes();
}


void GameState::doUfoSpawning()
{
  if(_ufo._isAlive)
    return;

  if(_isAliensSpawning || _isAliensDropping)
    return;

  if(_isGameOver || _isVictory || _isRoundIntro)
    return;

  if(_alienPopulation <= 8)
    return;

  --_tillUfo;
  if(_tillUfo <= 0){
    if(pxr::randUniformSignedInt(0, schrodingerSpawnChance) == 0)
     spawnUfo(SCHRODINGER);
    else
      spawnUfo(SAUCER);
    _tillUfo = pxr::randUniformSignedInt(tillUfoMin, tillUfoMax);
  }
}

void GameState::doAlienMorphing(float dt)
  // predicate: _alienMorpher->_col != gridWidth - 1
{
  if(!_isAliensMorphing)
    return;

  if(_alienPopulation == 0)
    return;

  _alienMorphClock -= dt;
  if(_alienMorphClock > 0)
    return;

  _alienMorpher->_classId = CUTTLE;

  Alien& neighbour = _grid[_alienMorpher->_row][_alienMorpher->_col + 1];
  neighbour._classId = CUTTLE;
  neighbour._isAlive = true;
  ++(_columnPops[neighbour._col]);
  ++(_rowPops[neighbour._row]);
  ++_alienPopulation;

  _isAliensMorphing = false;
  _isAliensFrozen = false;
  _alienMorpher = nullptr;
}

void GameState::doCannonMoving(float dt)
{
  if(!_cannon._isAlive)
    return;

  if(_isVictory)
    return;

  bool lKey = pxr::input->isKeyDown(Input::KEY_LEFT);
  bool rKey = pxr::input->isKeyDown(Input::KEY_RIGHT);
  if(lKey && !rKey){
    _cannon._moveDirection = -1;
  }
  else if(!lKey && rKey){
    _cannon._moveDirection = 1;
  }
  else{
    _cannon._moveDirection = 0;
  }

  _cannon._position._x += _cannon._speed * _cannon._moveDirection * dt;
  _cannon._position._x = std::clamp(
      _cannon._position._x, 
      static_cast<float>(_worldLeftBorderX), 
      static_cast<float>(_worldRightBorderX - _cannon._width)
  );
}

void GameState::doCannonBooming(float dt)
{
  if(!_cannon._isBooming)
    return;

  _cannon._boomClock -= dt;
  if(_cannon._boomClock <= 0){
    _cannon._isBooming = false;
    spawnCannon(true);
    return;
  }

  _cannon._boomFrameClock -= dt;
  if(_cannon._boomFrameClock <= 0){
    _cannon._boomFrame = pxr::wrap(++_cannon._boomFrame, 0, cannonBoomFramesCount - 1);
    _cannon._boomFrameClock = _cannon._boomFrameDuration;
  }
}

void GameState::doCannonFiring()
{
  if(!_cannon._isAlive)
    return;

  if(_laser._isAlive)
    return;

  if(_isVictory)
    return;

  if(pxr::input->isKeyDown(Input::KEY_SPACE)){
    Vector2f position = _cannon._position;
    position._x += _cannon._width / 2;
    position._y += _cannon._height;
    _laser._position = position;
    _laser._isAlive = true;
    mixer->playSound(SpaceInvaders::SK_SHOOT);
    ++_shotCounter;
  }
}

void GameState::doAlienMoving(int32_t beats)
{
  // Aliens move at fixed displacements independent of time, thus alien movement speed is an 
  // emergent property of the rate of update ticks, and importantly, the number of aliens moved 
  // in each tick. Note the engine guarantees a tick rate of 60Hz thus alien speed in game is 
  // controlled by the second factor; the number of aliens moved in each tick. Since the game is 
  // ticked at 60Hz, if a single alien is moved in each tick then 55 aliens will be moved in 55 
  // ticks so in 55/60 seconds. Moving 2 aliens per tick will result in twice the speed. 
  //
  // The number of aliens moved each update tick is equal to the number of beats performed in that
  // tick (see cycles note in spaceinvaders.h). Thus a cycle such as {1, 2, end} will mean update 
  // 1 alien in tick N, 2 aliens in tick N + 1, and repeat. This results in (when you do the math) 
  // all 55 aliens moving in 37 ticks, i.e. it takes 37/60 seconds to complete one full grid 
  // movement, giving a frequency of grid movements of 60/37 Hz.
  //
  // note: this design makes alien movement speed dependent on the number of aliens in the grid. 
  // The more aliens the more beats are required to complete one grid movement and each grid 
  // movement results in a fixed grid displacement. Thus if you change the number of aliens you 
  // must also change all cycles to tune the gameplay.

  if(_isRoundIntro)
    return;

  if(_isAliensFrozen)
    return;

  if(_alienPopulation == 0)
    return;

  for(int i = 0; i < beats; ++i){
    Alien& alien = _grid[_nextMover._row][_nextMover._col];

    if(_isAliensDropping){
      alien._position += _isAliensSpawning ? _alienSpawnDropDisplacement : _alienDropDisplacement;
    }
    else{
      alien._position += _alienShiftDisplacement * _alienMoveDirection;
    }

    alien._frame = !alien._frame;

    bool looped = incrementGridIndex(_nextMover);

    if(looped){
      if(_isAliensDropping){
        ++_dropsDone;

        if(_isAliensSpawning){
          mixer->playSound(SpaceInvaders::SK_FAST4);
          if(_dropsDone >= _levels[_levelIndex]._spawnDrops){
            endSpawning();
          }
        }
        else{
          _isAliensDropping = false;
          _alienMoveDirection *= -1;
        }
      }
      else if(doCollisionsAliensBorders()){
        _isAliensDropping = true;
      }
    }
  }
}

void GameState::doBombMoving(int32_t beats, float dt)
{
  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;

    const BombClass& bombClass = _bombClasses[bomb._classId];

    bomb._position._y += bombClass._speed * dt;

    bomb._frameClock -= beats;
    if(bomb._frameClock <= 0){
      bomb._frame = pxr::wrap(++bomb._frame, 0, bombFramesCount - 1);
      bomb._frameClock = bombClass._frameInterval;
    }
  }
}

void GameState::doLaserMoving(float dt)
{
  if(!_laser._isAlive)
    return;

  _laser._position._y += _laser._speed * dt;
}

void GameState::doUfoMoving(float dt)
{
  if(!_ufo._isAlive)
    return;

  _ufo._position._x += _ufoSpeed * _ufoDirection * dt;
}

void GameState::doUfoPhasing(float dt)
{
  if(!_ufo._isAlive) 
    return;

  const UfoClass& uc = _ufoClasses[_ufo._classId];

  if(!uc._isPhaser)
    return;

  _ufo._phaseClockSeconds += dt;
  if(_ufo._phaseClockSeconds >= uc._phasePeriodSeconds){
    _ufo._phase = !_ufo._phase;
    _ufo._phaseClockSeconds = 0.f;
  }
}

void GameState::doAlienBombing(int32_t beats)
{
  // Cycles determine alien bomb rate. Aliens bomb every N beats, thus the higher beat rate
  // the higher the rate of bombing. Randomness is added in a random deviation to the bomb 
  // interval and to the choice of alien which does the bombing.

  if(_isAliensFrozen)
    return;

  if(_isAliensSpawning)
    return;

  if(_alienPopulation == 0)
    return;
  
  _bombClock -= beats;
  if(_bombClock > 0)
    return;

  // Select the column to bomb from, taking into account unpopulated columns.
  int32_t populatedCount = gridWidth - std::count(_columnPops.begin(), _columnPops.end(), 0);

  // This condition should of already been detected as a level win.
  assert(populatedCount != 0);

  int32_t colShift = randUniformSignedInt(1, gridWidth) % populatedCount;
  int32_t col {-1};
  do {
    ++col;
    while(_columnPops[col] == 0) 
      ++col;
  }
  while(--colShift >= 0);

  // Find the alien that will do the bombing.
  Alien* alien {nullptr};
  for(int32_t row = 0; row < gridHeight; ++row){
    if(_grid[row][col]._isAlive){
      alien = &_grid[row][col];
      break;
    }
  }

  assert(alien != nullptr); // The column selection should ensure this never happens.
  
  const AlienClass& alienClass = _alienClasses[alien->_classId];

  BombClassId classId = static_cast<BombClassId>(randUniformSignedInt(CROSS, ZAGZIG));
  const BombClass& bombClass = _bombClasses[classId]; 

  Vector2f position {};
  position._x += alien->_position._x + (alienClass._width * 0.5f);
  position._y += alien->_position._y - bombClass._height;

  spawnBomb(position, classId);

  _bombClock = _bombIntervals[_activeCycle];
}

void GameState::doAlienBooming(float dt)
{
  if(!_isAliensBooming)
    return;

  _alienBoomClock -= dt;
  if(_alienBoomClock <= 0.f){
    _alienBoomer = nullptr;
    _isAliensFrozen = false;
    _isAliensBooming = false;
  }
}

void GameState::doUfoBoomScoring(float dt)
{
  if(!(_isUfoBooming || _isUfoScoring))
    return;

  _ufoBoomScoreClock += dt;
  if(_ufoBoomScoreClock >= _ufoBoomScoreDuration){
    if(_isUfoBooming){
      const UfoClass& uc = _ufoClasses[_ufo._classId];
      _uidUfoScoringText = _hud->addTextLabel({
        Vector2i(_ufo._position._x, _ufo._position._y),
        _colorPalette[uc._colorIndex],
        std::to_string(_ufoLastScoreGiven)
      });
      _isUfoBooming = false;
      _isUfoScoring = true;
    }
    else{
      _hud->removeTextLabel(_uidUfoScoringText);
      _isUfoScoring = false;
    }
    _ufoBoomScoreClock = 0.f;
  }
}

void GameState::doBombBoomBooming(float dt)
{
  for(auto& boom : _bombBooms){
    if(!boom._isAlive)
      continue;

    boom._boomClock -= dt;
    if(boom._boomClock <= 0.f)
      boom._isAlive = false;
  }
}

void GameState::doUfoReinforcing(float dt)
{
  // TODO
}

void GameState::doCollisionsUfoBorders()
{
  if(!_ufo._isAlive)
    return;

  if((_ufoDirection == -1 && _ufo._position._x < 0) ||
     (_ufoDirection == 1  && _ufo._position._x > _worldSize._x))
  {
    _ufo._isAlive = false;
    mixer->stopChannel(_ufoSfxChannel);
  }
}

void GameState::doCollisionsBombsHitbar()
{
  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;

    if(bomb._position._y > _hitbar->_positionY)
      continue;

    BombClass& bc = _bombClasses[bomb._classId];

    int32_t bithit = bomb._position._x - ((_bombBoomWidth - bc._width) / 2);
    
    // Apply damage to the bar.
    for(int32_t i = 0; i < _bombBoomWidth; ++i){
      bool bitval = (bithit + i / _worldScale) % 2;
      for(int32_t j = 0; j < _hitbar->_height; ++j)
        _hitbar->_bitmap.setBit(j, bithit + i, bitval, false);
    }
    _hitbar->_bitmap.regenerateBytes();

    Vector2i boomPosition {bithit, _hitbar->_positionY + _hitbar->_height};
    boomBomb(bomb, true, boomPosition, BOMBHIT_BOTTOM);
  }
}

void GameState::boomAllBombs()
{
  for(auto& bomb : _bombs){
    if(!bomb._isAlive) continue;
    boomBomb(bomb, false);
  }
}

void GameState::doCollisionsBombsCannon()
{
  if(!_cannon._isAlive)
    return;

  if(_cannon._isBooming)
    return;

  if(_bombCount == 0)
    return;

  if(_isAliensAboveInvasionRow)
    return;

  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  aPosition._x = _cannon._position._x;
  aPosition._y = _cannon._position._y;

  aBitmap = &(pxr::assets->getBitmap(_cannon._cannonKey, _worldScale));

  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;
  
    bPosition._x = bomb._position._x;
    bPosition._y = bomb._position._y;

    BombClass& bc = _bombClasses[bomb._classId];

    bBitmap = &(pxr::assets->getBitmap(bc._bitmapKeys[bomb._frame], _worldScale));

    const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

    if(c._isCollision){
      boomCannon();
      boomBomb(bomb);
    }
  }
}

void GameState::doCollisionsBombsLaser()
{
  if(!_laser._isAlive)
    return;

  if(_bombCount == 0)
    return;

  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  aPosition._x = _laser._position._x;
  aPosition._y = _laser._position._y;

  aBitmap = &(pxr::assets->getBitmap(_laser._bitmapKey, _worldScale));

  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;
  
    bPosition._x = bomb._position._x;
    bPosition._y = bomb._position._y;

    BombClass& bc = _bombClasses[bomb._classId];

    bBitmap = &(pxr::assets->getBitmap(bc._bitmapKeys[bomb._frame], _worldScale));

    const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

    if(c._isCollision){
      boomLaser(true);
      if(randUniformSignedInt(0, bc._laserSurvivalChance) != 0) boomBomb(bomb);
    }
  }
}

void GameState::doCollisionsLaserAliens()
{
  if(!_laser._isAlive)
    return;

  if(_isAliensSpawning)
    return;

  if(_isAliensFrozen)
    return;

  if(_alienPopulation == 0)
    return;

  Vector2i aPosition {}; 
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  aPosition._x = _laser._position._x;
  aPosition._y = _laser._position._y;

  aBitmap = &(pxr::assets->getBitmap(_laser._bitmapKey, _worldScale));

  for(auto& row : _grid){
    for(auto& alien : row){
      if(!alien._isAlive)
        continue;

      const AlienClass& ac = _alienClasses[alien._classId];

      bPosition = alien._position; 
      bBitmap = &(pxr::assets->getBitmap(ac._bitmapKeys[alien._frame], _worldScale));

      const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

      if(c._isCollision){
        if(alien._classId == CUTTLETWIN){
          _alienMorpher = nullptr;
          _isAliensMorphing = false;
        }
        if(_levels[_levelIndex]._isCuttlesOn && alien._classId == CRAB && alien._col != gridWidth - 1)
          morphAlien(alien);
        else
          boomAlien(alien);

        boomLaser(false);
        return;
      }
    }
  }
}

void GameState::doCollisionsLaserUfo()
{
  if(!_laser._isAlive)
    return;

  if(!_ufo._isAlive)
    return;

  if(!_ufo._phase)
    return;

  Vector2i aPosition {}; 
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  aPosition._x = _laser._position._x;
  aPosition._y = _laser._position._y;
  bPosition._x = _ufo._position._x;
  bPosition._y = _ufo._position._y;

  const UfoClass& uc = _ufoClasses[_ufo._classId];
  bBitmap = &(pxr::assets->getBitmap(uc._shipKey, _worldScale));
  aBitmap = &(pxr::assets->getBitmap(_laser._bitmapKey, _worldScale));

  const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

  if(c._isCollision){
    boomLaser(false);
    boomUfo();
  }
}

void GameState::doCollisionsLaserSky()
{
  if(!_laser._isAlive)
    return;

  if(_laser._position._y + _laser._height < _worldTopBorderY)
    return;

  _laser._isAlive = false;
  Vector2i position {};
  position._x = _laser._position._x - ((_bombBoomWidth - _laser._width) / 2);
  position._y = _laser._position._y;
  spawnBoom(position, BOMBHIT_MIDAIR, _laser._colorIndex);
}

bool GameState::doCollisionsAliensBorders()
{
  if(_isAliensSpawning)
    return false;

  if(_isAliensFrozen)
    return false;

  if(_alienPopulation == 0)
    return false;

  switch(_alienMoveDirection){

    // If moving left test against left border.
    case -1: 
      for(int32_t col = 0; col < gridWidth; ++col){
        for(int32_t row = 0; row < gridHeight; ++row){
          Alien& alien = _grid[row][col];

          if(!alien._isAlive)
            continue;

          if(alien._position._x <= _worldLeftBorderX)
            return true;
        }
      }
      break;

    // If moving right test against right border.
    case 1:
      for(int32_t col = gridWidth - 1; col > 0; --col){
        for(int32_t row = 0; row < gridHeight; ++row){
          Alien& alien = _grid[row][col];

          if(!alien._isAlive)
            continue;

          if(alien._position._x + _alienXSeperation >= _worldRightBorderX)
            return true;
        }
      }
      break;

    default:
      assert(0); // should never happen.
  }

  return false;
}

void GameState::doCollisionsBunkersBombs()
{
  if(_bombCount <= 0)
    return;

  if(_bunkers.size() <= 0)
    return;

  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;

    aPosition._y = bomb._position._y;

    if(aPosition._y < _bunkerSpawnY)
      continue;

    if(aPosition._y > _bunkerSpawnY + _bunkerHeight)
      continue;

    aPosition._x = bomb._position._x;

    const BombClass& bc = _bombClasses[bomb._classId];
    aBitmap = &(pxr::assets->getBitmap(bc._bitmapKeys[bomb._frame], _worldScale));

    for(auto iter = _bunkers.begin(); iter != _bunkers.end(); ++iter){
      Bunker& bunker = *(*iter);

      bPosition._x = bunker._position._x;
      bPosition._y = bunker._position._y;

      bBitmap = &bunker._bitmap;

      const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

      if(c._isCollision){
        boomBomb(bomb);
        boomBunker(bunker, c._bPixels.front());
        if(bunker._bitmap.isApproxEmpty(_bunkerDeleteThreshold))
          _bunkers.erase(iter);
        return;
      }
    }
  }
}

void GameState::doCollisionsBunkersLaser()
{
  if(!_laser._isAlive)
    return;

  if(_bunkers.size() == 0)
    return;

  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  aPosition._x = _laser._position._x;
  aPosition._y = _laser._position._y;

  aBitmap = &(pxr::assets->getBitmap(_laser._bitmapKey, _worldScale));

  for(auto iter = _bunkers.begin(); iter != _bunkers.end(); ++iter){
    Bunker& bunker = *(*iter);

    bPosition._x = bunker._position._x;
    bPosition._y = bunker._position._y;

    bBitmap = &(bunker._bitmap);

    const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);
    
    if(c._isCollision){
      boomLaser(false);
      boomBunker(bunker, c._bPixels.front());
      if(bunker._bitmap.isApproxEmpty(_bunkerDeleteThreshold))
        _bunkers.erase(iter);
      return;
    }
  }
}

void GameState::doCollisionsBunkersAliens()
{
  if(_isAliensSpawning)
    return;

  if(_isAliensFrozen)
    return;

  if(_bunkers.size() == 0)
    return;

  if(_alienPopulation == 0)
    return;

  int32_t bottomRow {0};
  while(_rowPops[bottomRow] == 0)
    ++bottomRow;

  if(_grid[bottomRow][0]._position._y > _bunkerSpawnY + _bunkerHeight)
    return;

  Vector2i aPosition {};
  Vector2i bPosition {};

  const Bitmap* aBitmap {nullptr};
  const Bitmap* bBitmap {nullptr};

  for(const auto& alien : _grid[bottomRow]){
    if(!alien._isAlive)
      continue;

    aPosition._x = alien._position._x;
    aPosition._y = alien._position._y;

    const AlienClass& ac = _alienClasses[alien._classId];
    Assets::Key_t bitmapKey = ac._bitmapKeys[alien._frame];
    aBitmap = &(pxr::assets->getBitmap(bitmapKey, _worldScale));

    for(auto iter = _bunkers.begin(); iter != _bunkers.end(); ++iter){
      Bunker& bunker = *(*iter);

      bPosition._x = bunker._position._x;
      bPosition._y = bunker._position._y;

      bBitmap = &(bunker._bitmap);

      const Collision& c = testCollision(aPosition, *aBitmap, bPosition, *bBitmap, false);

      if(c._isCollision){
        bunker._bitmap.setRect(c._bOverlap._ymin, c._bOverlap._xmin, 
                               c._bOverlap._ymax - 1, c._bOverlap._xmax - 1, false);

        if(bunker._bitmap.isApproxEmpty(_bunkerDeleteThreshold))
          _bunkers.erase(iter);

        return;
      }
    }
  }
}

bool GameState::incrementGridIndex(GridIndex& index)
{
  // Increments index from left-to-right along the columns, moving up a row and back to the left
  // most column upon reaching the end of the current column. Loops back to the bottom left most
  // column of the bottom row upon reaching the top-right of the grid. Returns true to indicate
  // a loop.
  
  ++index._col;
  if(index._col >= gridWidth){
    index._col = 0;
    ++index._row;
    if(index._row >= gridHeight){
      index._row = 0;
      return true;
    }
  }
  return false;
}

void GameState::addHudMsg(const char* msg, const Color3f& color)
{
  int32_t msgWidth = _font->calculateStringWidth(msg);
  Vector2i msgPosition {
    (_worldSize._x - msgWidth) / 2,
    msgHeight_px * _worldScale
  };

  _uidMsgText = _hud->addTextLabel({
    msgPosition, 
    color,
    msg,
    0.5f,
    true
  });
}

void GameState::removeHudMsg()
{
  _hud->removeTextLabel(_uidMsgText);
}

void GameState::doInvasionTest()
{
  if(_isGameOver)
    return;

  if(_alienPopulation == 0)
    return;

  if(_isAliensFrozen)
    return;

  if(_isAliensSpawning)
    return;

  // dont trigger invasion until all alive have dropped.
  if(_isAliensDropping)
    return;

  int32_t minY {std::numeric_limits<int32_t>::max()};
  for(auto& row : _grid)
    for(auto& alien : row)
      if(alien._isAlive)
        minY = std::min(minY, alien._position._y);

  if(minY == _alienInvasionHeight)
    startGameOver();
  else if(minY == _alienInvasionHeight + std::abs(_alienDropDisplacement._y)){
    _isAliensAboveInvasionRow = true;
  }
}

void GameState::startRoundIntro()
{
  int32_t round = static_cast<SpaceInvaders*>(_app)->getRound();
  std::string msg {};
  msg += msgRoundIntro;
  msg += " ";
  msg += std::to_string(round);
  addHudMsg(msg.c_str(), colors::red);
  _msgClockSeconds = 0.f;
  _isRoundIntro = true;
}

void GameState::doRoundIntro(float dt)
{
  if(!_isRoundIntro)
    return;

  _msgClockSeconds += dt;
  if(_msgClockSeconds >= msgPeriodSeconds){
    removeHudMsg();
    _isRoundIntro = false;
  }
}

void GameState::startGameOver()
{
  _cannon._isAlive = false;
  _isAliensFrozen = true;
  _beatBox.pause();
  
  if(_ufo._isAlive) 
    mixer->stopChannel(_ufoSfxChannel);

  static_cast<SpaceInvaders*>(_app)->startScoreHudFlash();

  addHudMsg(msgGameOver, colors::red);

  _msgClockSeconds = 0.f;
  _isGameOver = true;
}

void GameState::doGameOver(float dt)
{
  if(!_isGameOver)
    return;

  assert(!_isVictory);
  assert(!_isRoundIntro);

  _msgClockSeconds += dt;
  if(_msgClockSeconds >= msgPeriodSeconds){
    SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
    si->stopScoreHudFlash();
    removeHudMsg();
    if(si->isHiScore(si->getScore()))
      _app->switchState(HiScoreRegState::name);
    else 
      _app->switchState(HiScoreBoardState::name);
  }
}

void GameState::doVictoryTest()
{
  if(!_isVictory && _alienPopulation == 0){
    _beatBox.pause();
    if(_ufo._isAlive)
      return;
    else
      startVictory();
  }
}

void GameState::startVictory()
{
  _beatBox.pause();
  if(_ufo._isAlive) 
    mixer->stopChannel(_ufoSfxChannel);

  addHudMsg(msgVictory, colors::green);
  static_cast<SpaceInvaders*>(_app)->startScoreHudFlash();

  boomAllBombs();

  _msgClockSeconds = 0.f;
  _isVictory = true;
}

void GameState::doVictory(float dt)
{
  if(!_isVictory)
    return;

  assert(!_isGameOver);
  assert(!_isRoundIntro);

  _msgClockSeconds += dt;
  if(_msgClockSeconds > msgPeriodSeconds){
    SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
    si->addRound(1);
    si->stopScoreHudFlash();
    removeHudMsg();
    _app->switchState(SosState::name);
  }
}

void GameState::onUpdate(double now, float dt)
{
  int32_t beats = _cycles[_activeCycle][_activeBeat]; 

  doRoundIntro(dt);
  doAlienMorphing(dt);
  doBombMoving(beats, dt);
  doLaserMoving(dt);
  doAlienMoving(beats);
  doAlienBombing(beats);
  doCannonMoving(dt);
  doUfoMoving(dt);
  doUfoPhasing(dt);
  doUfoSpawning();
  doCannonBooming(dt);
  doAlienBooming(dt);
  doBombBoomBooming(dt);
  doUfoBoomScoring(dt);
  doCannonFiring();
  doCollisionsUfoBorders();
  doCollisionsBombsHitbar();
  doCollisionsBombsCannon();
  doCollisionsBombsLaser();
  doCollisionsLaserAliens();
  doCollisionsBunkersBombs();
  doCollisionsBunkersLaser();
  doCollisionsBunkersAliens();
  doCollisionsLaserUfo();
  doCollisionsLaserSky();


  //================================================================================
  
  if(pxr::input->isKeyPressed(Input::KEY_b))
    boomCannon();

  if(pxr::input->isKeyPressed(Input::KEY_a)){
    for(auto& row : _grid){
      for(auto& alien : row){
        if(alien._isAlive){
          boomAlien(alien);
          goto BOOMED;
        }
      }
    }
  }
  BOOMED:

  //================================================================================

  doVictoryTest();
  doVictory(dt);
  doInvasionTest();
  doGameOver(dt);
  updateActiveCycleBeat();
  _beatBox.doBeats(dt);
}

void GameState::drawGrid()
{
  if(_isRoundIntro) 
    return;

  for(const auto& row : _grid){
    for(const auto& alien : row){
      if(alien._isAlive){
        const AlienClass& ac = _alienClasses[alien._classId];

        Vector2f position(alien._position._x, alien._position._y);
        Assets::Key_t bitmapKey = ac._bitmapKeys[alien._frame];
        Color3f& color = _colorPalette[ac._colorIndex];

        renderer->blitBitmap(position, pxr::assets->getBitmap(bitmapKey, _worldScale), color);
      }
    }
  }

  if(_isAliensBooming){
    assert(_alienBoomer != nullptr);

    const AlienClass& ac = _alienClasses[_alienBoomer->_classId];

    Vector2f position(_alienBoomer->_position._x, _alienBoomer->_position._y);
    Assets::Key_t bitmapKey = SpaceInvaders::BMK_ALIENBOOM;
    Color3f& color = _colorPalette[ac._colorIndex];

    renderer->blitBitmap(position, pxr::assets->getBitmap(bitmapKey, _worldScale), color);
  }
}

void GameState::drawUfo()
{
  if(!((_ufo._isAlive && _ufo._phase) || _isUfoBooming || _isUfoScoring))
    return;

  const UfoClass& uc = _ufoClasses[_ufo._classId];

  Assets::Key_t bitmapKey;
  if(_ufo._isAlive)
    bitmapKey = uc._shipKey;
  else if(_isUfoBooming)
    bitmapKey = uc._boomKey;
  else if(_isUfoScoring){
    return; // drawn as a hud element instead.
  }

  renderer->blitBitmap(
      _ufo._position, 
      pxr::assets->getBitmap(bitmapKey, _worldScale), 
      _colorPalette[uc._colorIndex]
  );
}

void GameState::drawCannon()
{
  if(!(_cannon._isBooming || _cannon._isAlive))
    return;

  Assets::Key_t bitmapKey;
  if(_cannon._isBooming){
    bitmapKey = _cannon._boomKeys[_cannon._boomFrame];
  }
  else if(_cannon._isAlive){
    bitmapKey = _cannon._cannonKey;
  }

  Color3f& color = _colorPalette[_cannon._colorIndex];

  renderer->blitBitmap(_cannon._position, pxr::assets->getBitmap(bitmapKey, _worldScale), color);
}

void GameState::drawBombs()
{
  for(auto& bomb : _bombs){
    if(!bomb._isAlive)
      continue;

    const BombClass& bc = _bombClasses[bomb._classId];
    Assets::Key_t bitmapKey = bc._bitmapKeys[bomb._frame];
    Color3f& color = _colorPalette[bc._colorIndex];
    renderer->blitBitmap(bomb._position, pxr::assets->getBitmap(bitmapKey, _worldScale), color);
  }
}

void GameState::drawBombBooms()
{
  for(auto& boom : _bombBooms){
    if(!boom._isAlive)
      continue;

    Assets::Key_t bitmapKey = _bombBoomKeys[boom._hit];
    Color3f& color = _colorPalette[boom._colorIndex];
    Vector2f position = Vector2f(boom._position._x, boom._position._y);
    renderer->blitBitmap(position, pxr::assets->getBitmap(bitmapKey, _worldScale), color);
  }
}

void GameState::drawLaser()
{
  if(!_laser._isAlive)
    return;

  renderer->blitBitmap(_laser._position, pxr::assets->getBitmap(_laser._bitmapKey, _worldScale), _colorPalette[_laser._colorIndex]);
}

void GameState::drawHitbar()
{
  renderer->blitBitmap({0.f, _hitbar->_positionY}, _hitbar->_bitmap, _colorPalette[_hitbar->_colorIndex]);
}

void GameState::drawBunkers()
{
  for(const auto& bunker : _bunkers)
    renderer->blitBitmap(bunker->_position, bunker->_bitmap, _colorPalette[_bunkerColorIndex]);
}

void GameState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);
  drawGrid();
  drawUfo();
  drawCannon();
  drawBombs();
  drawBombBooms();
  drawLaser();
  drawBunkers();
  drawHitbar();
}

void GameState::onEnter()
{
  startNextLevel();
}

//===============================================================================================//
// ##>SOS STATE                                                                                  //
//===============================================================================================//

void SosState::initialize(Vector2i worldSize, int32_t worldScale)
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  _hud = &si->getHud();
  _font = &pxr::assets->getFont(SpaceInvaders::fontKey, worldScale);

  _worldSize = worldSize;
  _worldScale = worldScale;
  _exitHeight_px = worldSize._y - (baseTopMargin_px * worldScale);
  _worldLeftMargin_px = baseWorldMargin_px * worldScale;
  _worldRightMargin_px = worldSize._x - _worldLeftMargin_px;
  _spawnHeight_px = baseSpawnHeight_px * worldScale;
  _moveSpeed = baseMoveSpeed * worldScale;
  _sosTextPositionX = _worldSize._x - _font->calculateStringWidth(sosText) - (sosTextMargin_px * _worldScale);
}

void SosState::onUpdate(double now, float dt)
{
  doEngineCheck();
  doAlienAnimating(dt);
  doMoving(dt);
  doEngineFailing(dt);
  doWallColliding();
  doEndTest();
}

void SosState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);

  const GameState::AlienClass& ac = _gameState->_alienClasses[_alien._classId];
  Assets::Key_t bitmapKey;

  if(_hasEngineFailed){
    const Bitmap& bitmap = assets->getBitmap(SpaceInvaders::BMK_SOS_TRAIL, _worldScale);
    Vector2f trailPosition = _alien._failPosition;
    while(trailPosition._y < _alien._position._y){
      renderer->blitBitmap(
          trailPosition,
          bitmap,
          _gameState->_colorPalette[ac._colorIndex]
      );
      trailPosition._y += (sosTrailSpace_px * _worldScale);
    }
  }

  bitmapKey = ac._bitmapKeys[_alien._frame];
  renderer->blitBitmap(
      _alien._position, 
      assets->getBitmap(bitmapKey, _worldScale), 
      _gameState->_colorPalette[ac._colorIndex]
  );

  const GameState::UfoClass& uc = _gameState->_ufoClasses[_ufo._classId];
  bitmapKey = uc._shipKey;
  renderer->blitBitmap(
      _ufo._position, 
      assets->getBitmap(bitmapKey, _worldScale), 
      _gameState->_colorPalette[uc._colorIndex]
  );

}

void SosState::onEnter()
{
  _ufo._classId = GameState::UfoClassId::SAUCER;
  _ufo._position._y = _spawnHeight_px;
  _ufo._position._x = pxr::randUniformSignedInt(_worldLeftMargin_px, _worldRightMargin_px);

  _alien._classId = _gameState->_lastClassAlive; 
  _alien._frame = false;
  _alien._frameClockSeconds = 0.f;

  _ufo._width = _gameState->_ufoClasses[_ufo._classId]._width;
  int32_t ufoHeight = _gameState->_ufoClasses[_ufo._classId]._height;
  int32_t alienWidth = _gameState->_alienClasses[_alien._classId]._width;

  _alien._position._x = _ufo._position._x + ((_ufo._width - alienWidth) / 2);
  _alien._position._y = _spawnHeight_px + ufoHeight;

  _moveVelocity = {
    _moveSpeed * std::sin(moveAngleRadians), 
    _moveSpeed * std::cos(moveAngleRadians)
  };

  _isEngineFailing = false;
  _hasEngineFailed = false;

  _engineFailClockSeconds = 0.f;

  _woowooChannel = mixer->playSound(SpaceInvaders::SK_SOS, 100);
  _isWooing = true;

  _nextSosText = 0;

  static_cast<SpaceInvaders*>(_app)->showHud();
}

void SosState::doMoving(float dt)
{
  if(_isEngineFailing)
    return;

  if(_hasEngineFailed){
    _alien._position._y += 0.6f * _moveSpeed * dt;
  }
  else{
    _alien._position += _moveVelocity * dt;
    _ufo._position += _moveVelocity * dt;
  }
}

void SosState::doAlienAnimating(float dt)
{
  if(!_hasEngineFailed)
    return;

  _alien._frameClockSeconds += dt;
  if(_alien._frameClockSeconds >= Alien::framePeriodSeconds){
    _alien._frame = !_alien._frame;
    _alien._frameClockSeconds = 0.f;
  }
}

void SosState::doEngineFailing(float dt)
{
  if(!_isEngineFailing)
    return;

  _engineFailClockSeconds += dt;
  if(_engineFailClockSeconds >= engineFailPeriodSeconds){
    _alien._failPosition = _alien._position;
    _alien._frameClockSeconds = 0.f;
    _isEngineFailing = false;
    _hasEngineFailed = true;
  }
}

void SosState::doEngineCheck()
{
  if(_hasEngineFailed)
    return;

  if(!_isEngineFailing){
    if(pxr::randUniformSignedInt(0, engineFailChance) == engineFailHit){
      _isEngineFailing = true;
      _engineFailClockSeconds = 0.f;

      int32_t troubleWidth = _font->calculateStringWidth(troubleText);
      Vector2i troublePosition = {
        _ufo._position._x - (std::abs(_ufo._width - troubleWidth) / 2),
        _ufo._position._y - _font->getLineSpace()
      };
      _uidTroubleText = _hud->addTextLabel({
        troublePosition, 
        pxr::colors::magenta, 
        troubleText,
        1.0f,
        true
      });

      mixer->stopChannel(_woowooChannel);
      _isWooing = false;
    }
  }
}

void SosState::doWallColliding()
{
  if(_isEngineFailing || _hasEngineFailed)
    return;

  if(((_ufo._position._x < _worldLeftMargin_px) && _moveVelocity._x < 0) ||
     (((_ufo._position._x + _ufo._width) > _worldRightMargin_px) && _moveVelocity._x > 0))
  {
    _moveVelocity._x *= -1.f;

    // condition should never fail as array is calibrated to be big enough in all cases.
    if(_nextSosText < maxSosTextDrop){
      Vector2i sosPosition = {
        _sosTextPositionX,
        _ufo._position._y + (_font->getSize() / 2)
      };
      _uidSosText[_nextSosText++] = _hud->addTextLabel({
        sosPosition, 
        pxr::colors::magenta, 
        sosText
      });
    }
  }
}

void SosState::doEndTest()
{
  if(_alien._position._y > _exitHeight_px){
    if(_isWooing)
      mixer->stopChannel(_woowooChannel);
    if(_hasEngineFailed)
      _hud->removeTextLabel(_uidTroubleText);
    for(int32_t i = _nextSosText; i > 0; --i)
      _hud->removeTextLabel(_uidSosText[i - 1]);
    _app->switchState(GameState::name);
  }
}

//===============================================================================================//
// ##>MENU STATE                                                                                 //
//===============================================================================================//

void MenuState::initialize(Vector2i worldSize, int32_t worldScale)
{
  _worldScale = worldScale;
}

void MenuState::onUpdate(double now, float dt)
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  if(pxr::input->isKeyPressed(Input::KEY_ENTER)){
    depopulateHud();
    si->hideHud();
    _app->switchState(GameState::name);
  }
  if(pxr::input->isKeyPressed(Input::KEY_s)){
    depopulateHud();
    si->hideHud();
    _app->switchState(HiScoreBoardState::name);
  }
}

void MenuState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);
}

void MenuState::onEnter()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  populateHud();
  si->showHud();
  si->hideLivesHud();
  si->showTopHud();
  si->resetGameStats();
}

void MenuState::populateHud()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  HUD& hud = si->getHud();
  _uidMenuText = hud.addTextLabel({Vector2i{91, 204} * _worldScale, pxr::colors::cyan, "*MENU*"});
  _uidMenuBitmap = hud.addBitmapLabel({Vector2i{56, 182} * _worldScale, pxr::colors::white, &(pxr::assets->getBitmap(SpaceInvaders::BMK_MENU, _worldScale))});
  _uidControlsText = hud.addTextLabel({Vector2i{76, 162} * _worldScale, pxr::colors::cyan, "*CONTROLS*"});
  _uidControlsBitmap = hud.addBitmapLabel({Vector2i{58, 134} * _worldScale, pxr::colors::white, &(pxr::assets->getBitmap(SpaceInvaders::BMK_CONTROLS, _worldScale))});
  _uidTablesText = hud.addTextLabel({Vector2i{40, 108} * _worldScale, pxr::colors::cyan, "*SCORE ADVANCE TABLE*"});
  _uidSchroBitmap = hud.addBitmapLabel({Vector2i{62, 90} * _worldScale, pxr::colors::magenta, &(pxr::assets->getBitmap(SpaceInvaders::BMK_SCHRODINGER, _worldScale))});
  _uidSaucerBitmap = hud.addBitmapLabel({Vector2i{62, 74} * _worldScale, pxr::colors::magenta, &(pxr::assets->getBitmap(SpaceInvaders::BMK_SAUCER, _worldScale))});
  _uidSquidBitmap = hud.addBitmapLabel({Vector2i{66, 58} * _worldScale, pxr::colors::yellow, &(pxr::assets->getBitmap(SpaceInvaders::BMK_SQUID0, _worldScale))});
  _uidCuttleBitmap = hud.addBitmapLabel({Vector2i{52, 58} * _worldScale, pxr::colors::yellow, &(pxr::assets->getBitmap(SpaceInvaders::BMK_CUTTLE0, _worldScale))});
  _uidCrabBitmap = hud.addBitmapLabel({Vector2i{64, 42} * _worldScale, pxr::colors::yellow, &(pxr::assets->getBitmap(SpaceInvaders::BMK_CRAB0, _worldScale))});
  _uidOctopusBitmap = hud.addBitmapLabel({Vector2i{64, 26} * _worldScale, pxr::colors::red, &(pxr::assets->getBitmap(SpaceInvaders::BMK_OCTOPUS0, _worldScale))});
  _uid500PointsText = hud.addTextLabel({Vector2i{82, 90} * _worldScale, pxr::colors::magenta, "= 500 POINTS", 0.f, true});
  _uidMysteryPointsText = hud.addTextLabel({Vector2i{82, 74} * _worldScale, pxr::colors::magenta, "= ? MYSTERY", 1.f, true});
  _uid30PointsText = hud.addTextLabel({Vector2i{82, 58} * _worldScale, pxr::colors::yellow, "= 30 POINTS", 2.f, true});
  _uid20PointsText = hud.addTextLabel({Vector2i{82, 42} * _worldScale, pxr::colors::yellow, "= 20 POINTS", 3.f, true});
  _uid10PointsText = hud.addTextLabel({Vector2i{82, 26} * _worldScale, pxr::colors::red, "= 10 POINTS", 4.f, true});
}

void MenuState::depopulateHud()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  HUD& hud = si->getHud();
  hud.removeTextLabel(_uidMenuText);
  hud.removeTextLabel(_uidControlsText);
  hud.removeTextLabel(_uidTablesText);
  hud.removeTextLabel(_uid500PointsText);
  hud.removeTextLabel(_uidMysteryPointsText);
  hud.removeTextLabel(_uid30PointsText);
  hud.removeTextLabel(_uid20PointsText);
  hud.removeTextLabel(_uid10PointsText);
  hud.removeBitmapLabel(_uidMenuBitmap);
  hud.removeBitmapLabel(_uidControlsBitmap);
  hud.removeBitmapLabel(_uidSchroBitmap);
  hud.removeBitmapLabel(_uidSaucerBitmap);
  hud.removeBitmapLabel(_uidSquidBitmap);
  hud.removeBitmapLabel(_uidCuttleBitmap);
  hud.removeBitmapLabel(_uidCrabBitmap);
  hud.removeBitmapLabel(_uidOctopusBitmap);
}

//===============================================================================================//
// ##>HIGH SCORE REGISTRATION STATE                                                              //
//===============================================================================================//

HiScoreRegState::Keypad::Keypad(const Font& font, Vector2i worldSize, int32_t worldScale) :
  _keyText{},
  _keyScreenPosition{},
  _keyColor{colors::cyan},
  _specialKeyColor{colors::magenta},
  _cursorColor{colors::green},
  _cursorPadPosition{initialCursorPadPosition},
  _cursorScreenPosition{0, 0},
  _padScreenPosition{0, 0},
  _font{font}
{
  _keyText = {{
    {"\\", "/", "(", ")", "+", "^", "RUB", "", "", "END", ""},  // row[0] == bottom row
    {"W", "X", "Y", "Z", ".", "_", "-", "[", "]", "<", ">"},
    {"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V"},
    {"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}     // row[3] == top row
  }};

  int32_t fontSize = font.getSize();

  int32_t padWidth = keyColCount * (fontSize + (keySpace_px * worldScale));
  int32_t padHeight = keyRowCount * (fontSize + (keySpace_px * worldScale));
  _padScreenPosition._x = (worldSize._x - padWidth) / 2;
  _padScreenPosition._y = (worldSize._y - padHeight) / 2;

  for(size_t row{0}; row < keyRowCount; ++row){
    for(size_t col{0}; col < keyColCount; ++col){
      _keyScreenPosition[row][col] = {
        _padScreenPosition._x + (col * (fontSize + (keySpace_px * worldScale))),
        _padScreenPosition._y + (row * (fontSize + (keySpace_px * worldScale))),
      };
    }
  }

  updateCursorScreenPosition();
}

void HiScoreRegState::Keypad::moveCursor(int32_t colShift, int32_t rowShift)
{
  do{
    if(colShift) _cursorPadPosition._x = pxr::wrap(_cursorPadPosition._x + colShift, 0, keyColCount - 1);
    if(rowShift) _cursorPadPosition._y = pxr::wrap(_cursorPadPosition._y + rowShift, 0, keyRowCount - 1);
  }
  while(_keyText[_cursorPadPosition._y][_cursorPadPosition._x][0] == '\0');
  updateCursorScreenPosition();
}

void HiScoreRegState::Keypad::updateCursorScreenPosition()
{
  _cursorScreenPosition = _keyScreenPosition[_cursorPadPosition._y][_cursorPadPosition._x];
  _cursorScreenPosition._y -= cursorDrop_px;
  const char* keyText = _keyText[_cursorPadPosition._y][_cursorPadPosition._x];
  if(std::strncmp(keyText, "RUB", 3) == 0 || std::strncmp(keyText, "END", 3) == 0)
    _cursorScreenPosition._x += _font.getGlyph(keyText[0])._advance + _font.getGlyphSpace();
}

void HiScoreRegState::Keypad::reset()
{
  _cursorPadPosition = initialCursorPadPosition;
}

void HiScoreRegState::Keypad::draw()
{
  for(int row{0}; row < keyRowCount; ++row){
    for(int col{0}; col < keyColCount; ++col){
      const char* text = _keyText[row][col];
      if(text[0] == '\0') continue;
      Color3f color = _keyColor;
      if(strncmp(text, "RUB", 3) == 0 || strncmp(text, "END", 3) == 0)
        color = _specialKeyColor;
      renderer->blitText(_keyScreenPosition[row][col], text, _font, color);
    }
  }

  renderer->blitText(_cursorScreenPosition, cursorChar, _font, _cursorColor);
}

HiScoreRegState::NameBox::NameBox(const Font& font, Vector2i worldSize, int32_t worldScale) :
  _nameBuffer{},
  _font{font},
  _final{}
{
  for(auto& c : _nameBuffer)
    c = nullChar;

  composeFinal();

  int32_t finalWidth_px = font.calculateStringWidth(_final);
  _boxScreenPosition = {
    (worldSize._x - finalWidth_px) / 2,
    worldSize._y / 4
  };
}

void HiScoreRegState::NameBox::draw()
{
  renderer->blitText(_boxScreenPosition, _final, _font, colors::red);
}

bool HiScoreRegState::NameBox::pushBack(char c)
{
  if(isFull()) 
    return false;
  for(int i = _nameBuffer.size() - 1; i >= 0; --i){
    if(i == 0) 
      _nameBuffer[i] = c;
    else if(_nameBuffer[i] == nullChar && _nameBuffer[i - 1] != nullChar){ 
      _nameBuffer[i] = c; 
      break;
    }
  }
  composeFinal();
  return true;
}

bool HiScoreRegState::NameBox::popBack()
{
  if(isEmpty()) 
    return false;
  if(isFull()){
    _nameBuffer[_nameBuffer.size() - 1] = nullChar;
    composeFinal();
    return true;
  }
  for(int i = _nameBuffer.size() - 1; i >= 0; --i){
    if(i == 0) 
      _nameBuffer[i] = nullChar;
    else if(_nameBuffer[i] == nullChar && _nameBuffer[i - 1] != nullChar){
      _nameBuffer[i-1] = nullChar; 
      break;
    }
  }
  composeFinal();
  return true;
}

bool HiScoreRegState::NameBox::isFull() const
{
  return _nameBuffer[SpaceInvaders::hiscoreNameLen - 1] != nullChar;
}

bool HiScoreRegState::NameBox::isEmpty() const
{
  return _nameBuffer[0] == nullChar;
}

void HiScoreRegState::NameBox::composeFinal()
{
  _final.clear();
  _final += label;
  _final += ' ';
  _final += quoteChar;
  for(auto& c : _nameBuffer)
    _final += c;
  _final += quoteChar;
}

void HiScoreRegState::initialize(Vector2i worldSize, int32_t worldScale)
{
  _keypad = std::make_unique<Keypad>(assets->getFont(SpaceInvaders::fontKey, worldScale), worldSize, worldScale);
  _nameBox = std::make_unique<NameBox>(assets->getFont(SpaceInvaders::fontKey, worldScale), worldSize, worldScale);
}

void HiScoreRegState::onUpdate(double now, float dt)
{
  doInput();
}

void HiScoreRegState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);
  _keypad->draw();
  _nameBox->draw();
}

void HiScoreRegState::onEnter()
{
  if(_keypad != nullptr) _keypad->reset();
  static_cast<SpaceInvaders*>(_app)->showHud();
}

void HiScoreRegState::doInput()
{
  bool lKey = pxr::input->isKeyPressed(Input::KEY_LEFT);
  bool rKey = pxr::input->isKeyPressed(Input::KEY_RIGHT);
  bool uKey = pxr::input->isKeyPressed(Input::KEY_UP);
  bool dKey = pxr::input->isKeyPressed(Input::KEY_DOWN);

  int colShift {0}, rowShift {0};
  if(lKey) colShift += -1;
  if(rKey) colShift += 1;
  if(uKey) 
    rowShift += 1;
  if(dKey) 
    rowShift += -1;

  _keypad->moveCursor(colShift, rowShift);  

  bool eKey = pxr::input->isKeyPressed(Input::KEY_SPACE);
  if(eKey){
    const char* c = _keypad->getActiveKeyText();
    if(strncmp(c, "RUB", 3) == 0){
      if(!_nameBox->popBack())
        mixer->playSound(SpaceInvaders::SK_FAST1); 
    }
    else if(strncmp(c, "END", 3) == 0){
      if(!_nameBox->isFull()){
        mixer->playSound(SpaceInvaders::SK_FAST1); 
      }
      else{
        SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
        si->setPlayerName(_nameBox->getBufferText());
        si->hideHud();
        _app->switchState(HiScoreBoardState::name);
      }
    }
    else{
      if(!_nameBox->pushBack(c[0]))
        mixer->playSound(SpaceInvaders::SK_FAST1); 
    }
  }
}

//===============================================================================================//
// ##>HIGH SCORE BOARD STATE                                                                     //
//===============================================================================================//

void HiScoreBoardState::initialize(Vector2i worldSize, int32_t worldScale)
{
  _font = &(assets->getFont(SpaceInvaders::fontKey, worldScale));

  // world scale of glyphs is taken into account by the fontSize returned from font.
  int32_t glyphSize_px = _font->getSize() + _font->getGlyphSpace();
  int32_t nameWidth_px = glyphSize_px * SpaceInvaders::hiscoreNameLen;
  int32_t scaledColSeperation = colSeperation * worldScale;
  _scoreBoardSize._x = nameWidth_px + scaledColSeperation + (scoreDigitCountEstimate * glyphSize_px);
  _scoreBoardSize._y = (SpaceInvaders::hiscoreCount + 1) * (glyphSize_px + (rowSeperation * worldScale));

  _nameScreenPosition = {
    (worldSize._x - _scoreBoardSize._x) / 2,
    (worldSize._y - _scoreBoardSize._y) / 3
  };

  _scoreScreenPosition = _nameScreenPosition;
  _scoreScreenPosition._x += nameWidth_px + scaledColSeperation;
}

void HiScoreBoardState::onEnter()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  const auto& hiscores = si->getHiScores();

  _newScore._value = si->getScore();
  _newScore._name = si->getPlayerName();

  if(_newScore._name[0] == '\0')
    _newScore._name = placeHolderName;

  _scoreBoard[0] = &_newScore;

  for(int i{0}; i < SpaceInvaders::hiscoreCount; ++i)
    _scoreBoard[i + 1] = &hiscores[i];

  _eventNum = 0;
  _eventClock = 0.f;

  populateHud();
  si->showHud();
}

void HiScoreBoardState::onUpdate(double now, float dt)
{
  _eventClock += dt;
  if(_eventNum == 0){
    if(_eventClock > enterDelaySeconds){
      _eventClock = 0.f;
      ++_eventNum;
    }
  }
  else if(_eventNum > _scoreBoard.size()){ 
    if(_eventClock > _exitDelaySeconds){
      depopulateHud();
      SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
      si->hideHud();
      if(si->isHiScore(_newScore._value) && !si->isDuplicateHiScore(_newScore)){
        si->registerHiScore(_newScore);
        si->writeHiScores();
      }
      _app->switchState(MenuState::name);
    }
  }
  else {
    if(_eventClock > swapScoreDelaySeconds){
      _eventClock = 0.f;
      _eventNum += doScoreSwap() ? _scoreBoard.size() : 1;  // if done all swaps skip to end.
      if(_eventNum > _scoreBoard.size()){ 
        if(newScoreIsTop()){
          mixer->playSound(SpaceInvaders::SK_TOPSCORE); 
          _exitDelaySeconds = topScoreExitDelaySeconds;
        }
        else
          _exitDelaySeconds = normalExitDelaySeconds;
      }
    }
  }
}

void HiScoreBoardState::onDraw(double now, float dt)
{
  renderer->clearViewport(colors::black);
  Vector2i namePosition {_nameScreenPosition};
  Vector2i scorePosition {_scoreScreenPosition};
  std::string nameStr {};
  const Color3f* color;
  for(auto& score : _scoreBoard){
    color = (score == &_newScore) ? &newScoreColor : &oldScoreColor;

    // we do this because the name in SpaceInvaders::Score is not null terminated.
    nameStr.clear();
    for(int i{0}; i < SpaceInvaders::hiscoreNameLen; ++i)
      nameStr += score->_name[i];
  
    renderer->blitText(namePosition, nameStr, *_font, *color);
    renderer->blitText(scorePosition, std::to_string(score->_value), *_font, *color);

    namePosition._y += rowSeperation + _font->getLineSpace();
    scorePosition._y += rowSeperation + _font->getLineSpace();
  }
}

bool HiScoreBoardState::doScoreSwap()
{
  for(int i{0}; i < _scoreBoard.size(); ++i){
    if(_scoreBoard[i] != &_newScore) continue;
    if(i == (_scoreBoard.size() - 1)) return true;
    if(_scoreBoard[i]->_value <= _scoreBoard[i + 1]->_value) return true;
    std::swap(_scoreBoard[i], _scoreBoard[i + 1]);
    mixer->playSound(SpaceInvaders::SK_SCORE_BEEP); 
    return false;
  }
}

void HiScoreBoardState::populateHud()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  Vector2i worldSize = si->getWorldSize();
  int32_t worldScale = si->getWorldScale();
  int32_t titleWidth = _font->calculateStringWidth(titleString);
  Vector2i titlePosition {
    (worldSize._x - titleWidth) / 2,
    _nameScreenPosition._y + _scoreBoardSize._y + (boardTitleSeperation * worldScale)
  };
  HUD& hud = si->getHud();
  _uidTitleText = hud.addTextLabel({titlePosition, titleColor, titleString});
}

void HiScoreBoardState::depopulateHud()
{
  SpaceInvaders* si = static_cast<SpaceInvaders*>(_app);
  HUD& hud = si->getHud();
  hud.removeTextLabel(_uidTitleText);
}

bool HiScoreBoardState::newScoreIsTop()
{
  return _scoreBoard[_scoreBoard.size() - 1] == &_newScore;
}
