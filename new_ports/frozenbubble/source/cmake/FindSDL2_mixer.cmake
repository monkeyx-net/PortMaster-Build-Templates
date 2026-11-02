# FindSDL2_mixer.cmake
# Finds SDL2_mixer via cmake config files (SDL2_mixer >= 2.6) or pkg-config fallback.

find_package(SDL2_mixer ${SDL2_mixer_FIND_VERSION} CONFIG QUIET)
if(SDL2_mixer_FOUND)
    return()
endif()

find_package(PkgConfig REQUIRED)
pkg_check_modules(PC_SDL2_mixer QUIET SDL2_mixer)

find_path(SDL2_mixer_INCLUDE_DIR
    NAMES SDL_mixer.h
    HINTS ${PC_SDL2_mixer_INCLUDE_DIRS}
    PATH_SUFFIXES SDL2
)
find_library(SDL2_mixer_LIBRARY
    NAMES SDL2_mixer
    HINTS ${PC_SDL2_mixer_LIBRARY_DIRS}
)

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(SDL2_mixer
    REQUIRED_VARS SDL2_mixer_LIBRARY SDL2_mixer_INCLUDE_DIR
)

if(SDL2_mixer_FOUND)
    set(SDL2_MIXER_INCLUDE_DIRS "${SDL2_mixer_INCLUDE_DIR}")
    set(SDL2_MIXER_LIBRARIES    "${SDL2_mixer_LIBRARY}")
    if(NOT TARGET SDL2_mixer::SDL2_mixer)
        add_library(SDL2_mixer::SDL2_mixer UNKNOWN IMPORTED)
        set_target_properties(SDL2_mixer::SDL2_mixer PROPERTIES
            IMPORTED_LOCATION             "${SDL2_mixer_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${SDL2_mixer_INCLUDE_DIR}"
        )
    endif()
endif()
