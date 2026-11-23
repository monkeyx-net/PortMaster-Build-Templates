# check_port: pikifen

`check_port.sh pikifen --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 5 for --fix, 4 for --fix-bloat, 1 for --fix-license; 1 error(s) left, 7 warning(s)**

## Errors

- gameinfo.xml is missing \<releasedate>

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libbz2.so.1.0 (x86_64) is needed by libs.x86_64/libfreetype.so.6 but not bundled or in cfw_provided: check ldd on devices
- libbrotlidec.so.1 (x86_64) is needed by libs.x86_64/libfreetype.so.6 but not bundled or in cfw_provided: check ldd on devices
- libs.aarch64/liballegro_memfile.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.x86_64/liballegro_memfile.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- build.sh FILES copies libtheoradec.so.1 for x86_64 but libs.x86_64/ has none and only aarch64 loads it: move it to a [[ ${ARCH} == "aarch64" ]] && FILES+=( ) block
- README.md doesn't follow the template (## Notes / ## Controls / ## Compile): missing ## Compile. https://portmaster.games/tools/markdown-generator/ builds one

## To fix with --fix

- Pikifen.sh:23: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (pikifen.aarch64 needs GLIBC_2.29)
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Launcher, ## Controls -> ## Notes, ## Controls
- shrink screenshot.png from 624 to 94 kB (16 colours)

## To fix with --fix-bloat

- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)

## To fix with --fix-license

- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: liballegro_acodec.so.5.2 liballegro_audio.so.5.2 liballegro_color.so.5.2 liballegro_font.so.5.2 liballegro_image.so.5.2 liballegro_main.so.5.2 liballegro_primitives.so.5.2 liballegro.so.5.2 liballegro_ttf.so.5.2 libfreetype.so.6 libjpeg.so.8 libogg.so.0 libpng16.so.16 libtheoradec.so.1 libvorbisfile.so.3 libvorbis.so.0
- x86_64 loads from libs.x86_64: liballegro_acodec.so.5.2 liballegro_audio.so.5.2 liballegro_color.so.5.2 liballegro_font.so.5.2 liballegro_image.so.5.2 liballegro_main.so.5.2 liballegro_primitives.so.5.2 liballegro.so.5.2 liballegro_ttf.so.5.2 libfreetype.so.6 libjpeg.so.8 libogg.so.0 libpng16.so.16 libvorbisfile.so.3 libvorbis.so.0
