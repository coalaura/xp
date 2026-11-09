#include "platform.h"
#include <richedit.h>

namespace {
HWND mainWindow, tabWindow, canvasWindow, titleWindow, footerWindow, hiddenHost;
int currentPage;
const wchar_t *pageNames[] = {L"Buttons", L"Inputs", L"Tabs", L"Lists", L"Scroll", L"Sliders", L"Toolbars", L"Caption", L"Frames", L"Menus", L"Text", L"Panels", L"Explorer", L"Shell"};
constexpr int pageCount = sizeof(pageNames) / sizeof(pageNames[0]);
enum Theme { Button,
             Edit,
             Combo,
             Tab,
             List,
             Header,
             Tree,
             Scroll,
             Track,
             Progress,
             Toolbar,
             Rebar,
             Status,
             Window,
             Tooltip,
             Explorer,
             Taskbar,
             StartPanel,
             Taskband,
             Tray,
             MenuBand,
             ThemeCount };
const wchar_t *themeNames[] = {L"BUTTON", L"EDIT", L"COMBOBOX", L"TAB", L"LISTVIEW", L"HEADER", L"TREEVIEW", L"SCROLLBAR", L"TRACKBAR", L"PROGRESS", L"TOOLBAR", L"REBAR", L"STATUS", L"WINDOW", L"TOOLTIP", L"EXPLORERBAR", L"TASKBAR", L"STARTPANEL", L"TASKBAND", L"TRAYNOTIFY", L"MENUBAND"};
HTHEME themeHandles[ThemeCount];

int Min(int a, int b) {
    return a < b ? a : b;
}
int Max(int a, int b) {
    return a > b ? a : b;
}
void OpenThemes() {
    for (int i = 0; i < ThemeCount; ++i) {
        if (themeHandles[i])
            CloseThemeData(themeHandles[i]);
        themeHandles[i] = OpenThemeData(canvasWindow, themeNames[i]);
    }
}

void Text(HDC dc, LPCWSTR text, RECT area, UINT align = DT_LEFT, COLORREF color = GetSysColor(COLOR_BTNTEXT)) {
    SetTextColor(dc, color);
    SetBkMode(dc, TRANSPARENT);
    DrawTextW(dc, text, -1, &area, align | DT_SINGLELINE | DT_VCENTER | DT_NOPREFIX | DT_END_ELLIPSIS);
}

bool Background(HDC dc, Theme type, int part, int state, RECT area) {
    HTHEME theme = themeHandles[type];
    if (!theme)
        return false;
    // IsThemePartDefined is not a per-state rendering check; its state
    // parameter is documented as unused. Preserve the requested state and
    // let the drawing API determine whether it can render the sample.
    return SUCCEEDED(DrawThemeBackground(theme, dc, part, state, &area, NULL));
}

void ThemeText(HDC dc, Theme type, int part, int state, LPCWSTR text, RECT area, UINT align = DT_CENTER) {
    if (FAILED(DrawThemeText(themeHandles[type], dc, part, state, text, -1,
                             align | DT_SINGLELINE | DT_VCENTER | DT_NOPREFIX, 0, &area))) {
        Text(dc, text, area, align);
    }
}

SIZE NaturalSize(HDC dc, Theme type, int part, int state, int width, int height) {
    SIZE size = {Px(width), Px(height)};
    if (themeHandles[type]) {
        SIZE found = {};
        if (SUCCEEDED(GetThemePartSize(themeHandles[type], dc, part, state, NULL, TS_TRUE, &found)) && found.cx > 0 && found.cy > 0)
            size = found;
    }
    return size;
}

enum Kind { Plain,
            Glyph,
            Push,
            Choice,
            Group,
            Input,
            Password,
            ComboInput,
            ListItem,
            TreeItem,
            HeaderItem,
            TabItem,
            CaptionButton,
            CaptionStrip,
            ToolbarButton,
            HorizontalProgress,
            VerticalProgress,
            FocusPush,
            FocusChoice };

struct Sheet {
    HDC dc;
    int width, y, columns;
    Sheet(HDC context, int w) : dc(context), width(w), y(Px(8)), columns(4) {}
    void Note(LPCWSTR note) {
        Text(dc, note, MakeRect(Px(12), y, width - Px(24), Px(22)), DT_LEFT, GetSysColor(COLOR_GRAYTEXT));
        y += Px(26);
    }
    void Head(LPCWSTR a, LPCWSTR b = NULL, LPCWSTR c = NULL, LPCWSTR d = NULL, LPCWSTR e = NULL, LPCWSTR f = NULL) {
        LPCWSTR labels[] = {a, b, c, d, e, f};
        columns = 1;
        for (int i = 1; i < 6; ++i)
            if (labels[i])
                columns = i + 1;
        int cell = (width - Px(184)) / columns;
        for (int i = 0; i < columns; ++i)
            Text(dc, labels[i] ? labels[i] : L"", MakeRect(Px(176) + i * cell, y, cell, Px(24)), DT_CENTER);
        y += Px(27);
    }
    void Row(LPCWSTR label, Theme type, int part, Kind kind, int w, int h,
             int s1, int s2 = -1, int s3 = -1, int s4 = -1, int s5 = -1, int s6 = -1, int rowHeight = 38);
};

bool DrawSample(HDC dc, Theme type, int part, int state, Kind kind, RECT r) {
    const int width = r.right - r.left, height = r.bottom - r.top;
    if (kind == CaptionButton) {
        RECT caption = r;
        InflateRect(&caption, Px(9), Px(3));
        if (!Background(dc, Window, WP_CAPTION, CS_ACTIVE, caption))
            return false;
    }
    if (kind == HorizontalProgress || kind == VerticalProgress) {
        if (!Background(dc, Progress, kind == HorizontalProgress ? PP_BAR : PP_BARVERT, 0, r))
            return false;
        RECT inner = r;
        InflateRect(&inner, -Px(3), -Px(3));
        int chunk = Px(6), gap = Px(2);
        int size = 0;
        if (SUCCEEDED(GetThemeInt(themeHandles[Progress], 0, 0, TMT_PROGRESSCHUNKSIZE, &size)) && size > 0)
            chunk = size;
        if (SUCCEEDED(GetThemeInt(themeHandles[Progress], 0, 0, TMT_PROGRESSSPACESIZE, &size)) && size >= 0)
            gap = size;
        if (kind == HorizontalProgress) {
            int end = inner.left + (inner.right - inner.left) * state / 100;
            for (int x = inner.left; x < end; x += chunk + gap)
                if (!Background(dc, Progress, PP_CHUNK, 0, MakeRect(x, inner.top, Min(chunk, end - x), inner.bottom - inner.top)))
                    return false;
        } else {
            int end = inner.bottom - (inner.bottom - inner.top) * state / 100;
            for (int y = inner.bottom; y > end; y -= chunk + gap)
                if (!Background(dc, Progress, PP_CHUNKVERT, 0, MakeRect(inner.left, Max(end, y - chunk), inner.right - inner.left, Min(chunk, y - end))))
                    return false;
        }
        return true;
    }
    if (kind == ComboInput) {
        if (!Background(dc, Edit, EP_EDITTEXT, state == CBXS_DISABLED ? ETS_DISABLED : ETS_NORMAL, r))
            return false;
        RECT arrow = {r.right - Px(18), r.top + Px(1), r.right - Px(1), r.bottom - Px(1)};
        if (!Background(dc, Combo, CP_DROPDOWNBUTTON, state, arrow))
            return false;
        RECT content = {r.left + Px(5), r.top, arrow.left - Px(3), r.bottom};
        Text(dc, L"Option", content, DT_LEFT, GetSysColor(state == 4 ? COLOR_GRAYTEXT : COLOR_WINDOWTEXT));
        return true;
    }
    if (kind == ListItem || kind == TreeItem) {
        // XP draws these selection backgrounds using system colors. Newer
        // Explorer-only theme parts are deliberately not substituted here.
        int bg = state == 3 ? COLOR_HIGHLIGHT : state == 5 ? COLOR_BTNFACE
                                                           : COLOR_WINDOW;
        int fg = state == 3 ? COLOR_HIGHLIGHTTEXT : state == 4 ? COLOR_GRAYTEXT
                                                               : COLOR_WINDOWTEXT;
        FillRect(dc, &r, GetSysColorBrush(bg));
        RECT text = r;
        text.left += Px(6);
        LPCWSTR caption = kind == ListItem ? L"List item" : L"Tree item";
        if (state == 2) {
            LOGFONTW lf = {};
            GetObjectW(uiFont, sizeof(lf), &lf);
            lf.lfUnderline = TRUE;
            HFONT underlined = CreateFontIndirectW(&lf);
            HGDIOBJ old = SelectObject(dc, underlined);
            Text(dc, caption, text, DT_LEFT, GetSysColor(COLOR_HOTLIGHT));
            SelectObject(dc, old);
            DeleteObject(underlined);
        } else
            Text(dc, caption, text, DT_LEFT, GetSysColor(fg));
        return true;
    }
    if (kind == Choice || kind == FocusChoice) {
        SIZE glyph = NaturalSize(dc, type, part, state, 13, 13);
        RECT g = MakeRect(r.left + Px(8), r.top + (height - glyph.cy) / 2, glyph.cx, glyph.cy);
        if (!Background(dc, type, part, state, g))
            return false;
        RECT text = {g.right + Px(4), r.top, r.right, r.bottom};
        ThemeText(dc, type, part, state, L"Option", text, DT_LEFT);
        if (kind == FocusChoice) {
            SIZE extent = {};
            GetTextExtentPoint32W(dc, L"Option", 6, &extent);
            RECT focus = MakeRect(text.left - Px(1), text.top + (height - extent.cy) / 2 - Px(1), extent.cx + Px(2), extent.cy + Px(2));
            DrawFocusRect(dc, &focus);
        }
        return true;
    }
    if (kind == Group) {
        int saved = SaveDC(dc);
        ExcludeClipRect(dc, r.left + Px(7), r.top, r.left + Px(48), r.top + Px(12));
        RECT border = r;
        border.top += Px(6);
        bool drawn = Background(dc, type, part, state, border);
        RestoreDC(dc, saved);
        if (!drawn)
            return false;
        RECT title = MakeRect(r.left + Px(9), r.top - Px(2), Px(48), Px(16));
        ThemeText(dc, type, part, state, L"Group", title, DT_LEFT);
        return true;
    }
    if (kind == Glyph || kind == CaptionButton) {
        SIZE size = NaturalSize(dc, type, part, state, width * 96 / dpi, height * 96 / dpi);
        int gw = Min(size.cx, width), gh = Min(size.cy, height);
        r = MakeRect(r.left + (width - gw) / 2, r.top + (height - gh) / 2, gw, gh);
    }
    if (!Background(dc, type, part, state, r))
        return false;
    switch (kind) {
    case Push:
    case FocusPush:
        ThemeText(dc, type, part, state, state == PBS_DEFAULTED ? L"Default" : L"Button", r);
        if (kind == FocusPush) {
            InflateRect(&r, -Px(4), -Px(4));
            DrawFocusRect(dc, &r);
        }
        break;
    case Input:
    case Password: {
        RECT content = r;
        content.left += Px(5);
        content.right -= Px(4);
        LPCWSTR text = kind == Password ? L"********" : L"Sample text";
        if (state == ETS_SELECTED) {
            RECT selection = content;
            InflateRect(&selection, 0, -Px(3));
            FillRect(dc, &selection, GetSysColorBrush(COLOR_HIGHLIGHT));
            Text(dc, text, content, DT_LEFT, GetSysColor(COLOR_HIGHLIGHTTEXT));
        } else {
            Text(dc, text, content, DT_LEFT, GetSysColor(state == ETS_DISABLED ? COLOR_GRAYTEXT : COLOR_WINDOWTEXT));
        }
        if (state == ETS_FOCUSED) {
            RECT caret = MakeRect(content.left + Px(1), content.top + Px(4), Px(1), height - Px(8));
            FillRect(dc, &caret, GetSysColorBrush(COLOR_WINDOWTEXT));
        }
        break;
    }
    case HeaderItem:
        ThemeText(dc, type, part, state, L"Column", r);
        break;
    case TabItem:
        ThemeText(dc, type, part, state, L"Tab", r);
        break;
    case CaptionStrip: {
        RECT text = r;
        text.left += Px(6);
        HGDIOBJ old = SelectObject(dc, captionFont);
        ThemeText(dc, type, part, state, L"Window", text, DT_LEFT);
        SelectObject(dc, old);
        break;
    }
    case ToolbarButton:
        ThemeText(dc, type, part, state, L"Action", r);
        break;
    default:
        break;
    }
    return true;
}

void Sheet::Row(LPCWSTR label, Theme type, int part, Kind kind, int w, int h, int s1, int s2, int s3, int s4, int s5, int s6, int rowHeight) {
    int states[] = {s1, s2, s3, s4, s5, s6};
    Text(dc, label, MakeRect(Px(12), y, Px(156), Px(rowHeight)));
    int cell = (width - Px(184)) / columns;
    for (int i = 0; i < columns; ++i) {
        RECT cellRect = MakeRect(Px(176) + i * cell, y, cell, Px(rowHeight));
        if (states[i] < 0) {
            Text(dc, L"-", cellRect, DT_CENTER, GetSysColor(COLOR_GRAYTEXT));
            continue;
        }
        int sw = Min(Px(w), cell - Px(12)), sh = Min(Px(h), Px(rowHeight - 6));
        RECT r = MakeRect(cellRect.left + (cell - sw) / 2, y + (Px(rowHeight) - sh) / 2, sw, sh);
        if (!DrawSample(dc, type, part, states[i], kind, r)) {
            FillRect(dc, &cellRect, GetSysColorBrush(COLOR_BTNFACE));
            Text(dc, L"draw failed", cellRect, DT_CENTER, GetSysColor(COLOR_GRAYTEXT));
        }
    }
    y += Px(rowHeight);
}

void Buttons(Sheet &s) {
    s.Note(L"Fixed XP theme states. Hovering or clicking a sample does not change it.");
    s.Head(L"Normal", L"Hot", L"Pressed", L"Disabled", L"Default");
    s.Row(L"Push button", Button, BP_PUSHBUTTON, Push, 112, 26, 1, 2, 3, 4, 5);
    s.Row(L"Checkbox / empty", Button, BP_CHECKBOX, Choice, 112, 26, 1, 2, 3, 4);
    s.Row(L"Checkbox / checked", Button, BP_CHECKBOX, Choice, 112, 26, 5, 6, 7, 8);
    s.Row(L"Checkbox / mixed", Button, BP_CHECKBOX, Choice, 112, 26, 9, 10, 11, 12);
    s.Row(L"Radio / empty", Button, BP_RADIOBUTTON, Choice, 112, 26, 1, 2, 3, 4);
    s.Row(L"Radio / selected", Button, BP_RADIOBUTTON, Choice, 112, 26, 5, 6, 7, 8);
    s.Row(L"Group box", Button, BP_GROUPBOX, Group, 112, 34, 1, -1, -1, 2, -1, -1, 44);
    s.Head(L"Button", L"Default button", L"Checkbox", L"Checked", L"Radio");
    s.Row(L"Keyboard focus", Button, BP_PUSHBUTTON, FocusPush, 112, 26, 1, 5);
    s.Row(L"Focus on choice", Button, BP_CHECKBOX, FocusChoice, 112, 26, -1, -1, 1, 5);
    s.Row(L"Focus on radio", Button, BP_RADIOBUTTON, FocusChoice, 112, 26, -1, -1, -1, -1, 5);
    s.Note(L"Primary = default button. There is no separate disabled-default theme state.");
}

void Inputs(Sheet &s) {
    s.Note(L"Edit fields show native theme borders; the caret and text selection are fixed GDI samples.");
    s.Head(L"Normal", L"Hot", L"Focused", L"Disabled", L"Read-only", L"Selected");
    s.Row(L"Edit border", Edit, EP_EDITTEXT, Plain, 126, 23, 1, 2, 5, 4, 6, 3);
    s.Row(L"Text input", Edit, EP_EDITTEXT, Input, 126, 23, 1, 2, 5, 4, 6, 3);
    s.Row(L"Password input", Edit, EP_EDITTEXT, Password, 126, 23, 1, 2, 5, 4, 6, 3);
    s.Head(L"Normal", L"Hot", L"Pressed", L"Disabled");
    s.Row(L"Combo box", Combo, CP_DROPDOWNBUTTON, ComboInput, 150, 24, 1, 2, 3, 4, -1, -1, 42);
    s.Row(L"Dropdown arrow", Combo, CP_DROPDOWNBUTTON, Plain, 18, 22, 1, 2, 3, 4);
}

// SPIN and PAGE use the same four state IDs and small native arrow parts.
void NamedPartRow(Sheet &s, LPCWSTR label, LPCWSTR className, int part, int w, int h, int count, int rowHeight = 34) {
    HTHEME theme = OpenThemeData(canvasWindow, className);
    Text(s.dc, label, MakeRect(Px(12), s.y, Px(156), Px(rowHeight)));
    int cell = (s.width - Px(184)) / s.columns;
    for (int i = 0; i < s.columns; ++i) {
        RECT area = MakeRect(Px(176) + i * cell + (cell - Px(w)) / 2, s.y + (Px(rowHeight - h)) / 2, Px(w), Px(h));
        RECT cellRect = MakeRect(Px(176) + i * cell, s.y, cell, Px(rowHeight));
        if (i >= count) {
            Text(s.dc, L"-", cellRect, DT_CENTER, GetSysColor(COLOR_GRAYTEXT));
        } else if (!theme || FAILED(DrawThemeBackground(theme, s.dc, part, i + 1, &area, NULL))) {
            FillRect(s.dc, &cellRect, GetSysColorBrush(COLOR_BTNFACE));
            Text(s.dc, L"draw failed", cellRect, DT_CENTER, GetSysColor(COLOR_GRAYTEXT));
        }
    }
    if (theme)
        CloseThemeData(theme);
    s.y += Px(rowHeight);
}

void Tabs(Sheet &s) {
    s.Note(L"Tab parts include the separate edge variants used when assembling a tab strip.");
    s.Head(L"Normal", L"Hot", L"Selected", L"Disabled", L"Focused");
    LPCWSTR labels[] = {L"Tab", L"Left edge", L"Right edge", L"Both edges", L"Top tab", L"Top / left edge", L"Top / right edge", L"Top / both edges"};
    for (int i = 0; i < 8; ++i)
        s.Row(labels[i], Tab, i + 1, TabItem, 112, 26, 1, 2, 3, 4, 5);
    s.Head(L"Pane", L"Body");
    s.Row(L"Container", Tab, TABP_PANE, Plain, 260, 52, 0, -1, -1, -1, -1, -1, 60);
    s.Row(L"Body background", Tab, TABP_BODY, Plain, 260, 30, -1, 0);
}

void Lists(Sheet &s) {
    s.Note(L"Selection samples use XP system colors. Header and expand/collapse glyphs use the theme API.");
    s.Head(L"Normal", L"Hot text", L"Selected", L"Disabled", L"Inactive selection");
    s.Row(L"List item", List, LVP_LISTITEM, ListItem, 126, 23, 1, 2, 3, 4, 5);
    s.Row(L"Tree item", Tree, TVP_TREEITEM, TreeItem, 126, 23, 1, 2, 3, 4, 5);
    s.Head(L"Normal", L"Hot", L"Pressed");
    s.Row(L"Header", Header, HP_HEADERITEM, HeaderItem, 170, 26, 1, 2, 3);
    s.Row(L"Header / left edge", Header, HP_HEADERITEMLEFT, HeaderItem, 170, 26, 1, 2, 3);
    s.Row(L"Header / right edge", Header, HP_HEADERITEMRIGHT, HeaderItem, 170, 26, 1, 2, 3);
    s.Head(L"Up / collapsed", L"Down / expanded");
    s.Row(L"Sort arrow", Header, HP_HEADERSORTARROW, Glyph, 20, 20, 1, 2);
    s.Row(L"Tree expander", Tree, TVP_GLYPH, Glyph, 16, 16, 1, 2);
    s.Head(L"Group", L"Detail", L"Sorted detail", L"Empty text");
    s.Row(L"List group part", List, LVP_LISTGROUP, Plain, 150, 30, 0);
    s.Row(L"Detail part", List, LVP_LISTDETAIL, Plain, 150, 30, -1, 0);
    s.Note(L"A dash means no state in this column; 'draw failed' means the theme could not render the sample.");
}

void Scrollbars(Sheet &s) {
    s.Head(L"Normal", L"Hot", L"Pressed", L"Disabled");
    LPCWSTR arrows[] = {L"Arrow / up", L"Arrow / down", L"Arrow / left", L"Arrow / right"};
    for (int i = 0; i < 4; ++i)
        s.Row(arrows[i], Scroll, SBP_ARROWBTN, Plain, 17, 17, 1 + i * 4, 2 + i * 4, 3 + i * 4, 4 + i * 4, -1, -1, 29);
    s.Row(L"Thumb / horizontal", Scroll, SBP_THUMBBTNHORZ, Plain, 100, 17, 1, 2, 3, 4, -1, -1, 30);
    s.Row(L"Thumb / vertical", Scroll, SBP_THUMBBTNVERT, Plain, 17, 27, 1, 2, 3, 4, -1, -1, 35);
    s.Row(L"Track / lower H", Scroll, SBP_LOWERTRACKHORZ, Plain, 120, 17, 1, 2, 3, 4, -1, -1, 29);
    s.Row(L"Track / upper H", Scroll, SBP_UPPERTRACKHORZ, Plain, 120, 17, 1, 2, 3, 4, -1, -1, 29);
    s.Row(L"Track / lower V", Scroll, SBP_LOWERTRACKVERT, Plain, 17, 26, 1, 2, 3, 4, -1, -1, 34);
    s.Row(L"Track / upper V", Scroll, SBP_UPPERTRACKVERT, Plain, 17, 26, 1, 2, 3, 4, -1, -1, 34);
    s.Row(L"Gripper / horizontal", Scroll, SBP_GRIPPERHORZ, Glyph, 25, 17, 1, 2, 3, 4, -1, -1, 29);
    s.Row(L"Gripper / vertical", Scroll, SBP_GRIPPERVERT, Glyph, 17, 25, 1, 2, 3, 4, -1, -1, 33);
    s.Head(L"Right aligned", L"Left aligned");
    s.Row(L"Resize grip", Scroll, SBP_SIZEBOX, Glyph, 20, 20, 1, 2, -1, -1, -1, -1, 30);
}

void Sliders(Sheet &s) {
    s.Head(L"Normal", L"Hot", L"Pressed", L"Focused", L"Disabled");
    LPCWSTR labels[] = {L"Thumb / rectangle", L"Thumb / bottom", L"Thumb / top", L"Thumb / vertical", L"Thumb / left", L"Thumb / right"};
    for (int i = 0; i < 6; ++i)
        s.Row(labels[i], Track, TKP_THUMB + i, Glyph, 36, 28, 1, 2, 3, 4, 5, -1, 36);
    s.Head(L"Horizontal", L"Vertical");
    s.Row(L"Track", Track, TKP_TRACK, Plain, 200, 5, 1, -1, -1, -1, -1, -1, 24);
    s.Row(L"Vertical track", Track, TKP_TRACKVERT, Plain, 5, 32, -1, 1, -1, -1, -1, -1, 38);
    s.Head(L"0%", L"25%", L"50%", L"75%", L"100%");
    s.Row(L"Progress", Progress, PP_BAR, HorizontalProgress, 126, 18, 0, 25, 50, 75, 100, -1, 36);
    s.Row(L"Vertical progress", Progress, PP_BARVERT, VerticalProgress, 18, 62, 0, 25, 50, 75, 100, -1, 70);
    s.Note(L"XP progress bars have no Vista-style paused or error states. Percentages are fill amounts.");
}

void Toolbars(Sheet &s) {
    s.Head(L"Normal", L"Hot", L"Pressed", L"Disabled", L"Checked", L"Hot + checked");
    s.Row(L"Toolbar button", Toolbar, TP_BUTTON, ToolbarButton, 96, 26, 1, 2, 3, 4, 5, 6);
    s.Row(L"Dropdown button", Toolbar, TP_DROPDOWNBUTTON, ToolbarButton, 96, 26, 1, 2, 3, 4, 5, 6);
    s.Row(L"Split / main", Toolbar, TP_SPLITBUTTON, ToolbarButton, 96, 26, 1, 2, 3, 4, 5, 6);
    s.Row(L"Split / arrow", Toolbar, TP_SPLITBUTTONDROPDOWN, Plain, 18, 26, 1, 2, 3, 4, 5, 6);
    s.Head(L"Normal", L"Hot", L"Pressed");
    s.Row(L"Chevron", Rebar, RP_CHEVRON, Glyph, 26, 22, 1, 2, 3);
    s.Row(L"Vertical chevron", Rebar, RP_CHEVRONVERT, Glyph, 22, 26, 1, 2, 3);
    s.Head(L"Horizontal", L"Vertical");
    s.Row(L"Toolbar separator", Toolbar, TP_SEPARATOR, Plain, 5, 25, 0);
    s.Row(L"Vertical separator", Toolbar, TP_SEPARATORVERT, Plain, 120, 5, -1, 0, -1, -1, -1, -1, 24);
    s.Row(L"Rebar grip", Rebar, RP_GRIPPER, Plain, 6, 25, 0);
    s.Row(L"Rebar grip / vertical", Rebar, RP_GRIPPERVERT, Plain, 120, 6, -1, 0, -1, -1, -1, -1, 24);
    s.Head(L"Pane", L"Gripper pane", L"Gripper");
    s.Row(L"Status bar", Status, SP_PANE, Plain, 170, 22, 0);
    s.Row(L"Sizing grip", Status, SP_GRIPPER, Glyph, 22, 22, -1, -1, 0);
}

void CaptionButtons(Sheet &s) {
    s.Head(L"Normal", L"Hot", L"Pressed", L"Disabled");
    LPCWSTR names[] = {L"System", L"MDI / system", L"Minimize", L"MDI / minimize", L"Maximize", L"Close", L"Small close", L"MDI / close", L"Restore", L"MDI / restore", L"Help", L"MDI / help"};
    for (int i = 0; i < 12; ++i)
        s.Row(names[i], Window, WP_SYSBUTTON + i, CaptionButton, 26, 23, 1, 2, 3, 4, -1, -1, 36);
    s.Note(L"The application's own title bar is native and supports minimize, maximize/restore, and close.");
}

void Frames(Sheet &s) {
    s.Head(L"Active", L"Inactive", L"Disabled");
    LPCWSTR names[] = {L"Caption", L"Small caption", L"Minimized caption", L"Small / minimized", L"Maximized caption", L"Small / maximized"};
    for (int i = 0; i < 6; ++i)
        s.Row(names[i], Window, WP_CAPTION + i, CaptionStrip, 200, 24, 1, 2, 3, -1, -1, -1, 33);
    s.Head(L"Active", L"Inactive");
    LPCWSTR frames[] = {L"Frame / left", L"Frame / right", L"Frame / bottom", L"Small frame / left", L"Small frame / right", L"Small frame / bottom"};
    for (int i = 0; i < 6; ++i)
        s.Row(frames[i], Window, WP_FRAMELEFT + i, Plain, i % 3 == 2 ? 220 : 5, i % 3 == 2 ? 5 : 23, 1, 2, -1, -1, -1, -1, 30);
    s.Row(L"Dialog background", Window, WP_DIALOG, Plain, 220, 26, 0, -1, -1, -1, -1, -1, 32);
}

void MenuCell(HDC dc, RECT r, int state, int variant) {
    bool selected = state == 2 || state == 4, disabled = state >= 3;
    FillRect(dc, &r, GetSysColorBrush(selected ? COLOR_HIGHLIGHT : COLOR_MENU));
    int color = disabled ? COLOR_GRAYTEXT : selected ? COLOR_HIGHLIGHTTEXT
                                                     : COLOR_MENUTEXT;
    RECT text = r;
    text.left += Px(24);
    text.right -= Px(16);
    HGDIOBJ old = SelectObject(dc, variant == 4 ? boldMenuFont : menuFont);
    Text(dc, variant == 4 ? L"Default" : L"Menu item", text, DT_LEFT, GetSysColor(color));
    SelectObject(dc, old);
    if (variant == 1 || variant == 2 || variant == 3) {
        RECT glyph = variant == 3 ? MakeRect(r.right - Px(15), r.top + Px(4), Px(11), r.bottom - r.top - Px(8)) : MakeRect(r.left + Px(4), r.top + Px(4), Px(14), r.bottom - r.top - Px(8));
        // DrawFrameControl provides the system's menu glyph shapes.
        HDC maskDC = CreateCompatibleDC(dc);
        HBITMAP mask = CreateBitmap(glyph.right - glyph.left, glyph.bottom - glyph.top, 1, 1, NULL);
        HGDIOBJ previous = SelectObject(maskDC, mask);
        RECT local = MakeRect(0, 0, glyph.right - glyph.left, glyph.bottom - glyph.top);
        DrawFrameControl(maskDC, &local, DFC_MENU, variant == 1 ? DFCS_MENUCHECK : variant == 2 ? DFCS_MENUBULLET
                                                                                                : DFCS_MENUARROW);
        COLORREF oldText = SetTextColor(dc, GetSysColor(color));
        COLORREF oldBg = SetBkColor(dc, GetSysColor(selected ? COLOR_HIGHLIGHT : COLOR_MENU));
        BitBlt(dc, glyph.left, glyph.top, local.right, local.bottom, maskDC, 0, 0, SRCCOPY);
        SetTextColor(dc, oldText);
        SetBkColor(dc, oldBg);
        SelectObject(maskDC, previous);
        DeleteObject(mask);
        DeleteDC(maskDC);
    }
}

void Menus(Sheet &s) {
    s.Note(L"Fixed menu samples use Windows' menu font, system colors, and native menu glyph rendering.");
    s.Head(L"Normal", L"Selected", L"Disabled", L"Disabled + selected");
    LPCWSTR labels[] = {L"Menu item", L"Checked item", L"Radio item", L"Submenu", L"Default item"};
    int cell = (s.width - Px(184)) / 4;
    for (int row = 0; row < 5; ++row) {
        Text(s.dc, labels[row], MakeRect(Px(12), s.y, Px(156), Px(44)));
        for (int i = 0; i < 4; ++i)
            MenuCell(s.dc, MakeRect(Px(176) + i * cell + Px(4), s.y + Px(8), cell - Px(8), Px(26)), i + 1, row);
        s.y += Px(44);
    }
    s.Note(L"Popup menu / normal items, separator, check, selection, disabled item, and submenu:");
    RECT popup = MakeRect(Px(186), s.y, Px(248), Px(160));
    FillRect(s.dc, &popup, GetSysColorBrush(COLOR_MENU));
    FrameRect(s.dc, &popup, GetSysColorBrush(COLOR_3DSHADOW));
    for (int i = 0; i < 5; ++i)
        MenuCell(s.dc, MakeRect(popup.left + Px(2), popup.top + Px(3 + i * 29), Px(244), Px(26)), i == 2 ? 2 : i == 3 ? 3
                                                                                                                      : 1,
                 i == 1 ? 1 : i == 4 ? 3
                                     : 0);
    RECT line = MakeRect(Px(478), s.y + Px(20), Px(190), Px(2));
    DrawEdge(s.dc, &line, EDGE_ETCHED, BF_TOP);
    Text(s.dc, L"Etched separator", MakeRect(Px(478), s.y + Px(35), Px(190), Px(24)));
}

void Typography(Sheet &s) {
    s.Note(L"Typography and GDI samples. Headings are font choices, not a separate XP widget class.");
    s.Head(L"Normal", L"Disabled", L"Selected", L"Inactive selection");
    int cell = (s.width - Px(184)) / 4;
    LPCWSTR labels[] = {L"Body text", L"Bold heading", L"Link color", L"Focus rectangle", L"Horizontal separator", L"Vertical separator", L"Sunken edge", L"Raised edge"};
    for (int row = 0; row < 8; ++row) {
        Text(s.dc, labels[row], MakeRect(Px(12), s.y, Px(156), Px(48)));
        for (int col = 0; col < 4; ++col) {
            RECT r = MakeRect(Px(176) + col * cell + Px(8), s.y + Px(8), cell - Px(16), Px(30));
            if (row <= 2) {
                int bg = col == 2 ? COLOR_HIGHLIGHT : col == 3 ? COLOR_BTNFACE
                                                               : COLOR_BTNFACE;
                FillRect(s.dc, &r, GetSysColorBrush(bg));
                HGDIOBJ old = SelectObject(s.dc, row == 1 ? headingFont : uiFont);
                Text(s.dc, row == 1 ? L"Heading" : row == 2 ? L"Example link"
                                                            : L"Regular text",
                     r, DT_CENTER, GetSysColor(col == 1 ? COLOR_GRAYTEXT : col == 2 ? COLOR_HIGHLIGHTTEXT
                                                                       : row == 2   ? COLOR_HOTLIGHT
                                                                                    : COLOR_BTNTEXT));
                SelectObject(s.dc, old);
            } else if (col == 0) {
                if (row == 3)
                    DrawFocusRect(s.dc, &r);
                if (row == 4) {
                    r.top += (r.bottom - r.top) / 2;
                    DrawEdge(s.dc, &r, EDGE_ETCHED, BF_TOP);
                }
                if (row == 5) {
                    r.left += (r.right - r.left) / 2;
                    DrawEdge(s.dc, &r, EDGE_ETCHED, BF_LEFT);
                }
                if (row == 6)
                    DrawEdge(s.dc, &r, EDGE_SUNKEN, BF_RECT);
                if (row == 7)
                    DrawEdge(s.dc, &r, EDGE_RAISED, BF_RECT);
            } else
                Text(s.dc, L"-", r, DT_CENTER, GetSysColor(COLOR_GRAYTEXT));
        }
        s.y += Px(48);
    }
}

// Hidden controls are used only as native rendering sources for complete
// panels. They are never shown, focused, or offered to the user for input.
HWND panelControls[6];
void CreatePanels() {
    hiddenHost = CreateWindowExW(0, L"STATIC", L"", WS_POPUP, 0, 0, Px(1000), Px(800), NULL, NULL, instance, NULL);
    panelControls[0] = CreateWindowExW(0, MONTHCAL_CLASSW, L"", WS_CHILD | WS_VISIBLE | MCS_NOTODAY, 0, 0, Px(220), Px(180), hiddenHost, NULL, instance, NULL);
    panelControls[1] = CreateWindowExW(0, MONTHCAL_CLASSW, L"", WS_CHILD | WS_VISIBLE | WS_DISABLED | MCS_NOTODAY, 0, 0, Px(220), Px(180), hiddenHost, NULL, instance, NULL);
    SYSTEMTIME date = {2001, 10, 4, 25, 10, 31, 0, 0};
    for (int i = 0; i < 2; ++i) {
        SetUIFont(panelControls[i]);
        SendMessageW(panelControls[i], MCM_SETCURSEL, 0, reinterpret_cast<LPARAM>(&date));
        SendMessageW(panelControls[i], MCM_SETTODAY, 0, reinterpret_cast<LPARAM>(&date));
        RECT r = {};
        SendMessageW(panelControls[i], MCM_GETMINREQRECT, 0, reinterpret_cast<LPARAM>(&r));
        SetWindowPos(panelControls[i], NULL, 0, 0, r.right, r.bottom, SWP_NOZORDER | SWP_NOMOVE);
    }
    for (int i = 2; i < 4; ++i) {
        panelControls[i] = CreateWindowExW(0, DATETIMEPICK_CLASSW, L"", WS_CHILD | WS_VISIBLE | (i == 3 ? WS_DISABLED : 0), 0, 0, Px(200), Px(24), hiddenHost, NULL, instance, NULL);
        SetUIFont(panelControls[i]);
        SendMessageW(panelControls[i], DTM_SETFORMATW, 0, reinterpret_cast<LPARAM>(L"dd MMM yyyy"));
        SendMessageW(panelControls[i], DTM_SETSYSTEMTIME, GDT_VALID, reinterpret_cast<LPARAM>(&date));
    }
    for (int i = 4; i < 6; ++i) {
        panelControls[i] = CreateWindowExW(WS_EX_CLIENTEDGE, L"LISTBOX", L"", WS_CHILD | WS_VISIBLE | LBS_NOTIFY | (i == 5 ? WS_DISABLED : 0), 0, 0, Px(200), Px(120), hiddenHost, NULL, instance, NULL);
        SetUIFont(panelControls[i]);
        LPCWSTR items[] = {L"First item", L"Second item", L"Third item", L"Fourth item", L"Fifth item"};
        for (LPCWSTR item : items)
            SendMessageW(panelControls[i], LB_ADDSTRING, 0, reinterpret_cast<LPARAM>(item));
        SendMessageW(panelControls[i], LB_SETCURSEL, 1, 0);
    }
}

void PrintPanel(HDC dc, HWND control, int x, int y) {
    RECT window = {};
    GetWindowRect(control, &window);
    RECT r = MakeRect(0, 0, window.right - window.left, window.bottom - window.top);
    int saved = SaveDC(dc);
    SetViewportOrgEx(dc, x, y, NULL);
    IntersectClipRect(dc, 0, 0, r.right, r.bottom);
    FillRect(dc, &r, GetSysColorBrush(COLOR_BTNFACE));
    SendMessageW(control, WM_PRINT, reinterpret_cast<WPARAM>(dc), PRF_CLIENT | PRF_NONCLIENT | PRF_ERASEBKGND | PRF_CHILDREN);
    RestoreDC(dc, saved);
}

void Panels(Sheet &s) {
    s.Note(L"Complete native controls printed into a static sheet. No sample accepts input.");
    s.Head(L"Normal", L"Disabled");
    int x1 = Px(218), x2 = Px(218) + (s.width - Px(184)) / 2;
    Text(s.dc, L"Month calendar", MakeRect(Px(12), s.y, Px(156), Px(24)));
    PrintPanel(s.dc, panelControls[0], x1, s.y);
    PrintPanel(s.dc, panelControls[1], x2, s.y);
    s.y += Px(188);
    Text(s.dc, L"Date picker", MakeRect(Px(12), s.y, Px(156), Px(24)));
    PrintPanel(s.dc, panelControls[2], x1, s.y);
    PrintPanel(s.dc, panelControls[3], x2, s.y);
    s.y += Px(42);
    Text(s.dc, L"List box", MakeRect(Px(12), s.y, Px(156), Px(24)));
    PrintPanel(s.dc, panelControls[4], x1, s.y);
    PrintPanel(s.dc, panelControls[5], x2, s.y);
    s.y += Px(130);
    s.Head(L"Normal", L"Link");
    s.Row(L"Tooltip background", Tooltip, TTP_STANDARD, Plain, 220, 30, 1, 2);
    s.Note(L"If a control draws disabled and enabled identically on XP, those samples will also match.");
}

void ExplorerParts(Sheet &s) {
    s.Head(L"Normal", L"Hot", L"Pressed", L"Selected", L"Selected + hot", L"Selected + pressed");
    s.Row(L"Header / close", Explorer, EBP_HEADERCLOSE, Glyph, 24, 24, 1, 2, 3);
    s.Row(L"Header / pin", Explorer, EBP_HEADERPIN, Glyph, 24, 24, 1, 2, 3, 4, 5, 6);
    s.Row(L"Bar menu", Explorer, EBP_IEBARMENU, Glyph, 24, 24, 1, 2, 3);
    s.Row(L"Normal / collapse", Explorer, EBP_NORMALGROUPCOLLAPSE, Glyph, 24, 24, 1, 2, 3);
    s.Row(L"Normal / expand", Explorer, EBP_NORMALGROUPEXPAND, Glyph, 24, 24, 1, 2, 3);
    s.Row(L"Special / collapse", Explorer, EBP_SPECIALGROUPCOLLAPSE, Glyph, 24, 24, 1, 2, 3);
    s.Row(L"Special / expand", Explorer, EBP_SPECIALGROUPEXPAND, Glyph, 24, 24, 1, 2, 3);
    s.Head(L"Header", L"Normal group", L"Special group");
    s.Row(L"Background", Explorer, EBP_HEADERBACKGROUND, Plain, 210, 30, 0);
    s.Row(L"Normal group body", Explorer, EBP_NORMALGROUPBACKGROUND, Plain, 210, 30, -1, 0);
    s.Row(L"Normal group head", Explorer, EBP_NORMALGROUPHEAD, Plain, 210, 30, -1, 0);
    s.Row(L"Special group body", Explorer, EBP_SPECIALGROUPBACKGROUND, Plain, 210, 30, -1, -1, 0);
    s.Row(L"Special group head", Explorer, EBP_SPECIALGROUPHEAD, Plain, 210, 30, -1, -1, 0);
}

void ShellParts(Sheet &s) {
    s.Note(L"XP shell theme parts; these are individual pieces, not full Explorer or Start menu replicas.");
    s.Head(L"Bottom", L"Right", L"Top", L"Left");
    s.Row(L"Taskbar / bottom", Taskbar, 1, Plain, 150, 24, 0);
    s.Row(L"Taskbar / right", Taskbar, 2, Plain, 24, 28, -1, 0);
    s.Row(L"Taskbar / top", Taskbar, 3, Plain, 150, 24, -1, -1, 0);
    s.Row(L"Taskbar / left", Taskbar, 4, Plain, 24, 28, -1, -1, -1, 0);
    s.Head(L"Background", L"Alternate / detail", L"Normal", L"Hot", L"Pressed");
    s.Row(L"Notification area", Tray, 1, Plain, 126, 28, 0);
    s.Row(L"Animated area", Tray, 2, Plain, 126, 28, -1, 0);
    s.Row(L"Start / user pane", StartPanel, 1, Plain, 126, 30, 0);
    s.Row(L"Start / programs", StartPanel, 4, Plain, 126, 30, 0);
    s.Row(L"Start / places", StartPanel, 6, Plain, 126, 30, 0);
    s.Row(L"Start / more arrow", StartPanel, 3, Glyph, 24, 24, -1, -1, 1, 2, 3);
    s.Row(L"Start / logoff", StartPanel, 9, Glyph, 28, 28, -1, -1, 1, 2, 3);
    s.Row(L"Task group count", Taskband, 1, Plain, 126, 24, 0);
}

void DrawSheet(HDC dc, int width) {
    Sheet s(dc, width);
    switch (currentPage) {
    case 0:
        Buttons(s);
        break;
    case 1:
        Inputs(s);
        NamedPartRow(s, L"Spin / up", L"SPIN", 1, 18, 14, 4);
        NamedPartRow(s, L"Spin / down", L"SPIN", 2, 18, 14, 4);
        NamedPartRow(s, L"Spin / right", L"SPIN", 3, 18, 18, 4);
        NamedPartRow(s, L"Spin / left", L"SPIN", 4, 18, 18, 4);
        NamedPartRow(s, L"Pager / up", L"PAGE", 1, 28, 18, 4);
        NamedPartRow(s, L"Pager / down", L"PAGE", 2, 28, 18, 4);
        break;
    case 2:
        Tabs(s);
        break;
    case 3:
        Lists(s);
        break;
    case 4:
        Scrollbars(s);
        break;
    case 5:
        Sliders(s);
        break;
    case 6:
        Toolbars(s);
        break;
    case 7:
        CaptionButtons(s);
        break;
    case 8:
        Frames(s);
        break;
    case 9:
        Menus(s);
        break;
    case 10:
        Typography(s);
        break;
    case 11:
        Panels(s);
        break;
    case 12:
        ExplorerParts(s);
        break;
    case 13:
        ShellParts(s);
        break;
    }
}

void SelectPage(int index) {
    currentPage = (index + pageCount) % pageCount;
    TabCtrl_SetCurSel(tabWindow, currentPage);
    InvalidateRect(canvasWindow, NULL, FALSE);
    wchar_t text[256];
    wsprintfW(text, L"%d / %d   |   %d DPI   |   Ctrl+PgUp / Ctrl+PgDn: pages   |   All samples are static", currentPage + 1, pageCount, dpi);
    SetWindowTextW(footerWindow, text);
}

void Layout() {
    if (!tabWindow)
        return;
    RECT r = {};
    GetClientRect(mainWindow, &r);
    MoveWindow(titleWindow, Px(16), Px(10), r.right - Px(32), Px(23), TRUE);
    MoveWindow(tabWindow, Px(12), Px(41), r.right - Px(24), r.bottom - Px(71), TRUE);
    RECT inside = {};
    GetClientRect(tabWindow, &inside);
    TabCtrl_AdjustRect(tabWindow, FALSE, &inside);
    MapWindowPoints(tabWindow, mainWindow, reinterpret_cast<POINT *>(&inside), 2);
    InflateRect(&inside, -Px(3), -Px(3));
    SetWindowPos(canvasWindow, HWND_TOP, inside.left, inside.top, inside.right - inside.left, inside.bottom - inside.top, SWP_NOACTIVATE);
    MoveWindow(footerWindow, Px(16), r.bottom - Px(24), r.right - Px(32), Px(20), TRUE);
    InvalidateRect(canvasWindow, NULL, FALSE);
}

LRESULT CALLBACK CanvasProc(HWND h, UINT msg, WPARAM w, LPARAM l) {
    if (msg == WM_ERASEBKGND)
        return 1;
    if (msg == WM_PAINT) {
        PAINTSTRUCT paint = {};
        HDC target = BeginPaint(h, &paint);
        RECT r = {};
        GetClientRect(h, &r);
        if (r.right > 0 && r.bottom > 0) {
            HDC buffer = CreateCompatibleDC(target);
            HBITMAP bitmap = CreateCompatibleBitmap(target, r.right, r.bottom);
            HGDIOBJ previous = SelectObject(buffer, bitmap);
            HGDIOBJ previousFont = SelectObject(buffer, uiFont);
            FillRect(buffer, &r, GetSysColorBrush(COLOR_BTNFACE));
            DrawSheet(buffer, r.right);
            BitBlt(target, 0, 0, r.right, r.bottom, buffer, 0, 0, SRCCOPY);
            SelectObject(buffer, previousFont);
            SelectObject(buffer, previous);
            DeleteObject(bitmap);
            DeleteDC(buffer);
        }
        EndPaint(h, &paint);
        return 0;
    }
    return DefWindowProcW(h, msg, w, l);
}

LRESULT CALLBACK MainProc(HWND h, UINT msg, WPARAM w, LPARAM l) {
    switch (msg) {
    case WM_SIZE:
        if (w != SIZE_MINIMIZED)
            Layout();
        return 0;
    case WM_GETMINMAXINFO: {
        RECT r = MakeRect(0, 0, Px(948), Px(650));
        AdjustWindowRectEx(&r, WS_OVERLAPPEDWINDOW, FALSE, 0);
        MINMAXINFO *info = reinterpret_cast<MINMAXINFO *>(l);
        info->ptMinTrackSize.x = r.right - r.left;
        info->ptMinTrackSize.y = r.bottom - r.top;
        return 0;
    }
    case WM_NOTIFY:
        if (reinterpret_cast<NMHDR *>(l)->hwndFrom == tabWindow && reinterpret_cast<NMHDR *>(l)->code == TCN_SELCHANGE)
            SelectPage(TabCtrl_GetCurSel(tabWindow));
        return 0;
    case WM_CTLCOLORSTATIC:
        return ButtonFaceColors(reinterpret_cast<HDC>(w), reinterpret_cast<HWND>(l));
    case WM_THEMECHANGED:
        OpenThemes();
        InvalidateRect(canvasWindow, NULL, FALSE);
        return 0;
    case WM_DESTROY:
        PostQuitMessage(0);
        return 0;
    }
    return DefWindowProcW(h, msg, w, l);
}

int Run() {
    InitializeUI();
    WNDCLASSEXW wc = {};
    wc.cbSize = sizeof(wc);
    wc.hInstance = instance;
    wc.hCursor = LoadCursorW(NULL, IDC_ARROW);
    wc.hIcon = LoadIconW(NULL, IDI_APPLICATION);
    wc.hbrBackground = GetSysColorBrush(COLOR_BTNFACE);
    wc.lpfnWndProc = MainProc;
    wc.lpszClassName = L"XPControlsAtlas";
    if (!RegisterClassExW(&wc))
        Fatal(L"Register window");
    wc.lpfnWndProc = CanvasProc;
    wc.lpszClassName = L"XPControlsSheet";
    if (!RegisterClassExW(&wc))
        Fatal(L"Register sheet");
    DWORD style = WS_OVERLAPPEDWINDOW | WS_CLIPCHILDREN | WS_CLIPSIBLINGS;
    RECT r = MakeRect(0, 0, Px(960), Px(660));
    AdjustWindowRectEx(&r, style, FALSE, 0);
    int width = r.right - r.left, height = r.bottom - r.top;
    RECT work = {};
    SystemParametersInfoW(SPI_GETWORKAREA, 0, &work, 0);
    int x = Max(work.left, work.left + (work.right - work.left - width) / 2), y = Max(work.top, work.top + (work.bottom - work.top - height) / 2);
    mainWindow = CreateWindowExW(0, L"XPControlsAtlas", L"Windows XP - UI State Atlas v2.1 (C++)", style, x, y, width, height, NULL, NULL, instance, NULL);
    if (!mainWindow)
        Fatal(L"Create window");
    titleWindow = CreateWindowExW(0, L"STATIC", L"Windows XP UI state atlas", WS_CHILD | WS_VISIBLE, 0, 0, 0, 0, mainWindow, NULL, instance, NULL);
    SendMessageW(titleWindow, WM_SETFONT, reinterpret_cast<WPARAM>(headingFont), FALSE);
    tabWindow = CreateWindowExW(0, WC_TABCONTROLW, L"", WS_CHILD | WS_VISIBLE | WS_TABSTOP | WS_CLIPSIBLINGS, 0, 0, 0, 0, mainWindow, NULL, instance, NULL);
    SetUIFont(tabWindow);
    for (int i = 0; i < pageCount; ++i) {
        TCITEMW item = {};
        item.mask = TCIF_TEXT;
        item.pszText = const_cast<LPWSTR>(pageNames[i]);
        TabCtrl_InsertItem(tabWindow, i, &item);
    }
    canvasWindow = CreateWindowExW(0, L"XPControlsSheet", L"", WS_CHILD | WS_VISIBLE | WS_CLIPSIBLINGS, 0, 0, 0, 0, mainWindow, NULL, instance, NULL);
    footerWindow = CreateWindowExW(0, L"STATIC", L"", WS_CHILD | WS_VISIBLE, 0, 0, 0, 0, mainWindow, NULL, instance, NULL);
    SetUIFont(footerWindow);
    OpenThemes();
    CreatePanels();
    Layout();
    SelectPage(0);
    ShowWindow(mainWindow, SW_SHOWNORMAL);
    UpdateWindow(mainWindow);
    MSG msg = {};
    BOOL result;
    while ((result = GetMessageW(&msg, NULL, 0, 0)) > 0) {
        if (msg.message == WM_KEYDOWN) {
            if ((GetKeyState(VK_CONTROL) & 0x8000) && (msg.wParam == VK_NEXT || msg.wParam == VK_PRIOR)) {
                SelectPage(currentPage + (msg.wParam == VK_NEXT ? 1 : -1));
                continue;
            }
            if (msg.wParam >= VK_F1 && msg.wParam <= VK_F12) {
                SelectPage(static_cast<int>(msg.wParam - VK_F1));
                continue;
            }
        }
        TranslateMessage(&msg);
        DispatchMessageW(&msg);
    }
    for (HTHEME theme : themeHandles)
        if (theme)
            CloseThemeData(theme);
    DestroyWindow(hiddenHost);
    ReleaseUI();
    return result == -1 ? 1 : 0;
}
} // namespace

extern "C" void WinMainCRTStartup() {
    ExitProcess(static_cast<UINT>(Run()));
}
