### Build Environment install

download and install `msys2-x86_64-yyyymmdd.exe` from  http://www.msys2.org/

```
pacman -Syuu
pacman -S base-devel
pacman -S mingw-w64-x86_64-toolchain
pacman -S mingw64/mingw-w64-x86_64-cmake
```

### libsound.io install

```
git clone https://github.com/andrewrk/libsoundio.git
cd libsoundio
mkdir build
cd build
cmake -G"MSYS Makefiles" -D CMAKE_INSTALL_PREFIX=/mingw64 ..
make
make install
```

### Go build

```
go build examples/sio_sine/main.go
```