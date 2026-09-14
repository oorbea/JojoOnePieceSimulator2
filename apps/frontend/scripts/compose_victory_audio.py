#!/usr/bin/env python3
"""Composes assets/audio/victory-full.wav from the owner-supplied one-shots
in C:\\Users\\meckp\\Desktop\\jojo_one_piece_simulator_sounds\\victory, laid
out on the exact beat map lib/outcome-cinematic.ts's
`cinematicTimeline('victory')` uses to drive the visual (see
ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md). See
audio_compose_lib.py for the mixing mechanics.

Beat map (ms, from lib/outcome-cinematic.ts):
  breath      0- 600  shimmer.wav             (aliento)
  dawn      600-1800  female_choral.wav        (amanecer, swell de coro)
  coronation 1800-3400 bell_epic_choir.wav     (campana + coro fortissimo)
  memory    3400-5400 arpegio.wav              (arpegio bajo los nombres)
  seal      5400-6600 final_chord_sustain.wav  (acorde final sostenido)
  epilogue  6600-7600 queue.wav                (cola)
Total 7600ms, matching CinematicKind 'victory' exactly.

tick.wav (the per-name "memory" beat sting) is deliberately NOT baked in
here - the actual number of winners varies per match, so it's played live,
once per name, by outcome-cinematic.tsx itself via the separate
name-tick.wav asset. See normalize_tick_sound() below for how that file was
produced from the same source folder.
"""
import os

from audio_compose_lib import compose, normalize_one_shot

SRC_DIR = r"C:\Users\meckp\Desktop\jojo_one_piece_simulator_sounds\victory"
AUDIO_DIR = os.path.join(os.path.dirname(__file__), "..", "assets", "audio")
OUT_PATH = os.path.join(AUDIO_DIR, "victory-full.wav")

LAYERS = [
    ("shimmer.wav", 0, 20.0),  # breath: was barely audible (-48 RMS)
    ("female_choral.wav", 600, 6.0),  # dawn: choir swell rising
    ("bell_epic_choir.wav", 1800, 8.0),  # coronation: the big hit, was quiet (-25 RMS)
    ("arpegio.wav", 3400, -1.0),  # memory: already strong, trim slightly
    ("final_chord_sustain.wav", 5400, 3.0),  # seal: the sealing chord
    ("queue.wav", 6900, -2.0),  # epilogue: the closing whoosh, already loud
]


def normalize_tick_sound():
    normalize_one_shot(
        os.path.join(SRC_DIR, "tick.wav"),
        os.path.join(AUDIO_DIR, "name-tick.wav"),
        target_peak_dbfs=-3.0,
    )


if __name__ == "__main__":
    compose(SRC_DIR, OUT_PATH, LAYERS, total_ms=7600, fade_out_ms=400)
    normalize_tick_sound()
