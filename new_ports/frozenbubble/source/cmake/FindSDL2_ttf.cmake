# FindSDL2_ttf.cmake
# Finds SDL2_ttf via cmake config files (SDL2_ttf >= 2.20) or pkg-config fallback.

find_package(SDL2_ttf ${SDL2_ttf_FIND_VERSION} CONFIG QUIET)
if(SDL2_ttf_FOUND)
    return()
endif()

find_package(PkgConfig REQUIRED)
pkg_check_modules(PC_SDL2_ttf QUIET SDL2_ttf)

find_path(SDL2_ttf_INCLUDE_DIR
    NAMES SDL_ttf.h
    HINTS ${PC_SDL2_ttf_INCLUDE_DIRS}
    PATH_SUFFIXES SDL2
)
find_library(SDL2_ttf_LIBRARY
    NAMES SDL2_ttf
    HINTS ${PC_SDL2_ttf_LIBRARY_DIRS}
)

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(SDL2_ttf
    REQUIRED_VARS SDL2_ttf_LIBRARY SDL2_ttf_INCLUDE_DIR
)

if(SDL2_ttf_FOUND)
    set(SDL2_TTF_INCLUDE_DIRS "${SDL2_ttf_INCLUDE_DIR}")
    set(SDL2_TTF_LIBRARIES    "${SDL2_ttf_LIBRARY}")
    if(NOT TARGET SDL2_ttf::SDL2_ttf)
        add_library(SDL2_ttf::SDL2_ttf UNKNOWN IMPORTED)
        set_target_properties(SDL2_ttf::SDL2_ttf PROPERTIES
            IMPORTED_LOCATION             "${SDL2_ttf_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${SDL2_ttf_INCLUDE_DIR}"
        )
    endif()
endif()
