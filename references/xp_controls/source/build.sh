#!/bin/sh
set -eu
cd "$(dirname "$0")"
compiler=${CXX:-i686-w64-mingw32-g++}
resource_compiler=${WINDRES:-i686-w64-mingw32-windres}
case "$("$compiler" -dumpmachine)" in
    i686-w64-mingw32) ;;
    *) echo 'Use an i686-w64-mingw32 (32-bit Windows) compiler.' >&2; exit 1 ;;
esac
"$resource_compiler" -O coff app.rc app-resource.o
"$compiler" -std=c++17 -Os -Wall -Wextra -Werror -fno-exceptions -fno-rtti -fno-stack-protector -ffreestanding -fno-builtin -nostdlib -static -mwindows -Wl,--entry,_WinMainCRTStartup -Wl,--subsystem,windows:5.1 -Wl,--major-os-version,5,--minor-os-version,1 -Wl,--no-insert-timestamp -s "gallery.cpp" platform.cpp app-resource.o -lcomctl32 -luxtheme -lgdi32 -luser32 -lkernel32 -o "XPControls.exe"
rm app-resource.o
