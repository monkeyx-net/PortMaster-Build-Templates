#include "pixiretro.h"

namespace pxr
{

//===============================================================================================//
// ##>RANDOM NUMBER GENERATION                                                                   //
//===============================================================================================//

constexpr xorwow::state_type xorwow::default_seed;

xorwow::xorwow(result_type value)
{
  seed(value);
}

xorwow::xorwow(const state_type& seeds)
{
  seed(seeds);
}

xorwow::xorwow(std::seed_seq& seq)
{
  seed(seq);
}

void xorwow::seed() 
{
  _state = default_seed;
}

void xorwow::seed(result_type seed)
{
  _state = default_seed;
  _state[state_size - 2] = seed;
}

void xorwow::seed(state_type seeds)
{
  for(int i{0}; i < state_size; ++i)
    _state[i] = seeds[i] != 0 ? seeds[i] : default_seed[i];
}

void xorwow::seed(std::seed_seq& seq)
{
  seq.generate(_state.begin(), _state.end());
  //*_state.end() = 0; why did I do this!? This has to be a bug!
}

//
// sequence generator from: https://en.wikipedia.org/wiki/Xorshift
//
xorwow::result_type xorwow::operator()()
{
  //
  // Algorithm "xorwow" from p. 5 of Marsaglia, "Xorshift RNGs"
  //
  result_type t = _state[4];
  result_type s = _state[0];
  _state[4] = _state[3];
  _state[3] = _state[2];
  _state[2] = _state[1];
  _state[1] = s;
  t ^= t >> 2;
  t ^= t << 1;
  t ^= s ^ (s << 4);
  _state[0] = t;
  _state[5] += 362437;
  return t + _state[5];
}

void xorwow::discard(unsigned long long z)
{
  while(z--)
    (*this)();
}

constexpr xorwow::result_type xorwow::min()
{
  return std::numeric_limits<result_type>::min();
}

constexpr xorwow::result_type xorwow::max()
{
  return std::numeric_limits<result_type>::max();
}

bool operator==(const xorwow& lhs, const xorwow& rhs)
{
  const auto& lhss = lhs.getState();
  const auto& rhss = rhs.getState();
  for(int i{0}; i < xorwow::state_size; ++i)
    if(lhss[i] != rhss[i])
      return false;
  return true;
}

bool operator!=(const xorwow& lhs, const xorwow& rhs)
{
  return !(lhs == rhs);
}

xorwow randGenerator;

//
// implementation note: creating a new distribution upon generating each number is actually not
// much slower than using a single distribution but comes with a massive advantage. Distributions
// cannot be copied or moved thus making them a pain to maintain within the objects that need 
// them since it restricts the ability to copy and move such objects. It is thus far more
// flexible to just call a function which requires the callee to maintain no internal state.
//
// note: my tests (on a different but similar xorshift algorithm) showed generating 10^6 random 
// ints, creating a new distribution each time, took around 3^4us on my machine, whereas using 
// the same distribution took around 2.4^4, i.e. around 80% of the time. This is an acceptable 
// difference for my needs.
//

int randUniformSignedInt(int lo, int hi)
{
  std::uniform_int_distribution<int> d {lo, hi};
  return d(randGenerator);
}

unsigned int randUniformUnsignedInt(unsigned int lo, unsigned int hi)
{
  std::uniform_int_distribution<unsigned int> d {lo, hi};
  return d(randGenerator);
}

double randUniformReal(double lo, double hi)
{
  std::uniform_real_distribution<double> d {lo, hi};
  return d(randGenerator);
}

//===============================================================================================//
// ##>LOG                                                                                        //
//===============================================================================================//

Log::Log()
{
  _os.open(filename, std::ios_base::trunc);
  if(!_os){
    log(ERROR, logstr::fail_open_log);
    log(INFO, logstr::info_stderr_log);
  }
}

Log::~Log()
{
  if(_os)
    _os.close();
}

void Log::log(Level level, const char* error, const std::string& addendum)
{
  std::ostream& o {_os ? _os : std::cerr}; 
  o << lvlstr[level] << delim << error;
  if(!addendum.empty())
    o << delim << addendum;
  o << std::endl;
}

std::unique_ptr<Log> log {nullptr};

//===============================================================================================//
// ##>INPUT                                                                                      //
//===============================================================================================//

Input::Input() : 
  _history{}
{
  for(auto& key : _keys)
    key._isDown = key._isReleased = key._isPressed = false;
}

void Input::onKeyEvent(const SDL_Event& event)
{
  assert(event.type == SDL_KEYDOWN || event.type == SDL_KEYUP);

  KeyCode key = convertSdlKeyCode(event.key.keysym.sym);

  if(key == KEY_COUNT) 
    return;

  if(event.type == SDL_KEYDOWN){
    _keys[key]._isDown = true;
    _keys[key]._isPressed = true;
    _history.push_back(key);
  }
  else{
    _keys[key]._isDown = false;
    _keys[key]._isReleased = true;
  }
}

void Input::onUpdate()
{
  for(auto& key : _keys)
    key._isPressed = key._isReleased = false;
  _history.clear();
}

Input::KeyCode Input::convertSdlKeyCode(int sdlCode)
{
  switch(sdlCode) {
    case SDLK_a: return KEY_a;
    case SDLK_b: return KEY_b;
    case SDLK_c: return KEY_c;
    case SDLK_d: return KEY_d;
    case SDLK_e: return KEY_e;
    case SDLK_f: return KEY_f;
    case SDLK_g: return KEY_g;
    case SDLK_h: return KEY_h;
    case SDLK_i: return KEY_i;
    case SDLK_j: return KEY_j;
    case SDLK_k: return KEY_k;
    case SDLK_l: return KEY_l;
    case SDLK_m: return KEY_m;
    case SDLK_n: return KEY_n;
    case SDLK_o: return KEY_o;
    case SDLK_p: return KEY_p;
    case SDLK_q: return KEY_q;
    case SDLK_r: return KEY_r;
    case SDLK_s: return KEY_s;
    case SDLK_t: return KEY_t;
    case SDLK_u: return KEY_u;
    case SDLK_v: return KEY_v;
    case SDLK_w: return KEY_w;
    case SDLK_x: return KEY_x;
    case SDLK_y: return KEY_y;
    case SDLK_z: return KEY_z;
    case SDLK_SPACE: return KEY_SPACE;
    case SDLK_BACKSPACE: return KEY_BACKSPACE;
    case SDLK_RETURN: return KEY_ENTER;
    case SDLK_LEFT: return KEY_LEFT;
    case SDLK_RIGHT: return KEY_RIGHT;
    case SDLK_DOWN: return KEY_DOWN;
    case SDLK_UP: return KEY_UP;
    default: return KEY_COUNT;
  }
}

int32_t Input::keyToAsciiCode(KeyCode key) const
{
  switch(key){
    case KEY_a: return 'a';
    case KEY_b: return 'b';
    case KEY_c: return 'c';
    case KEY_d: return 'd';
    case KEY_e: return 'e';
    case KEY_f: return 'f';
    case KEY_g: return 'g';
    case KEY_h: return 'h';
    case KEY_i: return 'i';
    case KEY_j: return 'j';
    case KEY_k: return 'k';
    case KEY_l: return 'l';
    case KEY_m: return 'm';
    case KEY_n: return 'n';
    case KEY_o: return 'o';
    case KEY_p: return 'p';
    case KEY_q: return 'q';
    case KEY_r: return 'r';
    case KEY_s: return 's';
    case KEY_t: return 't';
    case KEY_u: return 'u';
    case KEY_v: return 'v';
    case KEY_w: return 'w';
    case KEY_x: return 'x';
    case KEY_y: return 'y';
    case KEY_z: return 'z';
    case KEY_SPACE: return ' ';
    default: return -1;
  }
}

std::unique_ptr<Input> input {nullptr};

//===============================================================================================//
// ##>RESOURCES                                                                                  //
//===============================================================================================//

Dataset::Property::Property(int32_t key, std::string name, Value_t default_, Value_t min, Value_t max) :
  _key{key}, _name{name}, _default{default_}, _min{min}, _max{max}
{
  assert(_default.index() == _min.index() && _default.index() == _max.index());
  _value = unsetValue;
  _type = static_cast<Type>(_default.index());
}

Dataset::Dataset(std::initializer_list<Property> properties)
{
  for(const auto& property : properties)
    _properties.emplace(std::make_pair(property._key, property));
}

int32_t Dataset::load(const std::string& filename)
{
  std::ifstream file {filename};
  if(!file){
    log->log(Log::WARN, logstr::warn_cannot_open_dataset, filename);
    log->log(Log::INFO, logstr::info_using_property_defaults);
    for(auto& pair : _properties)
      pair.second._value = pair.second._default;
    return -1;
  }

  auto lineNoToString = [](int32_t l){return std::string{" ["} + std::to_string(l) + "] ";};
  auto isSpace = [](char c){return std::isspace<char>(c, std::locale::classic());};

  int32_t lineNo {0};
  int32_t nErrors {0};

  for(std::string line; std::getline(file, line);){
    ++lineNo;

    line.erase(std::remove_if(line.begin(), line.end(), isSpace), line.end());

    if(line.front() == comment) 
      continue;
     
    int32_t count {0};
    count = std::count(line.begin(), line.end(), seperator);
    if(count != 1){
      log->log(Log::WARN, logstr::warn_malformed_dataset, filename); 
      log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
      log->log(Log::INFO, logstr::info_unexpected_seperators, std::string{seperator});
      log->log(Log::INFO, logstr::info_ignoring_line);
      ++nErrors;
      continue;
    }

    std::size_t pos = line.find_first_of(seperator);
    std::string name {line.substr(0, pos)};
    std::string value {line.substr(pos + 1)};

    if(name.empty() || value.empty()){
      log->log(Log::ERROR, logstr::warn_malformed_dataset, filename); 
      log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
      log->log(Log::INFO, logstr::info_incomplete_property);
      log->log(Log::INFO, logstr::info_ignoring_line);
      ++nErrors;
      continue;
    }

    Property* p {nullptr};
    for(auto& pair : _properties){
      if(pair.second._name == name){
        p = &pair.second;
        break;
      }
    }

    if(p == nullptr){
      log->log(Log::WARN, logstr::warn_malformed_dataset, filename); 
      log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
      log->log(Log::INFO, logstr::info_unknown_dataset_property, name);
      log->log(Log::INFO, logstr::info_ignoring_line);
      continue;
    }

    switch(p->_type){
      case INT_PROPERTY:
        {
          int32_t result {0};
          if(!parseInt(value, result)){
            log->log(Log::WARN, logstr::warn_malformed_dataset, filename); 
            log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
            log->log(Log::INFO, logstr::info_expected_integer, value);
            log->log(Log::INFO, logstr::info_ignoring_line);
            ++nErrors;
            continue;
          }
          p->_value = std::clamp(result, std::get<int32_t>(p->_min), std::get<int32_t>(p->_max));
          if(std::get<int32_t>(p->_value) != result){
            log->log(Log::INFO, logstr::info_property_clamped, name);
          }
          log->log(Log::INFO, logstr::info_property_set, filename + lineNoToString(lineNo) + line);
          break;
        }
      case FLOAT_PROPERTY:
        {
          float result {0.f};
          if(!parseFloat(value, result)){
            log->log(Log::WARN, logstr::warn_malformed_dataset, filename); 
            log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
            log->log(Log::INFO, logstr::info_expected_float, value);
            log->log(Log::INFO, logstr::info_ignoring_line);
            ++nErrors;
            continue;
          }
          p->_value = std::clamp(result, std::get<float>(p->_min), std::get<float>(p->_max));
          if(std::get<float>(p->_value) != result){
            log->log(Log::INFO, logstr::info_property_clamped, name);
          }
          log->log(Log::INFO, logstr::info_property_set, filename + lineNoToString(lineNo) + line);
          break;
        }
      case BOOL_PROPERTY:
        {
          bool result {false};
          if(!parseBool(value, result)){
            log->log(Log::WARN, logstr::warn_malformed_dataset, filename); 
            log->log(Log::INFO, logstr::info_on_line, lineNoToString(lineNo) + line);
            log->log(Log::INFO, logstr::info_expected_bool, value);
            log->log(Log::INFO, logstr::info_ignoring_line);
            ++nErrors;
            continue;
          }
          p->_value = result;
          log->log(Log::INFO, logstr::info_property_set, filename + lineNoToString(lineNo) + line);
          break;
        }
    }
  }

  for(auto& pair : _properties){
    if(pair.second._value == unsetValue){
      log->log(Log::WARN, logstr::warn_property_not_set, pair.second._name);
      log->log(Log::INFO, logstr::info_using_property_defaults);
      pair.second._value = pair.second._default;
      ++nErrors;
    }
  }

  return nErrors;
}

int32_t Dataset::write(const std::string& filename, bool genComments)
{
  std::ofstream file {filename, std::ios_base::out | std::ios_base::trunc};
  if(!file){
    log->log(Log::WARN, logstr::warn_cannot_create_dataset, filename);
    return -1;
  }

  for(const auto& pair : _properties){
    const Property& p = pair.second;
    
    std::stringstream ss {};

    if(genComments){
      ss << comment;
      ss << " default=";
      printValue(p._default, ss);
      ss << " min=";
      printValue(p._min, ss);
      ss << " max=";
      printValue(p._max, ss);
      ss << '\n';
      file << ss.str();
    }

    std::stringstream().swap(ss);

    ss << p._name << "=";
    printValue(p._value, ss);
    ss << '\n';
    file << ss.str();
  }
}

int32_t Dataset::getIntValue(int32_t key) const
{
  assert(_properties.at(key)._type == INT_PROPERTY);
  return std::get<int32_t>(_properties.at(key)._value);
}

float Dataset::getFloatValue(int32_t key) const
{
  assert(_properties.at(key)._type == FLOAT_PROPERTY);
  return std::get<float>(_properties.at(key)._value);
}

bool Dataset::getBoolValue(int32_t key) const
{
  assert(_properties.at(key)._type == BOOL_PROPERTY);
  return std::get<bool>(_properties.at(key)._value);
}

void Dataset::setIntValue(int32_t key, int32_t value)
{
  assert(_properties.at(key)._type == INT_PROPERTY);
  _properties[key]._value = value;
}

void Dataset::setFloatValue(int32_t key, float value)
{
  assert(_properties.at(key)._type == FLOAT_PROPERTY);
  _properties[key]._value = value;
}

void Dataset::setBoolValue(int32_t key, bool value)
{
  assert(_properties.at(key)._type == BOOL_PROPERTY);
  _properties[key]._value = value;
}

void Dataset::scaleIntValue(int32_t key, int32_t scale)
{
  assert(_properties.at(key)._type == INT_PROPERTY);
  int32_t value = std::get<int32_t>(_properties[key]._value);
  _properties[key]._value = value * scale;
}

void Dataset::scaleFloatValue(int32_t key, float scale)
{
  assert(_properties.at(key)._type == FLOAT_PROPERTY);
  float value = std::get<float>(_properties[key]._value);
  _properties[key]._value = value * scale;
}

void Dataset::applyDefaults()
{
  for(auto& pair : _properties){
    pair.second._value = pair.second._default;
  }
}

bool Dataset::parseInt(const std::string& value, int32_t& result)
{
  auto isDigit = [](unsigned char c){return std::isdigit(c);};
  auto isSign = [](unsigned char c){return c == '+' || c == '-';};

  int32_t nSigns = std::count_if(value.begin(), value.end(), isSign);
  if(nSigns > 1){
    return false;
  }
  else if(nSigns == 1 && isSign(value.front()) == false){
      return false;
  }

  int32_t count = std::count_if(value.begin(), value.end(), isDigit);
  if(count != value.length() - nSigns){
    return false;
  }

  result = std::stoi(value);
  return true;
}

bool Dataset::parseFloat(const std::string& value, float& result)
{
  auto isDigit = [](unsigned char c){return std::isdigit(c);};
  auto isSign = [](unsigned char c){return c == '+' || c == '-';};

  int32_t nSigns = std::count_if(value.begin(), value.end(), isSign);
  if(nSigns > 1){
    return false;
  }
  else if(nSigns == 1 && isSign(value.front()) == false){
      return false;
  }

  int32_t nPoints = std::count(value.begin(), value.end(), '.');
  if(nPoints > 1)
    return false;

  int32_t count = std::count_if(value.begin(), value.end(), isDigit);
  if(count != value.length() - nPoints - nSigns)
    return false;

  result = std::stof(value);
  return true;
}

bool Dataset::parseBool(const std::string& value, bool& result)
{
  if(value == "true"){
    result = true;
    return true;
  }
  else if(value == "false"){
    result = false;
    return true;
  }

  return false;
}

void Dataset::printValue(const Value_t& value, std::ostream& os)
{
  switch(value.index()){
    case 0:
      os << std::get<int32_t>(value);
      break;
    case 1:
      os << std::get<float>(value);
      break;
    case 2:
      os << (std::get<bool>(value) ? "true" : "false");
      break;
  }
}

void Assets::loadBitmaps(const Manifest_t& manifest)
{
  for(const auto& tuple : manifest){
    Key_t key = std::get<0>(tuple);
    Name_t name = std::get<1>(tuple);
    Scale_t scale = std::get<2>(tuple);

    assert(scale < maxScale);

    auto search = _bitmaps.find(key);
    if(search == _bitmaps.end()){
      auto pair = _bitmaps.insert(std::make_pair(key, std::array<std::unique_ptr<Bitmap>, maxScale>{}));
      assert(pair.second);
      search = pair.first;
    }

    if((search->second)[scale] != nullptr){
      std::string addendum = std::string{name} + std::string{" : scale:"} + std::to_string(scale); 
      log->log(Log::WARN, logstr::warn_bitmap_already_loaded, std::move(addendum));
      log->log(Log::INFO, logstr::info_skipping_asset_loading);
      continue;
    }

    (search->second)[scale] = std::move(std::make_unique<Bitmap>(loadBitmap(bitmaps_path, name, scale)));
  }
}

Bitmap Assets::loadBitmap(std::string path, std::string name, Scale_t scale)
{
  assert(scale < maxScale);

  std::string bitpath {};
  bitpath += path;
  bitpath += path_seperator;
  bitpath += name;
  bitpath += bitmaps_extension;

  log->log(Log::INFO, logstr::info_loading_asset, bitpath);

  std::ifstream file {bitpath};
  if(!file){
    log->log(Log::WARN, logstr::warn_cannot_open_asset, bitpath);
    log->log(Log::INFO, logstr::info_using_error_bitmap);
    return generateErrorBitmap(scale);
  }

  auto isSpace = [](char c){return std::isspace<char>(c, std::locale::classic());};

  std::vector<std::string> rows {};

  for(std::string row; std::getline(file, row);){
    row.erase(std::remove_if(row.begin(), row.end(), isSpace), row.end());
    if(row.length() > 0)
      rows.emplace_back(std::move(row));
  }

  if(rows.size() == 0){
    log->log(Log::WARN, logstr::warn_empty_bitmap_file, bitpath);
    log->log(Log::INFO, logstr::info_using_error_bitmap);
    return generateErrorBitmap(scale);
  }

  auto isBinary = [](char ch){return ch == '0' || ch == '1';};

  int32_t rowNo {0};
  for(auto& row : rows){
    ++rowNo;

    int32_t count = std::count_if(row.begin(), row.end(), isBinary);
    if(count != row.length()){
      log->log(Log::WARN, logstr::warn_malformed_bitmap);
      std::string addendum = std::string{" ["} + std::to_string(rowNo) + std::string{"] "} + row;
      log->log(Log::INFO, logstr::info_on_row, addendum);
      return generateErrorBitmap(scale);
    }
  }

  std::reverse(rows.begin(), rows.end()); 

  Bitmap bitmap {};
  bitmap.initialize(std::move(rows), scale);

  return bitmap;
}

Bitmap Assets::generateErrorBitmap(Scale_t scale)
{
  std::vector<std::string> rows {};

  for(auto row : errorBits){
    rows.push_back(row); 
  }

  Bitmap bitmap {};
  bitmap.initialize(rows, scale);

  return bitmap;
}

void Assets::loadFonts(const Manifest_t& manifest)
{
  for(const auto& tuple : manifest){
    Key_t key = std::get<0>(tuple);
    Name_t name = std::get<1>(tuple);
    Scale_t scale = std::get<2>(tuple);

    assert(scale < maxScale);

    auto search = _fonts.find(key);
    if(search == _fonts.end()){
      auto pair = _fonts.insert(std::make_pair(key, std::array<std::unique_ptr<Font>, maxScale>{}));
      assert(pair.second);
      search = pair.first;
    }

    if((search->second)[scale] != nullptr){
      std::string addendum = std::string{name} + std::string{" : scale:"} + std::to_string(scale); 
      log->log(Log::WARN, logstr::warn_font_already_loaded, std::move(addendum));
      log->log(Log::INFO, logstr::info_skipping_asset_loading);
      continue;
    }

    (search->second)[scale] = std::move(std::make_unique<Font>(loadFont(name, scale)));
  }
}

Font Assets::loadFont(std::string name, Scale_t scale)
{
  Font::Meta meta;
  std::vector<Glyph> glyphs {};

  std::string basepath {};
  basepath += fonts_path;
  basepath += path_seperator;
  basepath += name;
  basepath += path_seperator;

  std::string fontpath {basepath};
  fontpath += name;
  fontpath += fonts_extension;

  log->log(Log::INFO, logstr::info_loading_asset, fontpath);

  FontData fontdata {};

  int32_t r = fontdata.load(fontpath);
  if(r != 0){
    const char* error = (r < 0) ? logstr::warn_cannot_open_asset : logstr::warn_asset_parse_errors;
    log->log(Log::WARN, error, fontpath);
    log->log(Log::INFO, logstr::info_using_error_font);
    return generateErrorFont(scale);
  }

  meta._lineSpace = fontdata.getIntValue(FontData::KEY_LINE_SPACE) * scale;
  meta._wordSpace = fontdata.getIntValue(FontData::KEY_WORD_SPACE) * scale;
  meta._glyphSpace = fontdata.getIntValue(FontData::KEY_GLYPH_SPACE) * scale;
  meta._size = fontdata.getIntValue(FontData::KEY_SIZE) * scale;

  GlyphData glyphdata {};

  for(int i = 0; i < asciiCharCount; ++i){
    int32_t ascii = static_cast<int32_t>('!') + i;

    std::string glyphpath {basepath};
    glyphpath += glyphFilenames[i];
    glyphpath += glyphs_extension;

    int32_t r = glyphdata.load(glyphpath);
    if(r != 0){
      const char* error = (r < 0) ? logstr::warn_cannot_open_asset : logstr::warn_asset_parse_errors;
      log->log(Log::WARN, error, glyphpath);
      log->log(Log::INFO, logstr::info_using_error_glyph);
      log->log(Log::INFO, logstr::info_ascii_code, std::to_string(ascii));
      glyphs.emplace_back(generateErrorGlyph(ascii, scale));
      continue;
    }

    Glyph glyph {};

    glyph._asciiCode = ascii;
    glyph._offsetX = glyphdata.getIntValue(GlyphData::KEY_OFFSET_X) * scale;
    glyph._offsetY = glyphdata.getIntValue(GlyphData::KEY_OFFSET_Y) * scale;
    glyph._advance = glyphdata.getIntValue(GlyphData::KEY_ADVANCE) * scale;
    glyph._width = glyphdata.getIntValue(GlyphData::KEY_WIDTH) * scale;
    glyph._height = glyphdata.getIntValue(GlyphData::KEY_HEIGHT) * scale;

    glyph._bitmap = loadBitmap(basepath, glyphFilenames[i], scale);

    glyphs.emplace_back(std::move(glyph));
  }

  Font font;
  font.initialize(meta, std::move(glyphs));

  return font;
}

Font Assets::generateErrorFont(Scale_t scale)
{
  Font::Meta meta;
  std::vector<Glyph> glyphs {};

  FontData fontdata {};
  fontdata.applyDefaults();

  meta._lineSpace = fontdata.getIntValue(FontData::KEY_LINE_SPACE) * scale;
  meta._wordSpace = fontdata.getIntValue(FontData::KEY_WORD_SPACE) * scale;
  meta._glyphSpace = fontdata.getIntValue(FontData::KEY_GLYPH_SPACE) * scale;
  meta._size = fontdata.getIntValue(FontData::KEY_SIZE) * scale;

  for(int i = 0; i < asciiCharCount; ++i){
    int32_t ascii = static_cast<int32_t>('!') + i;
    glyphs.emplace_back(generateErrorGlyph(ascii, scale));
  }

  Font font {};
  font.initialize(meta, std::move(glyphs));
  return font;
}

Glyph Assets::generateErrorGlyph(int32_t asciiCode, Scale_t scale)
{
  GlyphData glyphdata {};
  glyphdata.applyDefaults();

  Glyph glyph {};

  glyph._asciiCode = asciiCode;
  glyph._offsetX = glyphdata.getIntValue(GlyphData::KEY_OFFSET_X) * scale;
  glyph._offsetY = glyphdata.getIntValue(GlyphData::KEY_OFFSET_Y) * scale;
  glyph._advance = glyphdata.getIntValue(GlyphData::KEY_ADVANCE) * scale;
  glyph._width = glyphdata.getIntValue(GlyphData::KEY_WIDTH) * scale;
  glyph._height = glyphdata.getIntValue(GlyphData::KEY_HEIGHT) * scale;

  std::vector<std::string> bits {};
  for(auto row : errorBits)
    bits.push_back(row);

  glyph._bitmap.initialize(std::move(bits), scale);

  return glyph;
}

const Bitmap& Assets::getBitmap(Key_t key, Scale_t scale) const
{
  return *((_bitmaps.at(key))[scale]);
}

const Font& Assets::getFont(Key_t key, Scale_t scale) const
{
  return *((_fonts.at(key))[scale]);
}

Bitmap Assets::makeBlockBitmap(int32_t width, int32_t height)
{
  std::vector<std::string> bits;
  std::string bitrow {};
  for(int32_t col = 0; col < width; ++col)
    bitrow += '1';
  for(int32_t row = 0; row < height; ++row)
    bits.push_back(bitrow);
  Bitmap b {};
  b.initialize(bits);
  return b;
}

std::unique_ptr<Assets> assets {nullptr};

//===============================================================================================//
// ##>GRAPHICS                                                                                   //
//===============================================================================================//

void Bitmap::initialize(std::vector<std::string> bits, int32_t scale)
  // predicate: bit strings must contain only 0's and 1's
{
  // Strip trailing 0's on all rows leaving atleast one 0 on rows consisting of all zeros. 
  // This permits 'padding rows' to be created in bitmaps whilst stripping surplus data.
  for(auto& row : bits){
    while(row.back() == '0' && row.size() != 1){
      row.pop_back();
    }
  }

  // scale bits
  if(scale > 1){
    std::vector<std::string> sbits {};
    std::string srow {};
    for(auto& row : bits){
      for(auto c : row){
        for(int i = 0; i < scale; ++i){
          srow.push_back(c);
        }
      }
      row.shrink_to_fit();
      for(int i = 0; i < scale; ++i){
        sbits.push_back(srow);
      }
      srow.clear();
    }
    bits = std::move(sbits);
  }

  // compute bitmap dimensions
  size_t w {0};
  for(const auto& row : bits){
    w = std::max(w, row.length());
  }
  _width = w;
  _height = bits.size();

  // generate the bit data
  std::vector<bool> brow {};
  brow.reserve(w);
  for(const auto& row : bits){
    brow.clear();
    for(char c : row){
      bool bit = !(c == '0'); 
      brow.push_back(bit); 
    }
    int32_t n = w - row.size();
    while(n-- > 0){
      brow.push_back(false); // pad to make all rows equal length
    }
    brow.shrink_to_fit();
    _bits.push_back(brow);
  }

  regenerateBytes();
}

bool Bitmap::getBit(int32_t row, int32_t col) const
{
  assert(0 <= row && row < _height);
  assert(0 <= col && col < _width);
  return _bits[row][col];
}

void Bitmap::setBit(int32_t row, int32_t col, bool value, bool regen)
{
  assert(0 <= row && row < _height);
  assert(0 <= col && col < _width);
  _bits[row][col] = value;
  if(regen)
    regenerateBytes();
}

void Bitmap::setRect(int32_t rowMin, int32_t colMin, int32_t rowMax, int32_t colMax, bool value, bool regen)
{
  // note - inclusive range of rows and columns, i.e. [rowMin, rowMax] and [colMin, colMax]
  
  assert(rowMin >= 0 && rowMax < _height);
  assert(colMin >= 0 && colMax < _width);

  for(int32_t row = rowMin; row <= rowMax; ++row)
    for(int32_t col = colMin; col <= colMax; ++col)
      _bits[row][col] = value;

  if(regen)
    regenerateBytes();
}

void Font::initialize(Meta meta, std::vector<Glyph> glyphs)
  // predicate: expects glyphs for all ascii codes in the range 33-126 inclusive.
  // predicate: expects glyphs to be in ascending order of ascii code.
{
  _meta = meta;
  _glyphs = std::move(glyphs);
}

void Bitmap::regenerateBytes()
{
  _bytes.clear();
  uint8_t byte {0};
  int32_t bitNo {0};
  for(const auto& row : _bits){
    for(bool bit : row){
      if(bitNo > 7){
        _bytes.push_back(byte);
        byte = 0;
        bitNo = 0;
      }
      if(bit) 
        byte |= 0x01 << (7 - bitNo);
      bitNo++;
    }
    _bytes.push_back(byte);
    byte = 0;
    bitNo = 0;
  }
}

int32_t Font::calculateStringWidth(const std::string& str) const
{
  int32_t sum {0};
  for(char c : str){
    if(c == ' ') 
      sum += getWordSpace();
    else 
      sum += getGlyph(c)._advance + _meta._glyphSpace;
  }
  return sum;
}

bool Bitmap::isEmpty()
{
  for(const auto& row : _bits)
    for(const auto& bit : row)
      if(bit)
        return false;

  return true;
}

bool Bitmap::isApproxEmpty(int32_t threshold)
{
  int32_t count {0};
  for(const auto& row : _bits){
    for(const auto& bit : row){
      if(bit){
        ++count;
        if(count > threshold)
          return false;
      }
    }
  }
  return true;
}

void Bitmap::print(std::ostream& out) const
{
  for(auto iter = _bits.rbegin(); iter != _bits.rend(); ++iter){
    for(bool bit : *iter){
      out << bit;
    }
    out << '\n';
  }
  out << std::endl;
}

Renderer::Renderer(const Config& config)
{
  _config = config;

  uint32_t flags = SDL_WINDOW_OPENGL;
  if(_config._fullscreen){
    flags |= SDL_WINDOW_FULLSCREEN_DESKTOP;
    log->log(Log::INFO, logstr::info_fullscreen_mode);
  }

  std::stringstream ss {};
  ss << "{w:" << _config._windowWidth << ",h:" << _config._windowHeight << "}";
  log->log(Log::INFO, logstr::info_creating_window, std::string{ss.str()});

  _window = SDL_CreateWindow(
      _config._windowTitle.c_str(), 
      SDL_WINDOWPOS_UNDEFINED,
      SDL_WINDOWPOS_UNDEFINED,
      _config._windowWidth,
      _config._windowHeight,
      flags
  );

  if(_window == nullptr){
    log->log(Log::FATAL, logstr::fail_create_window, std::string{SDL_GetError()});
    exit(EXIT_FAILURE);
  }

  Vector2i wndsize = getWindowSize();
  std::stringstream().swap(ss);
  ss << "{w:" << wndsize._x << ",h:" << wndsize._y << "}";
  log->log(Log::INFO, logstr::info_created_window, std::string{ss.str()});

  _glContext = SDL_GL_CreateContext(_window);
  if(_glContext == nullptr){
    log->log(Log::FATAL, logstr::fail_create_opengl_context, std::string{SDL_GetError()});
    exit(EXIT_FAILURE);
  }

  if(SDL_GL_SetAttribute(SDL_GL_CONTEXT_MAJOR_VERSION, _config._openglVersionMajor) < 0){
    log->log(Log::FATAL, logstr::fail_set_opengl_attribute, std::string{SDL_GetError()});
    exit(EXIT_FAILURE);
  }
  if(SDL_GL_SetAttribute(SDL_GL_CONTEXT_MINOR_VERSION, _config._openglVersionMinor) < 0){
    log->log(Log::FATAL, logstr::fail_set_opengl_attribute, std::string{SDL_GetError()});
    exit(EXIT_FAILURE);
  }

  glPixelStorei(GL_UNPACK_ALIGNMENT, 1);
  setViewport(iRect{0, 0, _config._windowWidth, _config._windowHeight});
}

Renderer::~Renderer()
{
  SDL_GL_DeleteContext(_glContext);
  SDL_DestroyWindow(_window);
}

void Renderer::setViewport(iRect viewport)
{
  glMatrixMode(GL_PROJECTION);
  glLoadIdentity();
  glOrtho(0.0, viewport._w, 0.0, viewport._h, -1.0, 1.0);
  glMatrixMode(GL_MODELVIEW);
  glLoadIdentity();
  glViewport(viewport._x, viewport._y, viewport._w, viewport._h);
  _viewport = viewport;
}

void Renderer::blitText(Vector2f position, const std::string& text,  const Font& font, const Color3f& color)
{
  glColor3f(color.getRed(), color.getGreen(), color.getBlue());  
  glRasterPos2f(position._x, position._y);

  for(char c : text){
    if(!(' ' <= c && c <= '~'))
      continue;

    if(c == ' '){
      glBitmap(0, 0, 0, 0, font.getWordSpace(), 0, nullptr);
    }
    else{
      const Glyph& g = font.getGlyph(c);      
      glBitmap(g._width, g._height, g._offsetX, g._offsetY, g._advance + font.getGlyphSpace(), 0, g._bitmap.getBytes().data());
    }
  }
}

void Renderer::blitBitmap(Vector2f position, const Bitmap& bitmap, const Color3f& color)
{
  // For viewports which are a subegion of the window the glBitmap function will overdraw
  // the viewport bounds if the bitmap position is within the viewport but the bitmap itself
  // partially falls outside the viewport. Hence will clip any overdraw manually.
  if(position._x + bitmap.getWidth() > _viewport._w)
    return;

  if(position._y + bitmap.getHeight() > _viewport._h)
    return;

  glColor3f(color.getRed(), color.getGreen(), color.getBlue());  
  glRasterPos2f(position._x, position._y);
  glBitmap(bitmap.getWidth(), bitmap.getHeight(), 0, 0, 0, 0, bitmap.getBytes().data());
}

void Renderer::drawBorderRect(const iRect& rect, const Color3f& background, const Color3f& borderColor, int32_t borderWidth)
{
  int32_t x1, y1, x2, y2;
  x1 = rect._x - borderWidth;
  y1 = rect._y - borderWidth;
  x2 = rect._x + rect._w + borderWidth;
  y2 = rect._y + rect._h + borderWidth;
  glColor3f(borderColor.getRed(), borderColor.getGreen(), borderColor.getBlue());  
  glRecti(x1, y1, x2, y2);
  x1 += borderWidth;
  y1 += borderWidth;
  x2 -= borderWidth;
  y2 -= borderWidth;
  glColor3f(background.getRed(), background.getGreen(), background.getBlue());  
  glRecti(x1, y1, x2, y2);
}

void Renderer::clearWindow(const Color3f& color)
{
  glClearColor(color.getRed(), color.getGreen(), color.getBlue(), 1.f);
  glClear(GL_COLOR_BUFFER_BIT);
}

void Renderer::clearViewport(const Color3f& color)
{
  glEnable(GL_SCISSOR_TEST);
  glScissor(_viewport._x, _viewport._y, _viewport._w, _viewport._h);
  glClearColor(color.getRed(), color.getGreen(), color.getBlue(), 1.f);
  glClear(GL_COLOR_BUFFER_BIT);
  glDisable(GL_SCISSOR_TEST);
}

void Renderer::show()
{
  SDL_GL_SwapWindow(_window);
}

Vector2i Renderer::getWindowSize() const
{
  Vector2i size;
  SDL_GL_GetDrawableSize(_window, &size._x, &size._y);
  return size;
}

std::unique_ptr<Renderer> renderer {nullptr};

//===============================================================================================//
// ##>COLLISION DETECTION                                                                        //
//===============================================================================================//

static bool isAABBIntersection(const AABB& a, const AABB& b)
{
  return ((a._xmin <= b._xmax) && (a._xmax >= b._xmin)) && ((a._ymin <= b._ymax) && (a._ymax >= b._ymin));
}

static void calculateAABBOverlap(const AABB& aBounds, const AABB& bBounds, 
                                 AABB& aOverlap, AABB& bOverlap)
{
  // Overlap w.r.t screen space which is common to both.
  AABB overlap;
  overlap._xmax = std::min(aBounds._xmax, bBounds._xmax);
  overlap._xmin = std::max(aBounds._xmin, bBounds._xmin);
  overlap._ymax = std::min(aBounds._ymax, bBounds._ymax);
  overlap._ymin = std::max(aBounds._ymin, bBounds._ymin);

  // Overlaps w.r.t each local bitmap coordinate space.
  aOverlap._xmax = overlap._xmax - aBounds._xmin; 
  aOverlap._xmin = overlap._xmin - aBounds._xmin;
  aOverlap._ymax = overlap._ymax - aBounds._ymin; 
  aOverlap._ymin = overlap._ymin - aBounds._ymin;

  bOverlap._xmax = overlap._xmax - bBounds._xmin; 
  bOverlap._xmin = overlap._xmin - bBounds._xmin;
  bOverlap._ymax = overlap._ymax - bBounds._ymin; 
  bOverlap._ymin = overlap._ymin - bBounds._ymin;

  // The results should be the same overlap region w.r.t two different coordinate spaces.
  assert((aOverlap._xmax - aOverlap._xmin) == (bOverlap._xmax - bOverlap._xmin));
  assert((aOverlap._ymax - aOverlap._ymin) == (bOverlap._ymax - bOverlap._ymin));
}

static void findPixelIntersectionSets(const AABB& aOverlap, const Bitmap& aBitmap, 
                                      const AABB& bOverlap, const Bitmap& bBitmap,
                                      std::vector<Vector2i>& aPixels, std::vector<Vector2i>& bPixels,
                                      bool pixelLists)
{
  int32_t overlapWidth = aOverlap._xmax - aOverlap._xmin;
  int32_t overlapHeight = aOverlap._ymax - aOverlap._ymin;

  int32_t aBitRow, bBitRow, aBitCol, bBitCol, aBitValue, bBitValue;

  for(int32_t row = 0; row < overlapHeight; ++row){
    for(int32_t col = 0; col < overlapWidth; ++col){
      aBitRow = aOverlap._ymin + row;
      aBitCol = aOverlap._xmin + col;

      bBitRow = bOverlap._ymin + row;
      bBitCol = bOverlap._xmin + col;

      aBitValue = aBitmap.getBit(aBitRow, aBitCol);
      bBitValue = bBitmap.getBit(bBitRow, bBitCol);

      if(aBitValue == 0 || bBitValue == 0)
        continue;

      aPixels.push_back({aBitCol, aBitRow});
      bPixels.push_back({bBitCol, bBitRow});

      if(!pixelLists)
        return;
    }
  }
}

const Collision& testCollision(Vector2i aPosition, const Bitmap& aBitmap, 
                               Vector2i bPosition, const Bitmap& bBitmap, bool pixelLists)
{
  static Collision c;

  c._isCollision = false;
  c._aOverlap = {0, 0, 0, 0};
  c._bOverlap = {0, 0, 0, 0};
  c._aPixels.clear();
  c._bPixels.clear();

  c._aBounds = {aPosition._x, aPosition._y, 
                aPosition._x + aBitmap.getWidth(), aPosition._y + aBitmap.getHeight()};
  c._bBounds = {bPosition._x, bPosition._y, 
                bPosition._x + bBitmap.getWidth(), bPosition._y + bBitmap.getHeight()};

  if(!isAABBIntersection(c._aBounds, c._bBounds))
    return c;

  calculateAABBOverlap(c._aBounds, c._bBounds, c._aOverlap, c._bOverlap);

  findPixelIntersectionSets(c._aOverlap, aBitmap, 
                            c._bOverlap, bBitmap, 
                            c._aPixels, c._bPixels, pixelLists);

  assert(c._aPixels.size() == c._bPixels.size());

  if(c._aPixels.size() != 0)
    c._isCollision = true;  

  return c;
}

//===============================================================================================//
// ##>MIXER                                                                                      //
//===============================================================================================//

Mixer::Mixer()
{
  if(Mix_OpenAudio(sampleFreq, sampleFormat, numOutChannels, chunkSize) != 0){
    log->log(Log::FATAL, logstr::fail_open_audio, std::string{Mix_GetError()});
    exit(EXIT_FAILURE);
  }

  Mix_AllocateChannels(numMixChannels);
}

Mixer::~Mixer()
{
  stopChannel(allChannels);

  for(auto pair : _sounds)
    Mix_FreeChunk(pair.second);

  Mix_CloseAudio();
}

void Mixer::loadSoundsWAV(const Manifest_t& manifest)
{
  std::string path {};
  Mix_Chunk* chunk {nullptr};
  for(auto pair : manifest){
    path.clear();
    path += sounds_path;
    path += pair.second;
    path += sounds_extension;
    chunk = Mix_LoadWAV(path.c_str());
    if(!chunk){
      log->log(Log::WARN, logstr::warn_cannot_load_sound, std::string{Mix_GetError()});
      continue;
    }
    log->log(Log::INFO, logstr::info_loaded_sound, path + ", key:" + std::to_string(pair.first));
    _sounds.emplace(std::make_pair(pair.first, chunk));
    chunk = nullptr;
  }
}

Mixer::Channel_t Mixer::playSound(Key_t sndkey, int loops)
{
  Mix_Chunk* chunk = findChunk(sndkey);
  if(chunk == nullptr) return -1;
  Channel_t channel = Mix_PlayChannel(-1, chunk, loops);
  if(channel == -1){
    log->log(Log::WARN, logstr::warn_cannot_play_sound, std::string{sndkey});
    return nullChannel;
  }
  return channel;
}

Mixer::Channel_t Mixer::playSoundTimed(Key_t sndkey, int loops, int timeLimit_ms)
{
  Mix_Chunk* chunk = findChunk(sndkey);
  if(chunk == nullptr) return -1;
  Channel_t channel = Mix_PlayChannelTimed(-1, chunk, loops, timeLimit_ms);
  if(channel == -1){
    log->log(Log::WARN, logstr::warn_cannot_play_sound, std::string{sndkey});
    return nullChannel;
  }
  return channel;
}

Mixer::Channel_t Mixer::playSoundFadeIn(Key_t sndkey, int loops, int fadeInTime_ms)
{
  Mix_Chunk* chunk = findChunk(sndkey);
  if(chunk == nullptr) return -1;
  int channel = Mix_FadeInChannel(-1, chunk, loops, fadeInTime_ms);
  if(channel == -1){
    log->log(Log::WARN, logstr::warn_cannot_play_sound, std::string{sndkey});
    return nullChannel;
  }
  return channel;
}

Mixer::Channel_t Mixer::playSoundFadeInTimed(Key_t sndkey, int loops, int fadeInTime_ms, int timeLimit_ms)
{
  Mix_Chunk* chunk = findChunk(sndkey);
  if(chunk == nullptr) return -1;
  int channel = Mix_FadeInChannelTimed(-1, chunk, loops, fadeInTime_ms, timeLimit_ms);
  if(channel == -1){
    log->log(Log::WARN, logstr::warn_cannot_play_sound, std::string{sndkey});
    return nullChannel;
  }
  return channel;
}

void Mixer::stopChannel(Channel_t channel)
{
  if(channel == nullChannel) return;
  Mix_HaltChannel(channel);
}

void Mixer::pauseChannel(Channel_t channel)
{
  if(channel == nullChannel) return;
  Mix_Pause(channel);
}

void Mixer::resumeChannel(Channel_t channel)
{
  if(channel == nullChannel) return;
  Mix_Resume(channel);
}

void Mixer::setVolume(float volume)
{
  int ivolume = std::clamp(volume, 0.f, 1.f) * MIX_MAX_VOLUME;
  ivolume = Mix_Volume(allChannels, ivolume); // returns average volume of all mixing channels.
  _volume = static_cast<float>(ivolume) / MIX_MAX_VOLUME;
}

Mix_Chunk* Mixer::findChunk(Key_t sndkey)
{
  auto search = _sounds.find(sndkey);
  if(search == _sounds.end()){
    log->log(Log::WARN, logstr::warn_missing_sound, std::to_string(sndkey));
    return nullptr;
  }
  return search->second;
}

std::unique_ptr<Mixer> mixer;

//===============================================================================================//
// ##>UI                                                                                         //
//===============================================================================================//

HUD::TextLabel::TextLabel(Vector2i p, Color3f c, std::string t, float activeDelay, bool phase, bool flash) :
  _position{p},
  _color{c},
  _text{t},
  _value{},
  _charNo{0},
  _activeDelay{activeDelay},
  _activeTime{0},
  _isActive{false},
  _isVisible{true},
  _isHidden{false},
  _phase{phase},
  _flash{flash}
{
  _value.reserve(_text.length());
  if(!_phase)
    _value = _text;
}

HUD::IntLabel::IntLabel(Vector2i p, Color3f c, const int32_t* s, int32_t precision, float activeDelay, bool flash) :
  _position{p},
  _color{c},
  _source{s},
  _precision{precision},
  _value{-1},
  _text{std::to_string(_value)},
  _activeDelay{activeDelay},
  _activeTime{0},
  _isActive{false},
  _isVisible{true},
  _isHidden{false},
  _flash{flash}
{}

HUD::BitmapLabel::BitmapLabel(Vector2i p, Color3f c, const Bitmap* b, float activeDelay, bool flash) :
  _position{p},
  _color{c},
  _bitmap{b},
  _activeDelay{activeDelay},
  _activeTime{0},
  _isActive{false},
  _isVisible{true},
  _isHidden{false},
  _flash{flash}
{}

void HUD::initialize(const Font* font, float flashPeriod, float phasePeriod)
{
  _font = font;
  _nextUid = 0;
  _flashPeriod = flashPeriod;
  _phasePeriod = phasePeriod;
  _masterClock = 0.f;
  _flashClock = 0.f;
  _phaseClock = 0.f;
}

HUD::uid_t HUD::addTextLabel(TextLabel label)
{
  label._uid = _nextUid++;
  label._activeTime = _masterClock + label._activeDelay;
  _textLabels.push_back(std::move(label));
  return label._uid;
}

HUD::uid_t HUD::addIntLabel(IntLabel label)
{
  label._uid = _nextUid++;
  label._activeTime = _masterClock + label._activeDelay;
  _intLabels.push_back(std::move(label));
  return label._uid;
}

HUD::uid_t HUD::addBitmapLabel(BitmapLabel label)
{
  label._uid = _nextUid++;
  label._activeTime = _masterClock + label._activeDelay;
  _bitmapLabels.push_back(std::move(label));
  return label._uid;
}

bool HUD::removeTextLabel(uid_t uid)
{
  for(auto iter = _textLabels.begin(); iter != _textLabels.end(); ++iter){
    if((*iter)._uid == uid){
      _textLabels.erase(iter);
      return true;
    }
  }
  return false;
}

bool HUD::removeIntLabel(uid_t uid)
{
  for(auto iter = _intLabels.begin(); iter != _intLabels.end(); ++iter){
    if((*iter)._uid == uid){
      _intLabels.erase(iter);
      return true;
    }
  }
  return false;
}

bool HUD::removeBitmapLabel(uid_t uid)
{
  for(auto iter = _bitmapLabels.begin(); iter != _bitmapLabels.end(); ++iter){
    if((*iter)._uid == uid){
      _bitmapLabels.erase(iter);
      return true;
    }
  }
  return false;
}

void HUD::clear()
{
  _textLabels.clear();
  _intLabels.clear();
  _bitmapLabels.clear();
  _nextUid = 0;
}

void HUD::hideTextLabel(uid_t uid)
{
  for(auto& label : _textLabels)
    if(label._uid == uid)
      label._isHidden = true;
}

void HUD::hideIntLabel(uid_t uid)
{
  for(auto& label : _intLabels)
    if(label._uid == uid)
      label._isHidden = true;
}

void HUD::hideBitmapLabel(uid_t uid)
{
  for(auto& label : _bitmapLabels)
    if(label._uid == uid)
      label._isHidden = true;
}

void HUD::showTextLabel(uid_t uid)
{
  for(auto& label : _textLabels)
    if(label._uid == uid)
      label._isHidden = false;
}

void HUD::showIntLabel(uid_t uid)
{
  for(auto& label : _intLabels)
    if(label._uid == uid)
      label._isHidden = false;
}

void HUD::showBitmapLabel(uid_t uid)
{
  for(auto& label : _bitmapLabels)
    if(label._uid == uid)
      label._isHidden = false;
}

void HUD::startTextLabelFlash(uid_t uid)
{
  for(auto& label : _textLabels)
    if(label._uid == uid)
      label._flash = true;
}

void HUD::startIntLabelFlash(uid_t uid)
{
  for(auto& label : _intLabels)
    if(label._uid == uid)
      label._flash = true;
}

void HUD::startBitmapLabelFlash(uid_t uid)
{
  for(auto& label : _bitmapLabels)
    if(label._uid == uid)
      label._flash = true;
}

void HUD::stopTextLabelFlash(uid_t uid)
{
  for(auto& label : _textLabels){
    if(label._uid == uid){
      label._flash = false;
      label._isVisible = true;
    }
  }
}

void HUD::stopIntLabelFlash(uid_t uid)
{
  for(auto& label : _intLabels){
    if(label._uid == uid){
      label._flash = false;
      label._isVisible = true;
    }
  }
}

void HUD::stopBitmapLabelFlash(uid_t uid)
{
  for(auto& label : _bitmapLabels){
    if(label._uid == uid){
      label._flash = false;
      label._isVisible = true;
    }
  }
}

void HUD::onReset()
{
  for(auto& label : _textLabels){
    label._charNo = 0;
    label._activeTime = label._activeDelay;
    label._isActive = false;
    label._isVisible = true;
    label._isHidden = false;
    if(label._phase)
      label._value = std::string{};
  }

  for(auto& label : _intLabels){
    label._value = *(label._source);
    label._text = std::to_string(label._value);
    label._activeTime = label._activeDelay;
    label._isActive = false;
    label._isVisible = true;
    label._isHidden = false;
  }

  for(auto& label : _bitmapLabels){
    label._activeTime = label._activeDelay;
    label._isActive = false;
    label._isVisible = true;
    label._isHidden = false;
  }

  _masterClock = 0.f;
  _phaseClock = 0.f;
  _flashClock = 0.f;
}

void HUD::onUpdate(float dt)
{
  _masterClock += dt;

  _flashClock += dt;
  if(_flashClock >= _flashPeriod){
    flashLabels();
    _flashClock = 0.f;
  }

  _phaseClock += dt;
  if(_phaseClock >= _phasePeriod){
    phaseLabels();
    _phaseClock = 0.f;
  }

  updateIntLabels();
  activateLabels();
}

void HUD::onDraw()
{
  for(auto& label : _textLabels)
    if(label._isActive && label._isVisible && !label._isHidden)
      pxr::renderer->blitText(label._position, label._value, *_font, label._color);

  for(auto& label : _intLabels){
    if(label._isActive && label._isVisible && !label._isHidden)
      pxr::renderer->blitText(label._position, label._text, *_font, label._color);
  }

  for(auto& label : _bitmapLabels)
    if(label._isActive && label._isVisible && !label._isHidden)
      pxr::renderer->blitBitmap(label._position, *(label._bitmap), label._color);
}

void HUD::flashLabels()
{
  for(auto& label : _textLabels)
    if(label._isActive && label._flash)
      label._isVisible = !label._isVisible;

  for(auto& label : _intLabels)
    if(label._isActive && label._flash)
      label._isVisible = !label._isVisible;

  for(auto& label : _bitmapLabels)
    if(label._isActive && label._flash)
      label._isVisible = !label._isVisible;
}

void HUD::phaseLabels()
{
  for(auto& label : _textLabels){
    if(label._isActive && label._phase && label._charNo != label._text.length()){
      label._value += label._text[label._charNo];
      label._charNo++;
    }
  }
}

void HUD::updateIntLabels()
{
  for(auto& label : _intLabels){
    if(label._isActive && label._value != *(label._source)){
      label._value = *(label._source);
      std::string text = std::to_string(label._value); 
      label._text = std::string{};
      int sign {0};
      if(label._value < 0){
        assert(text.front() == '-');
        sign = 1; 
        label._text += '-';
      }
      for(int i = 0; i < label._precision - (text.length() - sign); ++i)
        label._text += '0';
      label._text.append(text.begin() + sign, text.end());
    }
  }
}

void HUD::activateLabels()
{
  for(auto& label : _textLabels)
    if(!label._isActive && label._activeTime < _masterClock)
      label._isActive = true;

  for(auto& label : _intLabels)
    if(!label._isActive && label._activeTime < _masterClock)
      label._isActive = true;

  for(auto& label : _bitmapLabels)
    if(!label._isActive && label._activeTime < _masterClock)
      label._isActive = true;
}

TextInput::TextInput(Font* font, std::string label, Color3f cursorColor)
{

}

const char* TextInput::processInput()
{
  bool isDone {false};
  auto& keyHistory = input->getHistory();
  for(auto key : keyHistory){
    if(key == Input::KEY_RIGHT){
      ++_cursorPos;
      if(_cursorPos >= _bufferSize)
        _cursorPos = _bufferSize - 1;
    }
    else if(key == Input::KEY_LEFT){
      --_cursorPos;
      if(_cursorPos < 0)
        _cursorPos = 0;
    }
    else if(key == Input::KEY_ENTER){
      isDone = true;
      break;
    }
    else if(key == Input::KEY_BACKSPACE){

      //
      // TODO TODO
      //

    }
    else if(Input::KEY_a <= key && key <= Input::KEY_z){
      _buffer[_cursorPos] = input->keyToAsciiCode(key);
      ++_cursorPos;
      if(_cursorPos >= _bufferSize)
        _cursorPos = _bufferSize - 1;
    }
  }
  return (isDone) ? _buffer.data() : nullptr;
}



//===============================================================================================//
// ##>APPLICATION                                                                                //
//===============================================================================================//

void Application::onWindowResize(int32_t windowWidth, int32_t windowHeight)
{
  Vector2i worldSize = getWorldSize();

  if((windowWidth < worldSize._x) || (windowHeight < worldSize._y)){
    _isWindowTooSmall = true;
    _engine->pause();
    _viewport._x = 0;
    _viewport._y = 0;
    _viewport._w = windowWidth;
    _viewport._h = windowHeight;
  }
  else{
    _isWindowTooSmall = false;
    _engine->unpause();
    _viewport._x = (windowWidth - worldSize._x) / 2;
    _viewport._y = (windowHeight - worldSize._y) / 2;
    _viewport._w = worldSize._x;
    _viewport._h = worldSize._y;
  }
}

bool Application::initialize(Engine* engine, int32_t windowWidth, int32_t windowHeight)
{
  _engine = engine;
  return true;
}

void Application::onUpdate(double now, float dt)
{
  assert(_activeState != nullptr);
  (*_activeState)->onUpdate(now, dt);
}

void Application::onDraw(double now, float dt)
{
  assert(_activeState != nullptr);

  pxr::renderer->setViewport(_viewport);

  if(_isWindowTooSmall)
    drawWindowTooSmall();
  else
    (*_activeState)->onDraw(now, dt);
}

void Application::addState(std::unique_ptr<ApplicationState>&& state)
{
  _states.emplace(std::make_pair(state->getName(), std::move(state)));
}

void Application::switchState(const std::string& name)
{
  assert(_states.find(name) != _states.end());
  _activeState = &_states[name];
  (*_activeState)->onEnter();
}

void Application::drawWindowTooSmall()
{
  pxr::renderer->clearViewport(colors::red);
}

//===============================================================================================//
// ##>ENGINE                                                                                     //
//===============================================================================================//

Engine::Duration_t Engine::RealClock::update()
{
  _now1 = Clock_t::now();
  _dt = _now1 - _now0;
  _now0 = _now1;
  return _dt;
}

Engine::Duration_t Engine::GameClock::update(Duration_t realDt)
{
  if(_isPaused) return Duration_t::zero();
  _dt = Duration_t{static_cast<int64_t>(realDt.count() * _scale)};
  _now += _dt;
  return _dt;
}

int64_t Engine::Metronome::doTicks(Duration_t gameNow)
{
  int ticks {0};
  while(_lastTickNow + _tickPeriod < gameNow) {
    _lastTickNow += _tickPeriod;
    ++ticks;
  }
  _totalTicks += ticks;
  return ticks;
}

void Engine::TPSMeter::recordTicks(Duration_t realDt, int32_t ticks)
{
  _timer += realDt;
  _ticks += ticks;
  if(_timer > oneSecond){
    _tps = _ticks;
    _ticks = 0;
    _timer = Duration_t::zero();
  }
}

void Engine::initialize(std::unique_ptr<Application> app)
{
  log = std::make_unique<Log>();

  // 
  // Testing the seed_seq on my system shows it just produces the same results with every run, 
  // which is obviously useless. However I get different results with std::random_device so have 
  // opted to use that instead. I am lead to believe however that this may differ on different 
  // systems. TODO Will have to test it.
  //
  // std::seed_seq seq{1, 2, 3, 4, 5};
  // randGenerator.seed(seq);
  //
  std::random_device rd{};
  xorwow::state_type seedstate {};
  for(auto& seed : seedstate)
    seed = rd();
  randGenerator.seed(seedstate);

  _app = std::move(app);

  if(_config.load(Config::filename) != 0)
    _config.write(Config::filename); // generate a default file if one doesn't exist.

  if(SDL_Init(SDL_INIT_VIDEO | SDL_INIT_AUDIO) != 0){
    log->log(Log::FATAL, logstr::fail_sdl_init, std::string{SDL_GetError()});
    exit(EXIT_FAILURE);
  }

  LoopTick* tick = &_loopTicks[LOOPTICK_UPDATE];
  tick->_onTick = &Engine::onUpdateTick;
  tick->_metronome.setTickPeriod(Duration_t{static_cast<int64_t>(1.0e9 / 60.0_hz)});
  tick->_ticksAccumulated = 0;
  tick->_ticksDoneThisFrame = 0;
  tick->_maxTicksPerFrame = 5;
  tick->_tickPeriod = 1.0 / 60.0_hz;

  tick = &_loopTicks[LOOPTICK_DRAW];
  tick->_onTick = &Engine::onDrawTick;
  tick->_metronome.setTickPeriod(Duration_t{static_cast<int64_t>(1.0e9 / 60.0_hz)});
  tick->_ticksAccumulated = 0;
  tick->_ticksDoneThisFrame = 0;
  tick->_maxTicksPerFrame = 1;
  tick->_tickPeriod = 1.0 / 60.0_hz;

  std::stringstream ss {};
  ss << _app->getName() 
     << " version:" 
     << _app->getVersionMajor() 
     << "." 
     << _app->getVersionMinor();

  Renderer::Config rconfig {
    std::string{ss.str()},
    _config.getIntValue(Config::KEY_WINDOW_WIDTH),
    _config.getIntValue(Config::KEY_WINDOW_HEIGHT),
    _config.getIntValue(Config::KEY_OPENGL_MAJOR),
    _config.getIntValue(Config::KEY_OPENGL_MINOR),
    _config.getBoolValue(Config::KEY_FULLSCREEN)
  };

  renderer = std::make_unique<Renderer>(rconfig);

  mixer = std::make_unique<Mixer>();

  input = std::make_unique<Input>();
  assets = std::make_unique<Assets>();

  Assets::Manifest_t manifest {{engineFontKey, engineFontName, engineFontScale}};
  assets->loadFonts(manifest);

  Vector2i windowSize = pxr::renderer->getWindowSize();

  _app->initialize(this, windowSize._x, windowSize._y);

  _frameNo = 0;
  _isSleeping = true;
  _isDrawingPerformanceStats = false;
  _isDone = false;
}

void Engine::run()
{
  _realClock.start();
  while(!_isDone) mainloop();
}

void Engine::mainloop()
{
  auto now0 = Clock_t::now();
  auto realDt = _realClock.update();
  auto gameDt = _gameClock.update(realDt);
  auto gameNow = _gameClock.getNow();
  auto realNow = _realClock.getNow();

  SDL_Event event;
  while(SDL_PollEvent(&event) != 0){
    switch(event.type){
      case SDL_QUIT:
        _isDone = true;
        return;
      case SDL_WINDOWEVENT:
        if(event.window.event == SDL_WINDOWEVENT_SIZE_CHANGED)
          _app->onWindowResize(event.window.data1, event.window.data2);
        break;
      case SDL_KEYDOWN:
        if(event.key.keysym.sym == SDLK_LEFTBRACKET){
          _gameClock.incrementScale(-0.1);
          break;
        }
        else if(event.key.keysym.sym == SDLK_RIGHTBRACKET){
          _gameClock.incrementScale(0.1);
          break;
        }
        else if(event.key.keysym.sym == SDLK_KP_HASH){
          _gameClock.setScale(1.f);
          break;
        }
        else if(event.key.keysym.sym == SDLK_p && !_app->isWindowTooSmall()){
          _gameClock.togglePause();
          break;
        }
        else if(event.key.keysym.sym == SDLK_BACKQUOTE){
          _isDrawingPerformanceStats = !_isDrawingPerformanceStats;
          break;
        }
        // FALLTHROUGH
      case SDL_KEYUP:
        pxr::input->onKeyEvent(event);
        break;
    }
  }

  //
  // TODO: make the update loop more elegant - needed to make drawing based on real clock not
  // on game clock so drawing doesnt stop when the game pauses or slow down when the timeline 
  // is scaled.
  //

  for(int32_t i = LOOPTICK_UPDATE; i < LOOPTICK_COUNT; ++i){
    LoopTick& tick = _loopTicks[i];

    // ugly here!!!!!
    auto now = (i == LOOPTICK_UPDATE) ? gameNow : realNow;
    tick._ticksAccumulated += tick._metronome.doTicks(now);

    tick._ticksDoneThisFrame = 0;
    while(tick._ticksAccumulated > 0 && tick._ticksDoneThisFrame < tick._maxTicksPerFrame){
      ++tick._ticksDoneThisFrame;
      --tick._ticksAccumulated;
      (this->*tick._onTick)(gameNow, gameDt, realDt, tick._tickPeriod);
    }
    tick._tpsMeter.recordTicks(realDt, tick._ticksDoneThisFrame);
  }
  
  if(_isSleeping){
    auto now1 {Clock_t::now()};
    auto framePeriod {now1 - now0};
    if(framePeriod < minFramePeriod)
      std::this_thread::sleep_for(minFramePeriod - framePeriod); 
  }

  ++_frameNo;
  _fpsMeter.recordTicks(realDt, 1);
}

void Engine::drawPerformanceStats(Duration_t realDt, Duration_t gameDt)
{
  Vector2i windowSize = renderer->getWindowSize();
  renderer->setViewport({0, 0, std::min(300, windowSize._x), std::min(70, windowSize._y)});
  renderer->clearViewport(colors::blue);

  const Font& engineFont = assets->getFont(engineFontKey, engineFontScale);

  std::stringstream ss {};

  LoopTick* tick = &_loopTicks[LOOPTICK_UPDATE];
  ss << std::setprecision(3);
  ss << "UTPS:"  << tick->_tpsMeter.getTPS() << "hz"
     << "  UTA:" << tick->_ticksAccumulated
     << "  UTD:" << tick->_ticksDoneThisFrame
     << "  UTT:"  << tick->_metronome.getTotalTicks();
  renderer->blitText({5.f, 50.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);

  tick = &_loopTicks[LOOPTICK_DRAW];
  ss << std::setprecision(3);
  ss << "DTPS:"  << tick->_tpsMeter.getTPS() << "hz"
     << "  DTA:" << tick->_ticksAccumulated
     << "  DTD:" << tick->_ticksDoneThisFrame
     << "  DTT:" << tick->_metronome.getTotalTicks();
  renderer->blitText({5.f, 40.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);
  
  ss << std::setprecision(3);
  ss << "FPS:"  << _fpsMeter.getTPS() << "hz"
     << "  FNo:" << _frameNo;
  renderer->blitText({5.f, 30.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);
  
  ss << std::setprecision(3);
  ss << "Gdt:"   << durationToMilliseconds(gameDt) << "ms";
  renderer->blitText({5.f, 20.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);

  ss << std::setprecision(3);
  ss << "  Rdt:" << durationToMilliseconds(realDt) << "ms";
  renderer->blitText({120.f, 20.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);

  ss << std::setprecision(3);
  ss << "GNow:"     << durationToMinutes(_gameClock.getNow()) << "min";
  renderer->blitText({5.f, 10.f}, ss.str(), engineFont, colors::white); 

  std::stringstream().swap(ss);

  ss << std::setprecision(3);
  ss << "  Uptime:" << durationToMinutes(_realClock.getNow()) << "min";
  renderer->blitText({120.f, 10.f}, ss.str(), engineFont, colors::white); 
}

void Engine::drawPauseDialog()
{
  Vector2i windowSize = renderer->getWindowSize();
  renderer->setViewport({0, 0, windowSize._x, windowSize._y});

  const Font& engineFont = assets->getFont(engineFontKey, engineFontScale);

  Vector2f position {};
  position._x = (windowSize._x / 2.f) - 20.f; 
  position._y = (windowSize._y / 2.f) - 5.f;

  renderer->blitText(position, "PAUSED", engineFont, colors::white); 
}

void Engine::onUpdateTick(Duration_t gameNow, Duration_t gameDt, Duration_t realDt, float tickDt)
{
  double now = durationToSeconds(gameNow);
  _app->onUpdate(now, tickDt);
  pxr::input->onUpdate();
}

void Engine::onDrawTick(Duration_t gameNow, Duration_t gameDt, Duration_t realDt, float tickDt)
{
  pxr::renderer->clearWindow(colors::black);

  double now = durationToSeconds(gameNow);

  _app->onDraw(now, tickDt);

  if(_gameClock.isPaused())
    drawPauseDialog();

  if(_isDrawingPerformanceStats)
    drawPerformanceStats(realDt, gameDt);

  pxr::renderer->show();
}

double Engine::durationToMilliseconds(Duration_t d)
{
  return static_cast<double>(d.count()) / static_cast<double>(oneMillisecond.count());
}

double Engine::durationToSeconds(Duration_t d)
{
  return static_cast<double>(d.count()) / static_cast<double>(oneSecond.count());
}

double Engine::durationToMinutes(Duration_t d)
{
  return static_cast<double>(d.count()) / static_cast<double>(oneMinute.count());
}

} // namespace pxr

