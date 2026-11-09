@echo off
setlocal
cd /d "%~dp0"
rem Requires the 32-bit MinGW-w64 GCC toolchain on PATH.
rem For an unprefixed Windows toolchain: set CXX=g++ and set WINDRES=windres.
if not defined CXX set "CXX=i686-w64-mingw32-g++"
if not defined WINDRES set "WINDRES=i686-w64-mingw32-windres"
"%CXX%" -dumpmachine | findstr /b /c:"i686-w64-mingw32" >nul
if errorlevel 1 exit /b 1
"%WINDRES%" -O coff app.rc app-resource.o
if errorlevel 1 exit /b 1
set "FLAGS=-std=c++17 -Os -Wall -Wextra -Werror -fno-exceptions -fno-rtti -fno-stack-protector -ffreestanding -fno-builtin -nostdlib -static -mwindows -Wl,--entry,_WinMainCRTStartup -Wl,--subsystem,windows:5.1 -Wl,--major-os-version,5,--minor-os-version,1 -Wl,--no-insert-timestamp -s"
set "LIBS=-lcomctl32 -luxtheme -lgdi32 -luser32 -lkernel32"
"%CXX%" %FLAGS% gallery.cpp platform.cpp app-resource.o %LIBS% -o XPControls.exe
if errorlevel 1 exit /b 1
del app-resource.o
