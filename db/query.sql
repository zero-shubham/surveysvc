-- name: CreateAnswer :one
INSERT INTO
  answers (
    selected_option,
    answer_text,
    user_id,
    question_id,
    question_set_id
  )
VALUES
  ($1, $2, $3, $4, $5)
  RETURNING *;

-- name: CreateQuestionMapping :one
INSERT INTO question_mappings (question_id, campaign_id, org_id)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetAnswersByQuestionID :many
SELECT id, selected_option, answer_text, user_id, question_id, question_set_id, created_at, updated_at
FROM answers WHERE question_id = $1;

-- name: GetQuestionMappingsByCampaignID :many
SELECT * FROM question_mappings WHERE campaign_id = $1;

-- name: UpdateQuestionMappingsByID :one
UPDATE question_mappings
SET campaign_id = $2, question_id = $3, org_id = $4, updated_at = NOW()
WHERE id = $1 RETURNING *;

