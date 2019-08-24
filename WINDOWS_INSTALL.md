### Build Environment install

1. cygwin install from https://www.cygwin.com/

2. install needed package
 * mingw64-x86_64-gcc-core
 * mingw64-x86_64-gcc-g++
 * cmake
 * make
 * git

3. add [cygwin path]/bin to PATH

4. Cygwin terminal
```
cd /bin
ln -f x86_64-w64-mingw32-gcc.exe gcc.exe
ln -f x86_64-w64-mingw32-g++.exe g++.exe
ln -f x86_64-w64-mingw32-ar.exe ar.exe
```

### libsound.io install with Cygwin terminal

```
git clone https://github.com/andrewrk/libsoundio.git
cd libsoundio
cat << EOF > mingw64.cmake
SET(CMAKE_SYSTEM_NAME Windows)

SET(CMAKE_C_COMPILER x86_64-w64-mingw32-gcc)
SET(CMAKE_CXX_COMPILER x86_64-w64-mingw32-g++)
SET(CMAKE_RC_COMPILER x86_64-w64-mingw32-windres)

set(CMAKE_C_FLAGS "-Wall -Wextra")
set(CMAKE_CXX_FLAGS "-Wall -Wextra")
set(CMAKE_C_FLAGS_DEBUG "-g3 -gdwarf-2")
set(CMAKE_CXX_FLAGS_DEBUG "-g3 -gdwarf-2")

SET(CMAKE_FIND_ROOT_PATH /usr/x86_64-w64-mingw32)

set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
EOF
mkdir build
cd build
cmake -DCMAKE_TOOLCHAIN_FILE=../mingw64.cmake -DCMAKE_INSTALL_PREFIX=/usr/x86_64-w64-mingw32/sys-root/mingw -DCMAKE_BUILD_TYPE=Release ..
make
make install
```

### Go build

```
go build examples/sio_sine/main.go
```