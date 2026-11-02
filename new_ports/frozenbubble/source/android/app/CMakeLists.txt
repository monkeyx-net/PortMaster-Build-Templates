cmake_minimum_required(VERSION 3.22)
project(frozen-bubble CXX C)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_EXTENSIONS OFF)

# SDL2 and extension libs built from source in jni/
set(JNI_DIR "${CMAKE_SOURCE_DIR}/jni")

# Disable SDL2 test programs to speed up build
set(SDL_TEST OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_SAMPLES OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_SAMPLES OFF CACHE BOOL "" FORCE)

# SDL2_mixer: disable formats that require vendored external deps
# Use system packages where possible; on Android we don't need OGG/FLAC/MP3
set(SDL2MIXER_VENDORED ON CACHE BOOL "" FORCE)
set(SDL2MIXER_OGG ON CACHE BOOL "" FORCE)
set(SDL2MIXER_OGG_VENDORED ON CACHE BOOL "" FORCE)
set(SDL2MIXER_FLAC OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MP3 OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_OPUS OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MIDI OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_WAVPACK OFF CACHE BOOL "" FORCE)
# Keep only WAV support; disable all MOD formats (no system libs on Android)
set(SDL2MIXER_WAV ON CACHE BOOL "" FORCE)
set(SDL2MIXER_MOD OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MOD_MIKMOD OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MOD_MODPLUG OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MOD_XMP OFF CACHE BOOL "" FORCE)
set(SDL2MIXER_MOD_XMP_LITE OFF CACHE BOOL "" FORCE)

# SDL2_image: PNG + GIF needed (bubble textures are GIFs)
set(SDL2IMAGE_AVIF OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_BMP OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_GIF ON CACHE BOOL "" FORCE)
set(SDL2IMAGE_JPG OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_JXL OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_LBM OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_PCX OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_PNM OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_QOI OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_SVG OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_TGA OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_TIF OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_WEBP OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_XCF OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_XPM OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_XV OFF CACHE BOOL "" FORCE)
set(SDL2IMAGE_PNG ON CACHE BOOL "" FORCE)
set(SDL2IMAGE_VENDORED ON CACHE BOOL "" FORCE)

add_subdirectory("${JNI_DIR}/SDL2"       SDL2_build)
add_subdirectory("${JNI_DIR}/SDL2_image" SDL2_image_build)
add_subdirectory("${JNI_DIR}/SDL2_mixer" SDL2_mixer_build)
add_subdirectory("${JNI_DIR}/SDL2_ttf"   SDL2_ttf_build)

# Generate SDL2/ prefix wrapper headers so <SDL2/SDL_mixer.h> etc. work.
# Extension lib sources expose SDL_mixer.h (no prefix), but game source uses <SDL2/...>.
set(SDL2_EXT_WRAPPERS "${CMAKE_BINARY_DIR}/sdl2_ext_wrappers")
file(MAKE_DIRECTORY "${SDL2_EXT_WRAPPERS}/SDL2")
file(WRITE "${SDL2_EXT_WRAPPERS}/SDL2/SDL_image.h" "#pragma once\n#include <SDL_image.h>\n")
file(WRITE "${SDL2_EXT_WRAPPERS}/SDL2/SDL_mixer.h" "#pragma once\n#include <SDL_mixer.h>\n")
file(WRITE "${SDL2_EXT_WRAPPERS}/SDL2/SDL_ttf.h"   "#pragma once\n#include <SDL_ttf.h>\n")

# iniparser static lib (vendored in jni/iniparser)
add_library(iniparser STATIC
    "${JNI_DIR}/iniparser/iniparser.c"
    "${JNI_DIR}/iniparser/dictionary.c"
)
target_include_directories(iniparser PUBLIC "${JNI_DIR}/iniparser")

# Game as shared library (Android requires this)
add_library(main SHARED
    ../../src/frozenbubble.cpp
    ../../src/main.cpp
    ../../src/menubutton.cpp
    ../../src/mainmenu.cpp
    ../../src/gamesettings.cpp
    ../../src/audiomixer.cpp
    ../../src/shaderstuff.cpp
    ../../src/bubblegame.cpp
    ../../src/transitionmanager.cpp
    ../../src/ttftext.cpp
    ../../src/highscoremanager.cpp
    ../../src/networkclient.cpp
    ../../src/logger.cpp
    ../../src/platform.cpp
)

target_compile_definitions(main PRIVATE
    __ANDROID__
    DATA_DIR=""
)

target_include_directories(main PRIVATE
    ../../src
    "${SDL2_EXT_WRAPPERS}"     # Generated SDL2/ wrappers: <SDL2/SDL_mixer.h> → <SDL_mixer.h>
    "${JNI_DIR}/include"       # SDL2 core headers with SDL2/ subdir
)

target_link_libraries(main
    SDL2
    SDL2_image
    SDL2_mixer
    SDL2_ttf
    iniparser
    android
    log
    GLESv1_CM
    GLESv2
    EGL
    atomic
)
