/*
 * Frozen-Bubble SDL2 C++ Port
 * Copyright (c) 2000-2012 The Frozen-Bubble Team
 * Copyright (c) 2026 Huy Chau
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 */

#ifndef TTFTEXT_H
#define TTFTEXT_H

#include <SDL2/SDL.h>
#include <SDL2/SDL_ttf.h>
#include <iostream>
#include <string>

// TTF_WRAPPED_ALIGN_* and TTF_SetFontWrappedAlign were added in SDL2_ttf 2.20.
// Provide no-op fallbacks for older distro packages (e.g. Ubuntu 22.04 ships 2.0.18).
#if !SDL_TTF_VERSION_ATLEAST(2, 20, 0)
#define TTF_WRAPPED_ALIGN_LEFT   0
#define TTF_WRAPPED_ALIGN_CENTER 1
#define TTF_WRAPPED_ALIGN_RIGHT  2
inline void TTF_SetFontWrappedAlign(TTF_Font*, int) {}
#endif

#define WINDOW_W 640
#define WINDOW_H 480
#define SURF_FORMAT SDL_PIXELFORMAT_ARGB8888

class TTFText final
{
public:
    void LoadFont(const char *path, int size);
    void LoadFont(TTF_Font *fnt);

    void UpdateText(const SDL_Renderer *rend, const char *txt, int wrapLength);
    void UpdateAlignment(int align);
    void UpdateColor(SDL_Color fg, SDL_Color bg);
    void UpdateStyle(int size, int style);
    void UpdateStyle(int style);
    void UpdatePosition(SDL_Point xy);

    const char *Text() { return curText; };
    SDL_Rect *Coords() { return &coords; };
    SDL_Texture *Texture() { return outTexture; };
    
    TTFText();
    ~TTFText();
    TTFText(const TTFText&) : curText(nullptr), coords{}, forecolor{}, backcolor{}, textFont(nullptr), ownsFont(false), outTexture(nullptr) {}
    TTFText& operator=(const TTFText&) { return *this; }
private:
    char *curText;

    SDL_Rect coords;
    SDL_Color forecolor, backcolor;

    TTF_Font *textFont = nullptr;
    bool ownsFont = false;
    SDL_Texture *outTexture = nullptr;
};

#endif //TTFTEXT_H