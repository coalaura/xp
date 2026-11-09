#pragma once
#ifndef UNICODE
#define UNICODE
#define _UNICODE
#endif
#define WINVER 0x0501
#define _WIN32_WINNT 0x0501
#define _WIN32_IE 0x0600
#define WIN32_LEAN_AND_MEAN
#define NOMINMAX
#include <windows.h>
#include <windowsx.h>

#include <commctrl.h>
#include <uxtheme.h>
#include <vssym32.h>

// No C++ runtime or redistributable is required: both EXEs use Win32 directly.
extern HINSTANCE instance;
extern HFONT uiFont;
extern HFONT headingFont;
extern HFONT menuFont;
extern HFONT boldMenuFont;
extern HFONT captionFont;
extern int dpi;
int Px(int value);
void InitializeUI();
void ReleaseUI();
void Fatal(LPCWSTR operation);
void SetUIFont(HWND control);
RECT MakeRect(int x, int y, int width, int height);
LRESULT ButtonFaceColors(HDC dc, HWND control);
