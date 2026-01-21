#!/bin/bash

# Pre build for SoH



# Do this process as part of build to allow disbaling of option
git clone https://github.com/leethomason/tinyxml2.git
cd tinyxml2
git checkout .
mkdir build-soh && cd build-soh
cmake -DBUILD_SHARED_LIBS=ON ..
make -j$(nproc)
make install

# prevent this file being found by cmake when SoH is compiled
cd ..
mv cmake/tinyxml2-config.cmake cmake/tinyxml2-config.cmake.disabled
cd ..
```
