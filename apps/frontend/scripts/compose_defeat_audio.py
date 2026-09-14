#!/usr/bin/env python3
"""Composes assets/audio/defeat-full.wav from the owner-supplied one-shots in
C:\\Users\\meckp\\Desktop\\jojo_one_piece_simulator_sounds\\defeat, laid out on
the exact beat map lib/outcome-cinematic.ts's `cinematicTimeline('defeat')`
uses to drive the visual (see ObsidianVault/game-victory-defeat-cinematic-
2026-09-14.md). See audio_compose_lib.py for the mixing mechanics.

Beat map (ms, from lib/outcome-cinematic.ts):
  impact    0- 400  thud.wav           (golpe seco)
  fracture  400-1200 crack.wav + scream.wav (crack + grito)
  ink      1200-2200 severe_blow.wav   (golpe grave)
  verdict  2200-4000 horn_of_doom.wav  (acorde de sentencia)
  silence  4000-5600 deep_drone.wav    (drone, casi silencio)
  epilogue 5600-7200 queue.wav         (cola)
Total 7200ms, matching CinematicKind 'defeat' exactly - the visual epilogue
fade (timeline.totalMs - 400 .. totalMs) and this file's own final 400ms
fade-out are timed to land together.
"""
import os

from audio_compose_lib import compose

SRC_DIR = r"C:\Users\meckp\Desktop\jojo_one_piece_simulator_sounds\defeat"
OUT_PATH = os.path.join(os.path.dirname(__file__), "..", "assets", "audio", "defeat-full.wav")

# (filename, start_ms, gain_db) - gain_db balances relative loudness between
# very differently recorded source files (some near-clipping, some barely
# above the noise floor); the final pass peak-normalizes the whole mix, so
# these only need to be relatively right, not absolutely safe.
LAYERS = [
    ("thud.wav", 0, 6.0),  # impact: was quiet (-20 RMS), wants a punchy hit
    ("crack.wav", 400, -1.0),  # fracture: already loud/peaky, trim slightly
    ("scream.wav", 500, 12.0),  # fracture/ink: was very quiet (-31 RMS)
    ("severe_blow.wav", 1200, -4.0),  # ink: source peaked at 0.0dBFS already
    ("horn_of_doom.wav", 2200, 2.0),  # verdict: the big sting, slight boost
    ("deep_drone.wav", 4000, 14.0),  # silence: was barely audible (-36 RMS)
    ("queue.wav", 5700, -2.0),  # epilogue: the closing whoosh, already loud
]

if __name__ == "__main__":
    compose(SRC_DIR, OUT_PATH, LAYERS, total_ms=7200, fade_out_ms=400)
