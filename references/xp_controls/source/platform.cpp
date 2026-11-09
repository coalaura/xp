#include "platform.h"

HINSTANCE instance;
HFONT uiFont;
HFONT headingFont;
HFONT menuFont;
HFONT boldMenuFont;
HFONT captionFont;
int dpi = 96;
static_assert(sizeof(NONCLIENTMETRICSW) == 500, "Use XP-compatible Win32 headers");

// The compiler can emit these for aggregate copies even in freestanding C++.
extern "C" void *memset(void *destination, int value, size_t count) {
    unsigned char *out = static_cast<unsigned char *>(destination);
    while (count--)
        *out++ = static_cast<unsigned char>(value);
    return destination;
}
extern "C" void *memcpy(void *destination, const void *source, size_t count) {
    unsigned char *out = static_cast<unsigned char *>(destination);
    const unsigned char *in = static_cast<const unsigned char *>(source);
    while (count--)
        *out++ = *in++;
    return destination;
}

int Px(int value) {
    return MulDiv(value, dpi, 96);
}

void Fatal(LPCWSTR operation) {
    wchar_t message[512];
    wsprintfW(message, L"%s failed (Windows error %lu).", operation, GetLastError());
    MessageBoxW(NULL, message, L"XP Controls", MB_OK | MB_ICONERROR);
    ExitProcess(1);
}

void InitializeUI() {
    instance = GetModuleHandleW(NULL);
    INITCOMMONCONTROLSEX controls = {sizeof(controls), ICC_WIN95_CLASSES | ICC_DATE_CLASSES | ICC_STANDARD_CLASSES | ICC_LINK_CLASS};
    if (!InitCommonControlsEx(&controls))
        Fatal(L"InitCommonControlsEx");
    HDC dc = GetDC(NULL);
    dpi = GetDeviceCaps(dc, LOGPIXELSY);
    ReleaseDC(NULL, dc);
    NONCLIENTMETRICSW metrics = {};
    metrics.cbSize = sizeof(metrics); // The headers are explicitly targeting XP.
    if (!SystemParametersInfoW(SPI_GETNONCLIENTMETRICS, sizeof(metrics), &metrics, 0))
        Fatal(L"Read system font");
    uiFont = CreateFontIndirectW(&metrics.lfMessageFont);
    captionFont = CreateFontIndirectW(&metrics.lfCaptionFont);
    menuFont = CreateFontIndirectW(&metrics.lfMenuFont);
    metrics.lfMenuFont.lfWeight = FW_BOLD;
    boldMenuFont = CreateFontIndirectW(&metrics.lfMenuFont);
    metrics.lfMessageFont.lfWeight = FW_BOLD;
    metrics.lfMessageFont.lfHeight = -Px(16);
    headingFont = CreateFontIndirectW(&metrics.lfMessageFont);
    if (!uiFont || !headingFont)
        Fatal(L"CreateFontIndirect");
}

void ReleaseUI() {
    DeleteObject(captionFont);
    DeleteObject(boldMenuFont);
    DeleteObject(menuFont);
    DeleteObject(headingFont);
    DeleteObject(uiFont);
}
void SetUIFont(HWND control) {
    SendMessageW(control, WM_SETFONT, reinterpret_cast<WPARAM>(uiFont), FALSE);
}
RECT MakeRect(int x, int y, int width, int height) {
    RECT r = {x, y, x + width, y + height};
    return r;
}

LRESULT ButtonFaceColors(HDC dc, HWND control) {
    SetBkColor(dc, GetSysColor(COLOR_BTNFACE));
    SetTextColor(dc, GetSysColor(IsWindowEnabled(control) ? COLOR_BTNTEXT : COLOR_GRAYTEXT));
    return reinterpret_cast<LRESULT>(GetSysColorBrush(COLOR_BTNFACE));
}
