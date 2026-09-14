#!/usr/bin/env python3
"""Produces assets/audio/round-win.wav and round-lose.wav from the owner-
supplied one-shots in
C:\\Users\\meckp\\Desktop\\jojo_one_piece_simulator_sounds\\rounds, for the
1.5s round flash (see round-flash.tsx, cinematicTimeline('roundWin'/
'roundLose') in lib/outcome-cinematic.ts).

Unlike the defeat/victory composites, these are single one-shots, not a
multi-layer mix - but they still need trimming: round-flash.tsx calls
sound.stop() the instant its own 1500ms timer ends, and both source clips
run ~2.2-2.4s, well past that. produce_one_shot() trims each to
MAX_DURATION_MS with a FADE_OUT_MS fade, timed to land inside the flash's
own visual fade-out (round-flash.tsx fades its text out over the last 350ms
of the 1500ms window) so the cut is never audible.
"""
import os

from audio_compose_lib import produce_one_shot

SRC_DIR = r"C:\Users\meckp\Desktop\jojo_one_piece_simulator_sounds\rounds"
OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "assets", "audio")

MAX_DURATION_MS = 1450
FADE_OUT_MS = 350

if __name__ == "__main__":
    produce_one_shot(
        os.path.join(SRC_DIR, "glissando_up.wav"),
        os.path.join(OUT_DIR, "round-win.wav"),
        max_duration_ms=MAX_DURATION_MS,
        fade_out_ms=FADE_OUT_MS,
        target_peak_dbfs=-1.0,
    )
    produce_one_shot(
        os.path.join(SRC_DIR, "rock_crack.wav"),
        os.path.join(OUT_DIR, "round-lose.wav"),
        max_duration_ms=MAX_DURATION_MS,
        fade_out_ms=FADE_OUT_MS,
        target_peak_dbfs=-1.0,  # source clipped at 0.0dBFS - needs headroom
    )
