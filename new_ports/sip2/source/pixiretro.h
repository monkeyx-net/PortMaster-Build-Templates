#ifndef _PIXIRETRO_H_
#define _PIXIRETRO_H_

#include <chrono>
#include <thread>
#include <cstdint>
#include <iostream>
#include <iomanip>
#include <algorithm>
#include <cassert>
#include <string>
#include <cstring>
#include <sstream>
#include <cmath>
#include <vector>
#include <initializer_list>
#include <memory>
#include <fstream>
#include <variant>
#include <tuple>
#include <random>
#include <limits>
#include <unordered_map>

#include <SDL2/SDL.h>
#include <SDL2/SDL_opengl.h>
#include <SDL2/SDL_mixer.h>

namespace pxr
{
//===============================================================================================//
// ##>UTILITY                                                                                    //
//===============================================================================================//

template<typename T>
inline T wrap(T value, T lo, T hi)
{
  return (value < lo) ? hi : (value > hi) ? lo : value;
}

//===============================================================================================//
// ##>MATH                                                                                       //
//===============================================================================================//

constexpr long double operator"" _hz(long double hz){return hz;}

struct Vector2f;

struct Vector2i
{
  constexpr Vector2i() : _x{0}, _y{0} {}
  constexpr Vector2i(int32_t x, int32_t y) : _x{x}, _y{y} {}
  constexpr Vector2i(const Vector2f& v);

  Vector2i(const Vector2i&) = default;
  Vector2i(Vector2i&&) = default;
  Vector2i& operator=(const Vector2i&) = default;
  Vector2i& operator=(Vector2i&&) = default;

  void zero() {_x = _y = 0;}
  bool isZero() {return _x == 0 && _y == 0;}
  Vector2i operator+(const Vector2i& v) const {return Vector2i{_x + v._x, _y + v._y};}
  void operator+=(const Vector2i& v) {_x += v._x; _y += v._y;}
  Vector2i operator-(const Vector2i& v) const {return Vector2i{_x - v._x, _y - v._y};}
  void operator-=(const Vector2i& v) {_x -= v._x; _y -= v._y;}
  Vector2i operator*(float scale) const {return Vector2i(_x * scale, _y * scale);}
  void operator*=(float scale) {_x *= scale; _y *= scale;}
  void operator*=(int32_t scale) {_x *= scale; _y *= scale;}
  float dot(const Vector2i& v) {return (_x * v._x) + (_y * v._y);}
  float cross(const Vector2i& v) const {return (_x * v._y) - (_y * v._x);}
  float length() const {return std::hypot(_x, _y);}
  float lengthSquared() const {return (_x * _x) + (_y * _y);}
  inline Vector2i normalized() const;
  inline void normalize();

  int32_t _x;
  int32_t _y;
};

struct Vector2f
{
  constexpr Vector2f() : _x{0.f}, _y{0.f} {}
  constexpr Vector2f(float x, float y) : _x{x}, _y{y} {}
  constexpr Vector2f(const Vector2i& v);

  Vector2f(const Vector2f&) = default;
  Vector2f(Vector2f&&) = default;
  Vector2f& operator=(const Vector2f&) = default;
  Vector2f& operator=(Vector2f&&) = default;

  void zero() {_x = _y = 0.f;}
  bool isZero() {return _x == 0.f && _y == 0.f;}
  Vector2f operator+(const Vector2f& v) const {return Vector2f{_x + v._x, _y + v._y};}
  void operator+=(const Vector2f& v) {_x += v._x; _y += v._y;}
  Vector2f operator-(const Vector2f& v) const {return Vector2f{_x - v._x, _y - v._y};}
  void operator-=(const Vector2f& v) {_x -= v._x; _y -= v._y;}
  Vector2f operator*(float scale) const {return Vector2f{_x * scale, _y * scale};}
  void operator*=(float scale) {_x *= scale; _y *= scale;}
  void operator*=(int32_t scale) {_x *= scale; _y *= scale;}
  float dot(const Vector2f& v) {return (_x * v._x) + (_y * v._y);}
  float cross(const Vector2f& v) const {return (_x * v._y) - (_y * v._x);}
  float length() const {return std::hypot(_x, _y);}
  float lengthSquared() const {return (_x * _x) + (_y * _y);}
  inline Vector2f normalized() const;
  inline void normalize();

  float _x;
  float _y;
};

constexpr Vector2i::Vector2i(const Vector2f& v) :
  _x{static_cast<int32_t>(v._x)},
  _y{static_cast<int32_t>(v._y)}
{}

Vector2i Vector2i::normalized() const
{
  Vector2i v = *this;
  v.normalize();
  return v;
}

void Vector2i::normalize()
{
  float l = (_x * _x) + (_y * _y);
  if(l) {
    l = std::sqrt(l);
    _x /= l;
    _y /= l;
  }
}

constexpr Vector2f::Vector2f(const Vector2i& v) :
  _x{static_cast<float>(v._x)},
  _y{static_cast<float>(v._y)}
{}

Vector2f Vector2f::normalized() const
{
  Vector2f v = *this;
  v.normalize();
  return v;
}

void Vector2f::normalize()
{
  float l = (_x * _x) + (_y * _y);
  if(l) {
    l = std::sqrt(l);
    _x /= l;
    _y /= l;
  }
}

template<typename T>
struct Rect
{
  T _x;
  T _y;
  T _w;
  T _h;
};

using iRect = Rect<int32_t>;
using fRect = Rect<float>;

//===============================================================================================//
// ##>RANDOM NUMBER GENERATION                                                                   //
//===============================================================================================//

class xorwow
{
public:
  using result_type = uint32_t;

  static constexpr int state_size = 6;
  using state_type = std::array<result_type, state_size>;

  constexpr static state_type default_seed {
    123456789, 975312468, 815652528, 175906542, 0, 0 
  };

  xorwow() : _state {default_seed}{};

  explicit xorwow(result_type value);
  explicit xorwow(const state_type& seed);
  explicit xorwow(std::seed_seq& seq);

  void seed();
  void seed(result_type seed);
  void seed(state_type seeds);

  void seed(std::seed_seq& seq);

  result_type operator()();

  void discard(unsigned long long z);

  static constexpr result_type min();
  static constexpr result_type max();

  const state_type& getState() const {return _state;}
  void setState(const state_type& state) {_state = state;}

  int32_t required_seed_size() const {return state_size - 2;}

private:
  state_type _state;
};

extern xorwow randGenerator;

//
// Two engines are equal if their internal states are equivilent.
//
bool operator==(const xorwow& lhs, const xorwow& rhs);
bool operator!=(const xorwow& lhs, const xorwow& rhs);

//
// returns a random signed integer uniformly distributed on the closed interval [li, hi].
//
int randUniformSignedInt(int lo, int hi);

//
// returns a random unsigned integer uniformly distributed on the closed interval [li, hi].
//
unsigned int randUniformUnsignedInt(unsigned int lo, unsigned int hi);

//
// returns a random real value uniformly distributed on the interval [li, hi).
//
double randUniformReal(double lo, double hi);

//===============================================================================================//
// ##>LOG                                                                                        //
//===============================================================================================//

namespace logstr
{
  constexpr const char* fail_open_log = "failed to open log";
  constexpr const char* fail_sdl_init = "failed to initialize SDL";
  constexpr const char* fail_create_opengl_context = "failed to create opengl context";
  constexpr const char* fail_set_opengl_attribute = "failed to set opengl attribute";
  constexpr const char* fail_create_window = "failed to create window";
  constexpr const char* fail_open_audio = "failed to open sdl mixer audio";

  constexpr const char* warn_cannot_open_dataset = "failed to open data file";
  constexpr const char* warn_cannot_create_dataset = "failed to create data file";
  constexpr const char* warn_malformed_dataset = "malformed data file";
  constexpr const char* warn_property_not_set = "property not set";
  constexpr const char* warn_malformed_bitmap = "malformed bitmap";
  constexpr const char* warn_bitmap_already_loaded = "bitmap already loaded";
  constexpr const char* warn_empty_bitmap_file = "empty bitmap file";
  constexpr const char* warn_font_already_loaded = "font already loaded";
  constexpr const char* warn_cannot_open_asset = "failed to open asset file";
  constexpr const char* warn_asset_parse_errors = "asset file parsing errors";
  constexpr const char* warn_cannot_play_sound = "failed to play a sound";
  constexpr const char* warn_cannot_load_sound = "failed to load a sound";
  constexpr const char* warn_missing_sound = "missing sound with key";

  constexpr const char* info_stderr_log = "logging to standard error";
  constexpr const char* info_using_default_config = "using default engine configuration";
  constexpr const char* info_fullscreen_mode = "activating fullscreen mode";
  constexpr const char* info_creating_window = "creating window";
  constexpr const char* info_created_window = "window created";
  constexpr const char* info_on_line = "on line";
  constexpr const char* info_on_row = "on row";
  constexpr const char* info_unknown_dataset_property = "unkown property";
  constexpr const char* info_unexpected_seperators = "expected a single seperator character";
  constexpr const char* info_incomplete_property = "missing property key or value";
  constexpr const char* info_expected_integer = "expected integer value";
  constexpr const char* info_expected_float = "expected float value";
  constexpr const char* info_expected_bool = "expected bool value";
  constexpr const char* info_ignoring_line = "ignoring line";
  constexpr const char* info_property_set = "dataset property set";
  constexpr const char* info_property_clamped = "property value clamped to min-max range";
  constexpr const char* info_using_property_defaults = "using property default values";
  constexpr const char* info_using_error_bitmap = "substituting with blank bitmap";
  constexpr const char* info_using_error_font = "substituting with blank font";
  constexpr const char* info_using_error_glyph = "substituting with blank glyph";
  constexpr const char* info_loading_asset = "loading asset";
  constexpr const char* info_skipping_asset_loading = "skipping asset loading";
  constexpr const char* info_ascii_code = "ascii code";
  constexpr const char* info_loaded_sound = "successfully loaded sound";
}; 

class Log
{
public:
  static constexpr const char* filename {"log"};
  static constexpr const char* delim {" : "};

  static constexpr std::array<const char*, 4> lvlstr {"fatal", "error", "warning", "info"};

public:
  enum Level
  {
    FATAL,
    ERROR,
    WARN,
    INFO
  };

public:
  Log();
  ~Log();

  void log(Level level, const char* error, const std::string& addendum = std::string{});

private:
  std::ofstream _os;
};

extern std::unique_ptr<Log> log;


//===============================================================================================//
// ##>INPUT                                                                                      //
//===============================================================================================//

class Input
{
public:
  enum KeyCode { 
    KEY_a, KEY_b, KEY_c, KEY_d, KEY_e, KEY_f, KEY_g, KEY_h, KEY_i, KEY_j, KEY_k, KEY_l, KEY_m, 
    KEY_n, KEY_o, KEY_p, KEY_q, KEY_r, KEY_s, KEY_t, KEY_u, KEY_v, KEY_w, KEY_x, KEY_y, KEY_z,
    KEY_SPACE, KEY_BACKSPACE, KEY_ENTER, KEY_LEFT, KEY_RIGHT, KEY_UP, KEY_DOWN, KEY_COUNT 
  };

  struct KeyLog
  {
    bool _isDown;
    bool _isPressed;
    bool _isReleased;
  };

public:
  Input();
  ~Input() = default;

  void onKeyEvent(const SDL_Event& event);
  void onUpdate();

  bool isKeyDown(KeyCode key) {return _keys[key]._isDown;}
  bool isKeyPressed(KeyCode key) {return _keys[key]._isPressed;}
  bool isKeyReleased(KeyCode key) {return _keys[key]._isReleased;}

  int32_t keyToAsciiCode(KeyCode key) const;

  const std::vector<KeyCode>& getHistory() const {return _history;}

private:
  KeyCode convertSdlKeyCode(int sdlCode);

private:
  std::array<KeyLog, KEY_COUNT> _keys;
  std::vector<KeyCode> _history;        // All keys pressed between calls to 'onUpdate'.
};

extern std::unique_ptr<Input> input;

//===============================================================================================//
// ##>RESOURCES                                                                                  //
//===============================================================================================//

class Bitmap;
class Font;
struct Glyph;

class Dataset
{
public:
  using Value_t = std::variant<int32_t, float, bool>;

  constexpr static Value_t unsetValue {std::numeric_limits<int32_t>::min()};

  enum Type {INT_PROPERTY, FLOAT_PROPERTY, BOOL_PROPERTY};

  struct Property
  {
    Property() = default;
    Property(int32_t key, std::string name, Value_t default_, Value_t min, Value_t max); 

    int32_t _key;
    std::string _name;
    Value_t _value;
    Value_t _default;
    Value_t _min;
    Value_t _max;
    Type _type;
  };

public:
  static constexpr char comment {'#'};
  static constexpr char seperator {'='};

public:
  int32_t load(const std::string& filename);
  int32_t write(const std::string& filename, bool genComments = true);

  int32_t getIntValue(int32_t key) const;
  float getFloatValue(int32_t key) const;
  bool getBoolValue(int32_t key) const;

  void setIntValue(int32_t key, int32_t value);
  void setFloatValue(int32_t key, float value);
  void setBoolValue(int32_t key, bool value);

  void scaleIntValue(int32_t key, int32_t scale);
  void scaleFloatValue(int32_t key, float scale);

  void applyDefaults();

protected:
  Dataset(std::initializer_list<Property> properties);

  Dataset() = delete;
  Dataset(const Dataset&) = delete;
  Dataset(Dataset&&) = delete;
  Dataset& operator=(const Dataset&) = delete;
  Dataset& operator=(Dataset&&) = delete;

private:
  bool parseInt(const std::string& value, int32_t& result);
  bool parseFloat(const std::string& value, float& result);
  bool parseBool(const std::string& value, bool& result);

  void printValue(const Value_t& value, std::ostream& os);
  
private:
  std::unordered_map<int32_t, Property> _properties;
};

class Assets
{
public:
  using Key_t = int32_t;
  using Name_t = const char*;
  using Scale_t = size_t;
  using Manifest_t = std::vector<std::tuple<Key_t, Name_t, Scale_t>>;
  
public:
  static constexpr const int32_t maxScale {16};

public:
  Assets() = default;
  ~Assets() = default;

  Assets(const Assets&) = delete;
  Assets(Assets&&) = delete;

  Assets& operator=(const Assets&) = delete;
  Assets& operator=(Assets&&) = delete;

  void loadBitmaps(const Manifest_t& manifest);
  void loadFonts(const Manifest_t& manifest);

  const Bitmap& getBitmap(Key_t key, Scale_t scale) const;
  const Font& getFont(Key_t key, Scale_t scale) const;

  Bitmap makeBlockBitmap(int32_t width, int32_t height);

private:
  static constexpr const char path_seperator = '/';

  static constexpr const char* bitmaps_path = "assets/bitmaps";
  static constexpr const char* fonts_path = "assets/fonts";
  static constexpr const char* bitmaps_extension = ".bitmap";
  static constexpr const char* fonts_extension = ".font";
  static constexpr const char* glyphs_extension = ".glyph";

  static constexpr std::array<const char*, 8> errorBits {
    "11111111", "11111111", "11111111", "11111111", "11111111", "11111111", "11111111", "11111111"
  };

  static constexpr int32_t asciiCharCount {94};

  static constexpr const std::array<const char*, asciiCharCount> glyphFilenames {
    "emark", "dquote", "hash", "dollar", "percent", "ampersand", "squote", "lrbracket", "rrbracket",
    "asterix", "plus", "comma", "minus", "dot", "fslash", "0", "1", "2", "3", "4", "5", "6", "7", 
    "8", "9", "colon", "scolon", "lcroc", "equals", "rcroc", "qmark", "at", "a", "b", "c", "d", "e",
    "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x",
    "y", "z", "lsbracket", "bslash", "rsbracket", "carrot", "underscore", "backtick", "a", "b", "c",
    "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", 
    "w", "x", "y", "z", "lcbracket", "pipe", "rcbracket", "tilde"
  };

  class FontData : public Dataset
  {
  public:
    enum Key { KEY_LINE_SPACE, KEY_WORD_SPACE, KEY_GLYPH_SPACE, KEY_SIZE };

    FontData() : Dataset({
      // key            name          default   min    max
      {KEY_LINE_SPACE , "lineSpace" , {10}    , {0}  , {1000}},
      {KEY_WORD_SPACE , "wordSpace" , {5}     , {0}  , {1000}},
      {KEY_GLYPH_SPACE, "glyphSpace", {2}     , {0}  , {1000}},
      {KEY_SIZE       , "size"      , {8}     , {0}  , {1000}}
    }){}
  };

  class GlyphData : public Dataset
  {
  public:
    enum Key { KEY_ASCII_CODE, KEY_OFFSET_X, KEY_OFFSET_Y, KEY_ADVANCE, KEY_WIDTH, KEY_HEIGHT };

    GlyphData() : Dataset({
      // key           name         default  min       max
      {KEY_ASCII_CODE, "asciiCode", {0}    , {0}     , {1000}},
      {KEY_OFFSET_X  , "offsetX"  , {0}    , {-1000} , {1000}},
      {KEY_OFFSET_Y  , "offsetY"  , {0}    , {-1000} , {1000}},
      {KEY_ADVANCE   , "advance"  , {8}    , {0}     , {1000}},
      {KEY_WIDTH     , "width"    , {8}    , {0}     , {1000}},
      {KEY_HEIGHT    , "height"   , {8}    , {0}     , {1000}}
    }){}
  };

private:
  Bitmap loadBitmap(std::string path, std::string name, Scale_t scale = 1);
  Font loadFont(std::string name, Scale_t scale = 1);
  Bitmap generateErrorBitmap(Scale_t scale);
  Font generateErrorFont(Scale_t scale);
  Glyph generateErrorGlyph(int32_t asciiCode, Scale_t scale);

private:
  std::unordered_map<Key_t, std::array<std::unique_ptr<Bitmap>, maxScale>> _bitmaps;
  std::unordered_map<Key_t, std::array<std::unique_ptr<Font>, maxScale>> _fonts;
};

extern std::unique_ptr<Assets> assets;

//===============================================================================================//
// ##>GRAPHICS                                                                                   //
//===============================================================================================//

// Bits in the bitmap are accessible via a [row][col] position mapped to screen space like:
//
//          row                               
//           ^
//           |                              y
//         7 | | | |█|█| | | |              ^
//         6 | | |█|█|█|█| | |              |      screen-space
//         5 | |█|█|█|█|█|█| |              |         axes
//         4 |█|█| |█|█| |█|█|       ==>    |
//         3 |█|█|█|█|█|█|█|█|              |
//         2 | | |█| | |█| | |              +----------> x
//         1 | |█| |█|█| |█| |
//         0 |█| |█| | |█| |█|           i.e bit[0][0] is the bottom-left most bit.
//           +-----------------> col
//            0 1 2 3 4 5 6 7
//
// Thus the bitmap can be considered to be a coordinate space in which the origin is the bottom-
// left most pixel of the bitmap.

class Bitmap final
{
  friend Assets;

public:
  Bitmap(const Bitmap& other);
  Bitmap(Bitmap&& other) noexcept;
  Bitmap& operator=(const Bitmap& other);
  Bitmap& operator=(Bitmap&& other) noexcept;
  ~Bitmap() = default;

  bool getBit(int32_t row, int32_t col) const;
  int32_t getWidth() const {return _width;}
  int32_t getHeight() const {return _height;}
  const std::vector<uint8_t>& getBytes() const {return _bytes;}
  void uploadTexture() const;
  GLuint getTexId() const {return _texId;}

  void setBit(int32_t row, int32_t col, bool value, bool regen = true);
  void setRect(int32_t rowMin, int32_t colMin, int32_t rowMax, int32_t colMax, bool value, bool regen = true);
  
  void regenerateBytes();

  bool isEmpty();
  bool isApproxEmpty(int32_t threshold);

  void print(std::ostream& out) const;

private:
  Bitmap() = default;

  void initialize(std::vector<std::string> bits, int32_t scale = 1);

private:
  std::vector<std::vector<bool>> _bits;  // used for bit manipulation ops - indexed [row][col]
  std::vector<uint8_t> _bytes;           // packed bits for legacy use
  int32_t _width;
  int32_t _height;
  mutable GLuint _texId {0};
  mutable bool _texDirty {true};
};

struct Glyph // note -- cannot nest in font as it needs to be forward declared.
{
  Bitmap _bitmap;
  int32_t _asciiCode;
  int32_t _offsetX;
  int32_t _offsetY;
  int32_t _advance;
  int32_t _width;
  int32_t _height;
};

class Font
{
  friend class Assets;

public:
  struct Meta
  {
    int32_t _lineSpace;
    int32_t _wordSpace;
    int32_t _glyphSpace;
    int32_t _size;
  };

public:
  Font(const Font&) = default;
  Font(Font&&) = default;
  Font& operator=(const Font&) = default;
  Font& operator=(Font&&) = default;

  const Glyph& getGlyph(char c) const {return _glyphs[static_cast<int32_t>(c - '!')];}
  int32_t getLineSpace() const {return _meta._lineSpace;}
  int32_t getWordSpace() const {return _meta._wordSpace;}
  int32_t getGlyphSpace() const {return _meta._glyphSpace;}
  int32_t getSize() const {return _meta._size;}

  int32_t calculateStringWidth(const std::string& str) const;

private:
  Font() = default;
  void initialize(Meta meta, std::vector<Glyph> glyphs);


private:
  std::vector<Glyph> _glyphs;
  Meta _meta;
};

class Color3f
{
  constexpr static float lo {0.f};
  constexpr static float hi {1.f};

public:
  Color3f() : _r{0.f}, _g{0.f}, _b{0.f}{}

  constexpr Color3f(float r, float g, float b) : 
    _r{std::clamp(r, lo, hi)},
    _g{std::clamp(g, lo, hi)},
    _b{std::clamp(b, lo, hi)}
  {}

  Color3f(const Color3f&) = default;
  Color3f(Color3f&&) = default;
  Color3f& operator=(const Color3f&) = default;
  Color3f& operator=(Color3f&&) = default;

  float getRed() const {return _r;}
  float getGreen() const {return _g;}
  float getBlue() const {return _b;}
  void setRed(float r){_r = std::clamp(r, lo, hi);}
  void setGreen(float g){_g = std::clamp(g, lo, hi);}
  void setBlue(float b){_b = std::clamp(b, lo, hi);}

private:
  float _r;
  float _g;
  float _b;
};

namespace colors
{
constexpr Color3f white {1.f, 1.f, 1.f};
constexpr Color3f black {0.f, 0.f, 0.f};
constexpr Color3f red {1.f, 0.f, 0.f};
constexpr Color3f green {0.f, 1.f, 0.f};
constexpr Color3f blue {0.f, 0.f, 1.f};
constexpr Color3f cyan {0.f, 1.f, 1.f};
constexpr Color3f magenta {1.f, 0.f, 1.f};
constexpr Color3f yellow {1.f, 1.f, 0.f};

// greys - more grays: https://en.wikipedia.org/wiki/Shades_of_gray 

constexpr Color3f gainsboro {0.88f, 0.88f, 0.88f};
constexpr Color3f lightgray {0.844f, 0.844f, 0.844f};
constexpr Color3f silver {0.768f, 0.768f, 0.768f};
constexpr Color3f mediumgray {0.76f, 0.76f, 0.76f};
constexpr Color3f spanishgray {0.608f, 0.608f, 0.608f};
constexpr Color3f gray {0.512f, 0.512f, 0.512f};
constexpr Color3f dimgray {0.42f, 0.42f, 0.42f};
constexpr Color3f davysgray {0.34f, 0.34f, 0.34f};
constexpr Color3f jet {0.208f, 0.208f, 0.208f};
};

class Renderer
{
public:
  struct Config
  {
    std::string _windowTitle;
    int32_t _windowWidth;
    int32_t _windowHeight;
    int32_t _openglVersionMajor;
    int32_t _openglVersionMinor;
    bool _fullscreen;
  };
  
public:
  Renderer(const Config& config);
  Renderer(const Renderer&) = delete;
  Renderer* operator=(const Renderer&) = delete;
  ~Renderer();
  void setViewport(iRect viewport);
  void blitText(Vector2f position, const std::string& text, const Font& font, const Color3f& color);
  void blitBitmap(Vector2f position, const Bitmap& bitmap, const Color3f& color);
  void drawBorderRect(const iRect& rect, const Color3f& background, const Color3f& borderColor, int32_t borderWidth = 1);
  void clearWindow(const Color3f& color);
  void clearViewport(const Color3f& color);
  void show();
  Vector2i getWindowSize() const;

private:
  SDL_Window* _window;
  SDL_GLContext _glContext;
  Config _config;
  iRect _viewport;
};

extern std::unique_ptr<Renderer> renderer;

//===============================================================================================//
// ##>COLLISION DETECTION                                                                        //
//===============================================================================================//

// COLLISION RESULTS
//
// The following is an explanation of the data returned from a collision test between two bitmaps,
// say bitmap A and bitmap B.
//
// To test for pixel perfect collisions between two bitmaps, each bitmap is considered to be
// bounded by an axis-aligned bounding box (AABB) calculated from the positions of the bitmaps
// and their dimensions. Two overlap AABBs are calculated from any intersection, one for each
// bitmap. The overlaps represent the local overlap (i.e. coordinates w.r.t the bitmap coordinates,
// see class Bitmap) of each bitmap, illustrated in figure 1.
//
//               Wa=20                     KEY
//           +----------+                  ===
//           |          |                  Pn = position of bitmap N
//     Ha=20 |          |                  Wn = width of bitmap N
//           |     +----|-----+            Hn = height of bitmap N
//           | A   | S  |     |
// Pa(20,20) o-----|----+     | Hb=20      S = overlap region of bitmaps A and B.
//                 |          |
//                 | B        |            There is only a single overlap region S for any collision.
//       Pb(30,10) o----------+            However this region can be expressed w.r.t the coordinate
//                     Wb=20               space of each bitmap.
//    
//     y                                   Both expressions will be returned. In this example we
//     ^                                   will have the result:
//     |  screen                                          left, right, top, bottom
//     |   axes                               aOverlap = {10  , 20   , 10 , 0 }
//     o-----> x                              bOverlap = {0   , 10   , 20 , 10}
//
//                                        Note that the overlap S w.r.t the screen would be:
//                                             Overlap = {30  , 40   , 30 , 20}
//                            [ figure 1]
//
//  Further, lists of all pixel intersections can also be returned. Intersecting pixels are returned
//  as two lists: the set of all pixels in bitmap A which intersect a pixel in bitmap B (aPixels)
//  and the set of all pixels in bitmap B which intersect a pixel in bitmap A (bPixels). The
//  pixels in the lists are expressed in coordinates w.r.t their bitmaps coordinate space. Note 
//  pixels are returned as 2D vectors where [x,y] = [col][row]. 
//
//  USAGE NOTES
//
//  The collision data returned from the collision test function is stored internally and returned
//  via constant reference to avoid allocating memory for each and every test. Consequently the
//  data returned persists only inbetween calls to the test function and is overwritten by 
//  subsequent calls. If you need persistence then copy construct the returned Collision struct.
//
//  Not every usage requires all collision data thus some collision data is optional where skipping
//  the collection of such data can provide performance benefits, notably the pixel lists. If a 
//  pixel list is not required the test resolution can shortcut with a positive result upon 
//  detecting the first pixel intersection. By default lists are generated.

// AABB:
//              +-------x (xmax, ymax)       y
//              |       |                    ^  screen
//              | AABB  |                    |   axes
//              |       |                    |   
// (xmin, ymin) o-------+                    o------> x
struct AABB
{
  int32_t _xmin;
  int32_t _ymin;
  int32_t _xmax;
  int32_t _ymax;
};

struct Collision
{
  bool _isCollision;
  AABB _aBounds;
  AABB _bBounds;
  AABB _aOverlap;
  AABB _bOverlap;
  std::vector<Vector2i> _aPixels;
  std::vector<Vector2i> _bPixels;
};

const Collision& testCollision(Vector2i aPosition, const Bitmap& aBitmap, 
                               Vector2i bPosition, const Bitmap& bBitmap, bool pixelLists = true);

//===============================================================================================//
// ##>MIXER                                                                                      //
//===============================================================================================//

class Mixer
{
public:
  using Key_t = int32_t;
  using Name_t = const char*;
  using Channel_t = int;
  using Manifest_t = std::vector<std::pair<Key_t, Name_t>>;

  static constexpr int sampleFreq = MIX_DEFAULT_FREQUENCY;
  static constexpr int sampleFormat = MIX_DEFAULT_FORMAT;
  static constexpr int numOutChannels = 2;   // 1=mono, 2=stereo, other=error
  static constexpr int chunkSize = 1024;
  static constexpr int numMixChannels = 16;

  static constexpr const char* sounds_path = "assets/sounds/";
  static constexpr const char* sounds_extension = ".wav";

  static constexpr Channel_t allChannels = -1;
  static constexpr Channel_t nullChannel = -2;

  Mixer();  
  ~Mixer();

  void loadSoundsWAV(const Manifest_t& manifest);

  Channel_t playSound(Key_t sndkey, int loops = 0);
  Channel_t playSoundTimed(Key_t sndkey, int loops, int timeLimit_ms);
  Channel_t playSoundFadeIn(Key_t sndkey, int loops, int fadeInTime_ms);
  Channel_t playSoundFadeInTimed(Key_t sndkey, int loops, int fadeInTime_ms, int timeLimit_ms);

  void stopChannel(Channel_t channel);
  void pauseChannel(Channel_t channel);
  void resumeChannel(Channel_t channel);

  void setVolume(float volume); // value from 0 to 1.
  int getVolume() const {return _volume;}

private:
  Mix_Chunk* findChunk(Key_t sndkey);

private:
  std::unordered_map<Key_t, Mix_Chunk*> _sounds;
  float _volume;
};

extern std::unique_ptr<Mixer> mixer;

//===============================================================================================//
// ##>UI                                                                                         //
//===============================================================================================//

//
// TODO - There is a fair amount of diplicate code in this class. Could be improved.
//

class HUD
{
public:
  using uid_t = int32_t;

  struct TextLabel
  {
    TextLabel(Vector2i p, Color3f c, std::string t, float activeDelay = 0.f, bool phase = false, bool flash = false);

    TextLabel(const TextLabel&) = default;
    TextLabel(TextLabel&&) = default;
    TextLabel& operator=(const TextLabel&) = default;
    TextLabel& operator=(TextLabel&&) = default;

    uid_t _uid;
    Vector2i _position;
    Color3f _color;
    std::string _text;       // Source text.
    std::string _value;      // The text being shown.
    int32_t _charNo;         // The index into '_text' of the last character shown if '_phaseIn' = true.
    float _activeDelay;
    float _activeTime;
    bool _isActive;          // Allows delayed activation.
    bool _isVisible;         // Used internally to flash the label.
    bool _isHidden;          // Allows manual hiding.
    bool _phase;
    bool _flash;
  };

  struct IntLabel
  {
    IntLabel(Vector2i p, Color3f c, const int32_t* s, int32_t precision, float activeDelay = 0.f, bool flash = false);

    IntLabel(const IntLabel&) = default;
    IntLabel(IntLabel&&) = default;
    IntLabel& operator=(const IntLabel&) = default;
    IntLabel& operator=(IntLabel&&) = default;

    uid_t _uid;
    Vector2i _position;
    Color3f _color;
    const int32_t* _source;  // The integer whom's value to display.
    int32_t _value;          // The current value of the source.
    int32_t _precision;      // Number of digits to display.
    std::string _text;       // Text generated from the value.
    float _activeDelay;
    float _activeTime;
    bool _isActive;
    bool _isVisible;
    bool _isHidden;
    bool _flash;
  };

  struct BitmapLabel
  {
    BitmapLabel(Vector2i p, Color3f c, const Bitmap* b, float activeDelay = 0.f, bool flash = false);

    BitmapLabel(const BitmapLabel&) = default;
    BitmapLabel(BitmapLabel&&) = default;
    BitmapLabel& operator=(const BitmapLabel&) = default;
    BitmapLabel& operator=(BitmapLabel&&) = default;

    uid_t _uid;
    Vector2i _position;
    Color3f _color;
    const Bitmap* _bitmap;
    float _activeDelay;
    float _activeTime;
    bool _isActive;
    bool _isVisible;
    bool _isHidden;
    bool _flash;
  };

public:
  HUD() = default;
  ~HUD() = default;
  HUD(const HUD&) = default;
  HUD(HUD&&) = default;
  HUD& operator=(const HUD&) = default;
  HUD& operator=(HUD&&) = default;

  void initialize(const Font* font, float flashPeriod, float phaseInPeriod);

  uid_t addTextLabel(TextLabel label);
  uid_t addIntLabel(IntLabel label);
  uid_t addBitmapLabel(BitmapLabel label);

  bool removeTextLabel(uid_t uid);
  bool removeIntLabel(uid_t uid);
  bool removeBitmapLabel(uid_t uid);

  void clear();

  void hideTextLabel(uid_t uid);
  void hideIntLabel(uid_t uid);
  void hideBitmapLabel(uid_t uid);
  void showTextLabel(uid_t uid);
  void showIntLabel(uid_t uid);
  void showBitmapLabel(uid_t uid);

  void startTextLabelFlash(uid_t uid);
  void startIntLabelFlash(uid_t uid);
  void startBitmapLabelFlash(uid_t uid);
  void stopTextLabelFlash(uid_t uid);
  void stopIntLabelFlash(uid_t uid);
  void stopBitmapLabelFlash(uid_t uid);

  void onReset();
  void onUpdate(float dt);
  void onDraw();

  void setFont(const Font* font) {_font = font;}
  void setFlashPeriod(float period) {_flashPeriod = period;}
  void setPhasePeriod(float period) {_phasePeriod = period;}

private:
  void flashLabels();
  void phaseLabels();
  void updateIntLabels();
  void activateLabels();

private:
  const Font* _font;

  std::vector<TextLabel> _textLabels;
  std::vector<IntLabel> _intLabels;
  std::vector<BitmapLabel> _bitmapLabels;

  int32_t _nextUid;

  float _masterClock;     // Unit: seconds. Master timeline which all timing is relative to.
  float _flashPeriod;     // Unit: seconds. Inverse of the frequency of flashing.
  float _phasePeriod;     // Unit: seconds. Period between each subsequent letter appearing.
  float _flashClock;
  float _phaseClock;
};

// 
// A UI text input box.
//
class TextInput
{
public:
  TextInput(Font* font, std::string label, Color3f cursorColor);

  const char* processInput();
  void draw(float dt);

private:
  Font* _font;
  std::vector<char> _buffer;
  int32_t _bufferSize;
  Color3f cursorColor;
  int32_t _cursorPos;
  float _cursorFlashPeriod;
  float _cursorFlashClock;
  std::string label;
  Vector2i _boxPosition;
  int32_t _boxW;
  int32_t _boxH;
};

//===============================================================================================//
// ##>APPLICATION                                                                                //
//===============================================================================================//

class Application;

class ApplicationState
{
public:
  ApplicationState(Application* app) : _app(app) {}
  virtual ~ApplicationState() = default;
  virtual void initialize(Vector2i worldSize, int32_t worldScale) = 0;
  virtual void onUpdate(double now, float dt) = 0;
  virtual void onDraw(double now, float dt) = 0;
  virtual void onEnter() = 0;
  virtual std::string getName() = 0;

protected:
  Application* _app;
};

class Engine;

class Application
{
public:
  static constexpr Input::KeyCode pauseKey {Input::KEY_p};

public:
  Application() = default;
  virtual ~Application() = default;
  
  virtual bool initialize(Engine* engine, int32_t windowWidth, int32_t windowHeight);

  virtual std::string getName() const = 0;
  virtual int32_t getVersionMajor() const = 0;
  virtual int32_t getVersionMinor() const = 0;

  void onWindowResize(int32_t windowWidth, int32_t windowHeight);
  virtual void onUpdate(double now, float dt);
  virtual void onDraw(double now, float dt);

  void switchState(const std::string& name);

  bool isWindowTooSmall() const {return _isWindowTooSmall;}

protected:
  virtual Vector2i getWorldSize() const = 0;

  void addState(std::unique_ptr<ApplicationState>&& state);

private:
  void drawWindowTooSmall();

private:
  std::unordered_map<std::string, std::unique_ptr<ApplicationState>> _states;
  std::unique_ptr<ApplicationState>* _activeState;
  iRect _viewport;
  Engine* _engine;
  bool _isWindowTooSmall;
};

//===============================================================================================//
// ##>ENGINE                                                                                     //
//===============================================================================================//

class Engine
{
public:
  using Clock_t = std::chrono::steady_clock;
  using TimePoint_t = std::chrono::time_point<Clock_t>;
  using Duration_t = std::chrono::nanoseconds;

  constexpr static Duration_t oneMillisecond {1'000'000};
  constexpr static Duration_t oneSecond {1'000'000'000};
  constexpr static Duration_t oneMinute {60'000'000'000};
  constexpr static Duration_t minFramePeriod {1'000'000};

  constexpr static Assets::Key_t engineFontKey {0}; // engine reserves this font key for itself.
  constexpr static Assets::Name_t engineFontName {"engine"};
  constexpr static Assets::Scale_t engineFontScale {1};

  class RealClock
  {
  public:
    RealClock() : _start{}, _now0{}, _now1{}, _dt{}{}
    ~RealClock() = default;
    void start(){_now0 = _start = Clock_t::now();}
    Duration_t update();
    Duration_t getDt() const {return _dt;}
    Duration_t getNow() {return Clock_t::now() - _start;}

  private:
    TimePoint_t _start;
    TimePoint_t _now0;
    TimePoint_t _now1;
    Duration_t _dt;
  };

  class GameClock
  {
  public:
    GameClock() : _now{}, _scale{1.f}, _isPaused{false}{}
    ~GameClock() = default;
    Duration_t update(Duration_t realDt);
    Duration_t getNow() const {return _now;}
    Duration_t getDt() const {return _dt;}
    void incrementScale(float increment){_scale += increment;}
    void setScale(float scale){_scale = scale;}
    float getScale() const {return _scale;}
    void pause(){_isPaused = true;}
    void unpause(){_isPaused = false;}
    void togglePause(){_isPaused = !_isPaused;}
    bool isPaused() const {return _isPaused;}

  private:
    Duration_t _now;
    Duration_t _dt;
    float _scale;
    bool _isPaused;
  };

  class Metronome
  {
  public:
    Metronome() : _lastTickNow{}, _tickPeriod{}, _totalTicks{0}{}
    ~Metronome() = default;
    int64_t doTicks(Duration_t gameNow);
    void setTickPeriod(Duration_t period) {_tickPeriod += period;}
    Duration_t getTickPeriod() const {return _tickPeriod;}
    int64_t getTotalTicks() const {return _totalTicks;}

  private:
    Duration_t _lastTickNow;
    Duration_t _tickPeriod;
    int64_t _totalTicks;
  };

  class TPSMeter
  {
  public:
    TPSMeter() : _timer{}, _ticks{0}, _tps{0} {}
    ~TPSMeter() = default;
    void recordTicks(Duration_t realDt, int32_t ticks);
    int32_t getTPS() const {return _tps;}
    
  private:
    Duration_t _timer;
    int32_t _ticks;
    int32_t _tps;
  };

  enum LoopTicks {LOOPTICK_UPDATE, LOOPTICK_DRAW, LOOPTICK_COUNT};

  struct LoopTick
  {
    void (Engine::*_onTick)(Duration_t, Duration_t, Duration_t, float);
    TPSMeter _tpsMeter;
    Metronome _metronome;
    int64_t _ticksAccumulated;
    int32_t _ticksDoneThisFrame;
    int32_t _maxTicksPerFrame;
    float _tickPeriod;
  };
  
  class Config final : public Dataset
  {
  public:
    static constexpr const char* filename = "engine.config";
    
    enum Key
    {
      KEY_WINDOW_WIDTH, 
      KEY_WINDOW_HEIGHT, 
      KEY_FULLSCREEN, 
      KEY_OPENGL_MAJOR, 
      KEY_OPENGL_MINOR,
    };

    Config() : Dataset({
      //    key               name        default   min      max
      {KEY_WINDOW_WIDTH,  "windowWidth",  {500},   {300},   {1000}},
      {KEY_WINDOW_HEIGHT, "windowHeight", {500},   {300},   {1000}},
      {KEY_FULLSCREEN,    "fullscreen",   {false}, {false}, {true}},
      {KEY_OPENGL_MAJOR,  "openglMajor",  {2},     {2},     {2},  },
      {KEY_OPENGL_MINOR,  "openglMinor",  {1},     {1},     {1},  }
    }){}
  };

public:
  Engine() = default;
  ~Engine() = default;
  void initialize(std::unique_ptr<Application> app);
  void run();
  void pause(){_gameClock.pause();}
  void unpause(){_gameClock.unpause();}
  void togglePause(){_gameClock.togglePause();}

private:
  void mainloop();
  void drawPerformanceStats(Duration_t realDt, Duration_t gameDt);
  void drawPauseDialog();
  void onUpdateTick(Duration_t gameNow, Duration_t gameDt, Duration_t realDt, float tickDt);
  void onDrawTick(Duration_t gameNow, Duration_t gameDt, Duration_t realDt, float tickDt);

  double durationToMilliseconds(Duration_t d);
  double durationToSeconds(Duration_t d);
  double durationToMinutes(Duration_t d);

private:
  Config _config;
  std::array<LoopTick, LOOPTICK_COUNT> _loopTicks;
  TPSMeter _fpsMeter;
  int64_t _frameNo;
  RealClock _realClock;
  GameClock _gameClock;
  std::unique_ptr<Application> _app;
  bool _isSleeping;
  bool _isDrawingPerformanceStats;
  bool _isDone;
};

} // namespace pxr

#endif
