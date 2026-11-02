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

#include "audiomixer.h"
#include "platform.h"

const struct MusicFile
{
    const char *id;
    const char *file;
} musicFiles[3] = {
    { "intro", "/snd/introzik.ogg" },
    { "main1p", "/snd/frozen-mainzik-1p.ogg" },
    { "main2p", "/snd/frozen-mainzik-2p.ogg" },
};

AudioMixer *AudioMixer::ptrInstance = NULL;

AudioMixer *AudioMixer::Instance()
{
    if(ptrInstance == NULL)
        ptrInstance = new AudioMixer();
    return ptrInstance;
}

AudioMixer::AudioMixer()
{
    gameSettings = GameSettings::Instance();

    int freq = gameSettings->useClassicAudio() ? 22050 : MIX_DEFAULT_FREQUENCY;
    int sampl = gameSettings->useClassicAudio() ? 1024 : 4096;

    if (Mix_OpenAudio(freq, MIX_DEFAULT_FORMAT, 2, sampl) < 0) {
        SDL_LogError(1, "Could not open audio mixer! Music will be disabled. (%s)", Mix_GetError());
        mixerEnabled = false;
    }
}

Mix_Chunk* AudioMixer::GetSFX(const char *sfx){
    if(sfxFiles[sfx] == nullptr) {
        char rel[128];
        snprintf(rel, sizeof(rel), "/snd/%s.ogg", sfx);
        sfxFiles[sfx] = Mix_LoadWAV(ASSET(rel).c_str());
    }

    return sfxFiles[sfx];
}

AudioMixer::~AudioMixer(){
}

void AudioMixer::Dispose(){
    Mix_Quit();
    this->~AudioMixer();
}

void AudioMixer::PlayMusic(const char *track)
{
    if(mixerEnabled == false || !gameSettings->canPlayMusic() || haltedMixer == true) return;

#ifdef __WASM_PORT__
    // Blocking wait loops hang the browser — just halt immediately
    Mix_HaltMusic();
#else
    while (Mix_FadingMusic()) SDL_Delay(10);
    if (Mix_PlayingMusic()) Mix_FadeOutMusic(500);
    SDL_Delay(400);
    while (Mix_PlayingMusic()) SDL_Delay(10);
#endif
    
    std::string path;
    for (const MusicFile &musFile: musicFiles)
    {
        if (0 == strcmp(track, musFile.id))
        {
            path = ASSET(musFile.file);
            break;
        }
    }

    //if(curMusic) curMusic = NULL;
    curMusic = Mix_LoadMUS(path.c_str());
    if(curMusic) Mix_PlayMusic(curMusic, -1);
}

void AudioMixer::PlaySFX(const char *sfx)
{
    if(mixerEnabled == false || gameSettings->canPlaySFX() == false || haltedMixer == true) return;
    if (Mix_PlayChannel(-1, GetSFX(sfx), 0) < 0) SDL_LogError(1, "Could not play sound because of: %s", SDL_GetError());
}

void AudioMixer::PauseMusic(bool enable){
    if (enable == true) Mix_ResumeMusic();
    else Mix_PauseMusic();
}

void AudioMixer::MuteAll(bool enable){
    if(enable == true) haltedMixer = false;
    else {
        Mix_HaltMusic();
        Mix_HaltChannel(-1);
        haltedMixer = true;
    }
}