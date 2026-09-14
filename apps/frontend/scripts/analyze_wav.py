#!/usr/bin/env python3
"""Pure-stdlib WAV analyzer (no numpy/ffmpeg available on this machine):
duration, channels, sample rate/width, peak and RMS level in dBFS."""
import struct
import sys
import wave
import math
import os


def analyze(path):
    with wave.open(path, "rb") as w:
        nch = w.getnchannels()
        sw = w.getsampwidth()
        fr = w.getframerate()
        nframes = w.getnframes()
        raw = w.readframes(nframes)

    if sw != 2:
        return {"path": path, "error": f"unsupported sampwidth {sw}"}

    count = len(raw) // 2
    samples = struct.unpack("<%dh" % count, raw[: count * 2])

    peak = max(abs(s) for s in samples) if samples else 0
    sumsq = sum(s * s for s in samples)
    rms = math.sqrt(sumsq / len(samples)) if samples else 0

    def dbfs(x):
        if x <= 0:
            return float("-inf")
        return 20 * math.log10(x / 32768.0)

    duration = nframes / fr if fr else 0

    # Trailing silence: how much of the tail is below -50dBFS (rough measure
    # of unused headroom at the end, useful to know before trimming/extending).
    return {
        "path": os.path.basename(path),
        "channels": nch,
        "rate": fr,
        "duration_s": round(duration, 3),
        "peak_dbfs": round(dbfs(peak), 1),
        "rms_dbfs": round(dbfs(rms), 1),
    }


if __name__ == "__main__":
    for p in sys.argv[1:]:
        info = analyze(p)
        print(info)
