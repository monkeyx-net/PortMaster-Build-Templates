# Toolchain-mingw-w64.cmake - MinGW-w64 cross-compilation toolchain

set(CMAKE_SYSTEM_NAME Windows)
set(CMAKE_SYSTEM_PROCESSOR x86_64)

# Specify the cross compiler
set(CMAKE_C_COMPILER x86_64-w64-mingw32-gcc)
set(CMAKE_CXX_COMPILER x86_64-w64-mingw32-g++)

# Where to look for target environment (MinGW + your local libs)
set(CMAKE_FIND_ROOT_PATH
    /usr/x86_64-w64-mingw32
    ${CMAKE_SOURCE_DIR}/pc/lib/SDL2
)

# Programs should be found on host
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)

# Allow searching BOTH target + local project paths (important!)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY BOTH)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE BOTH)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE BOTH)

# Optional but helpful: explicitly hint SDL2 paths
set(SDL2_INCLUDE_DIR "${CMAKE_SOURCE_DIR}/pc/lib/SDL2/include")
set(SDL2_LIBRARY "${CMAKE_SOURCE_DIR}/pc/lib/SDL2/lib/libSDL2.a")

# (Optional) If SDL2main is required by your project:
set(SDL2MAIN_LIBRARY "${CMAKE_SOURCE_DIR}/pc/lib/SDL2/lib/libSDL2main.a")