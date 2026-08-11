-- name: RecordGameResult :exec
INSERT INTO game_results (game_id, mode, winner, rounds_played, aborted)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (game_id) DO NOTHING;

-- name: RecordGameResultParticipant :exec
INSERT INTO game_result_participants (game_id, participant_id, user_id, display_name, team_id, is_bot)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (game_id, participant_id) DO NOTHING;

-- name: ListGameResultsByUser :many
SELECT gr.game_id, gr.mode, gr.winner, gr.rounds_played, gr.aborted, gr.recorded_at
FROM game_results gr
         JOIN game_result_participants grp ON grp.game_id = gr.game_id
WHERE grp.user_id = $1
ORDER BY gr.recorded_at DESC;
