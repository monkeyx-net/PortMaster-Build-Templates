# FindSDL2_image.cmake
# Finds SDL2_image via cmake config files (SDL2_image >= 2.6) or pkg-config fallback.
# Used on Ubuntu 22.04 where apt's SDL2_image 2.0.5 has no cmake config files.

find_package(SDL2_image ${SDL2_image_FIND_VERSION} CONFIG QUIET)
if(SDL2_image_FOUND)
    return()
endif()

find_package(PkgConfig REQUIRED)
pkg_check_modules(PC_SDL2_image QUIET SDL2_image)

find_path(SDL2_image_INCLUDE_DIR
    NAMES SDL_image.h
    HINTS ${PC_SDL2_image_INCLUDE_DIRS}
    PATH_SUFFIXES SDL2
)
find_library(SDL2_image_LIBRARY
    NAMES SDL2_image
    HINTS ${PC_SDL2_image_LIBRARY_DIRS}
)

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(SDL2_image
    REQUIRED_VARS SDL2_image_LIBRARY SDL2_image_INCLUDE_DIR
)

if(SDL2_image_FOUND)
    set(SDL2_IMAGE_INCLUDE_DIRS "${SDL2_image_INCLUDE_DIR}")
    set(SDL2_IMAGE_LIBRARIES    "${SDL2_image_LIBRARY}")
    if(NOT TARGET SDL2_image::SDL2_image)
        add_library(SDL2_image::SDL2_image UNKNOWN IMPORTED)
        set_target_properties(SDL2_image::SDL2_image PROPERTIES
            IMPORTED_LOCATION             "${SDL2_image_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${SDL2_image_INCLUDE_DIR}"
        )
    endif()
endif()
