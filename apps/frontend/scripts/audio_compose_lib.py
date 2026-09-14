#!/usr/bin/env python3
"""Shared pure-stdlib WAV mixing helpers for compose_defeat_audio.py and
compose_victory_audio.py - no ffmpeg/numpy available on this machine, so
resampling is simple linear interpolation (fine for short SFX, not meant
for music mastering) and levelling is peak-based rather than true LUFS.
"""
import math
import os
import struct
import wave


def read_wav_stereo_f(path):
    """Returns (left, right, sample_rate) as float samples in [-1, 1]."""
    with wave.open(path, "rb") as w:
        nch = w.getnchannels()
        sw = w.getsampwidth()
        fr = w.getframerate()
        nframes = w.getnframes()
        raw = w.readframes(nframes)
    if sw != 2:
        raise ValueError(f"{path}: unsupported sample width {sw}")
    count = len(raw) // 2
    samples = struct.unpack("<%dh" % count, raw[: count * 2])
    if nch == 1:
        left = [s / 32768.0 for s in samples]
        right = left
    elif nch == 2:
        left = [s / 32768.0 for s in samples[0::2]]
        right = [s / 32768.0 for s in samples[1::2]]
    else:
        raise ValueError(f"{path}: unsupported channel count {nch}")
    return left, right, fr


def resample_linear(chan, src_rate, dst_rate):
    if src_rate == dst_rate:
        return chan
    ratio = dst_rate / src_rate
    dst_len = max(1, round(len(chan) * ratio))
    out = [0.0] * dst_len
    for i in range(dst_len):
        pos = i / ratio
        i0 = int(pos)
        i1 = min(i0 + 1, len(chan) - 1)
        frac = pos - i0
        out[i] = chan[i0] * (1 - frac) + chan[i1] * frac
    return out


def gain(chan, db):
    factor = 10 ** (db / 20)
    return [s * factor for s in chan]


def compose(src_dir, out_path, layers, total_ms, fade_out_ms, target_rate=44100, target_peak_dbfs=-1.0):
    """Mixes `layers` (filename, start_ms, gain_db) from src_dir into a single
    stereo WAV at out_path, exactly total_ms long, with a linear fade-out over
    the last fade_out_ms (meant to land on the matching visual cross-fade -
    see outcome-cinematic.tsx's epilogue timing), peak-normalized to
    target_peak_dbfs."""
    total_samples = round(total_ms / 1000 * target_rate)
    mix_l = [0.0] * total_samples
    mix_r = [0.0] * total_samples

    for filename, start_ms, gain_db in layers:
        path = os.path.join(str(src_dir), filename)
        left, right, src_rate = read_wav_stereo_f(path)
        left = gain(resample_linear(left, src_rate, target_rate), gain_db)
        right = gain(resample_linear(right, src_rate, target_rate), gain_db)
        start_sample = round(start_ms / 1000 * target_rate)
        for i, (sl, sr) in enumerate(zip(left, right)):
            idx = start_sample + i
            if idx >= total_samples:
                break
            mix_l[idx] += sl
            mix_r[idx] += sr
        print(f"mixed {filename}: {len(left)} samples @ {start_ms}ms, gain {gain_db:+.1f}dB")

    fade_samples = round(fade_out_ms / 1000 * target_rate)
    for i in range(fade_samples):
        idx = total_samples - fade_samples + i
        mult = 1.0 - (i / fade_samples)
        mix_l[idx] *= mult
        mix_r[idx] *= mult

    peak = max(max(abs(s) for s in mix_l), max(abs(s) for s in mix_r))
    target_peak_lin = 10 ** (target_peak_dbfs / 20)
    norm = (target_peak_lin / peak) if peak > 0 else 1.0
    print(f"pre-normalize peak: {20 * math.log10(peak):.1f}dBFS, applying {20 * math.log10(norm):+.1f}dB")

    def to_int16(s):
        v = max(-1.0, min(1.0, s * norm))
        return int(round(v * 32767))

    out_l = [to_int16(s) for s in mix_l]
    out_r = [to_int16(s) for s in mix_r]

    interleaved = [0] * (total_samples * 2)
    interleaved[0::2] = out_l
    interleaved[1::2] = out_r

    with wave.open(out_path, "wb") as w:
        w.setnchannels(2)
        w.setsampwidth(2)
        w.setframerate(target_rate)
        w.writeframes(struct.pack("<%dh" % len(interleaved), *interleaved))

    print(f"wrote {out_path} ({total_ms}ms, {target_rate}Hz stereo)")


def produce_one_shot(src_path, out_path, max_duration_ms, fade_out_ms, target_peak_dbfs=-1.0, target_rate=44100):
    """Resamples, peak-normalizes AND trims a one-shot to at most
    max_duration_ms with a fade-out over its last fade_out_ms - for a sound
    played inside a fixed-length overlay (e.g. the 1.5s round flash) whose
    caller stops playback outright the moment the overlay's own timer ends;
    without this, a source clip longer than that window gets cut off
    mid-ring instead of fading out cleanly beforehand."""
    left, right, src_rate = read_wav_stereo_f(src_path)
    left = resample_linear(left, src_rate, target_rate)
    right = resample_linear(right, src_rate, target_rate)

    max_samples = round(max_duration_ms / 1000 * target_rate)
    left = left[:max_samples]
    right = right[:max_samples]

    fade_samples = min(round(fade_out_ms / 1000 * target_rate), len(left))
    for i in range(fade_samples):
        idx = len(left) - fade_samples + i
        mult = 1.0 - (i / fade_samples)
        left[idx] *= mult
        right[idx] *= mult

    peak = max(max(abs(s) for s in left), max(abs(s) for s in right))
    target_peak_lin = 10 ** (target_peak_dbfs / 20)
    norm = (target_peak_lin / peak) if peak > 0 else 1.0
    print(
        f"{os.path.basename(src_path)} trimmed to {len(left) / target_rate * 1000:.0f}ms, "
        f"pre-normalize peak: {20 * math.log10(peak):.1f}dBFS, applying {20 * math.log10(norm):+.1f}dB"
    )

    def to_int16(s):
        v = max(-1.0, min(1.0, s * norm))
        return int(round(v * 32767))

    interleaved = [0] * (len(left) * 2)
    interleaved[0::2] = [to_int16(s) for s in left]
    interleaved[1::2] = [to_int16(s) for s in right]

    with wave.open(out_path, "wb") as w:
        w.setnchannels(2)
        w.setsampwidth(2)
        w.setframerate(target_rate)
        w.writeframes(struct.pack("<%dh" % len(interleaved), *interleaved))
    print(f"wrote {out_path} ({len(left) / target_rate * 1000:.0f}ms, {target_rate}Hz stereo)")


def normalize_one_shot(src_path, out_path, target_peak_dbfs=-3.0, target_rate=44100):
    """Resamples a single one-shot to target_rate and peak-normalizes it -
    for a sound that's played as a live event (e.g. the per-name tick),
    not mixed into a composite track."""
    left, right, src_rate = read_wav_stereo_f(src_path)
    left = resample_linear(left, src_rate, target_rate)
    right = resample_linear(right, src_rate, target_rate)
    peak = max(max(abs(s) for s in left), max(abs(s) for s in right))
    target_peak_lin = 10 ** (target_peak_dbfs / 20)
    norm = (target_peak_lin / peak) if peak > 0 else 1.0
    print(f"{os.path.basename(src_path)} pre-normalize peak: {20 * math.log10(peak):.1f}dBFS, applying {20 * math.log10(norm):+.1f}dB")

    def to_int16(s):
        v = max(-1.0, min(1.0, s * norm))
        return int(round(v * 32767))

    interleaved = [0] * (len(left) * 2)
    interleaved[0::2] = [to_int16(s) for s in left]
    interleaved[1::2] = [to_int16(s) for s in right]

    with wave.open(out_path, "wb") as w:
        w.setnchannels(2)
        w.setsampwidth(2)
        w.setframerate(target_rate)
        w.writeframes(struct.pack("<%dh" % len(interleaved), *interleaved))
    print(f"wrote {out_path} ({len(left) / target_rate * 1000:.0f}ms, {target_rate}Hz stereo)")
