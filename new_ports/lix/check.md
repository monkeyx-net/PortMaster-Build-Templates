# check_port: lix

`check_port.sh lix --fix --fix-license` on 2026-11-12 23:54, commit 0448ab0.

**Result: ok, 8 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libs.x86_64/liballegro_main.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.x86_64/liballegro_color.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.x86_64/liballegro_memfile.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/liballegro_main.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/liballegro_color.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/liballegro_memfile.so.5.2 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- build.sh FILES copies libtheoradec.so.1 for x86_64 but libs.x86_64/ has none and only aarch64 loads it: move it to a [[ ${ARCH} == "aarch64" ]] && FILES+=( ) block

## Notes

- aarch64 loads from libs.aarch64: liballegro.so.5.2 liballegro_acodec.so.5.2 liballegro_audio.so.5.2 liballegro_font.so.5.2 liballegro_image.so.5.2 liballegro_primitives.so.5.2 liballegro_ttf.so.5.2 libenet.so libenet.so.7 libjpeg.so.8 libogg.so.0 libtheoradec.so.1 libvorbis.so.0 libvorbisfile.so.3
- x86_64 loads from libs.x86_64: liballegro.so.5.2 liballegro_acodec.so.5.2 liballegro_audio.so.5.2 liballegro_font.so.5.2 liballegro_image.so.5.2 liballegro_primitives.so.5.2 liballegro_ttf.so.5.2 libenet.so libenet.so.7 libjpeg.so.8 libogg.so.0 libvorbis.so.0 libvorbisfile.so.3
