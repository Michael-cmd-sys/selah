-- name: CreateGroup :one
INSERT INTO groups (name, description, creator_id, is_private, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: ListUserGroups :many
SELECT g.* FROM groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE gm.user_id = $1
ORDER BY gm.joined_at DESC;

-- name: JoinGroup :exec
INSERT INTO group_members (group_id, user_id, role)
VALUES ($1, $2, 'member')
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: LeaveGroup :exec
DELETE FROM group_members WHERE group_id = $1 AND user_id = $2;

-- name: IsGroupMember :one
SELECT EXISTS(
    SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2
) AS is_member;

-- name: CreateChallenge :one
INSERT INTO challenges (group_id, creator_id, title, description, type, target, unit, start_date, end_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListActiveChallenges :many
SELECT c.*, cp.progress AS my_progress
FROM challenges c
LEFT JOIN challenge_participants cp ON cp.challenge_id = c.id AND cp.user_id = $1
WHERE c.end_date >= CURRENT_DATE
  AND (c.group_id IS NULL OR c.group_id IN (
      SELECT group_id FROM group_members WHERE user_id = $1
  ))
ORDER BY c.start_date DESC
LIMIT 20;

-- name: JoinChallenge :exec
INSERT INTO challenge_participants (challenge_id, user_id)
VALUES ($1, $2)
ON CONFLICT (challenge_id, user_id) DO NOTHING;

-- name: UpdateChallengeProgress :one
INSERT INTO challenge_participants (challenge_id, user_id, progress, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (challenge_id, user_id) DO UPDATE SET
    progress   = EXCLUDED.progress,
    updated_at = NOW()
RETURNING *;

-- name: CreatePost :one
INSERT INTO posts (user_id, group_id, post_type, content, verse_ref, verse_text)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPostByID :one
SELECT p.*, u.name AS author_name, u.avatar_url AS author_avatar
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.id = $1;

-- name: ListFeedPosts :many
SELECT p.*, u.name AS author_name, u.avatar_url AS author_avatar,
    (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.reaction_type = 'amen')    AS amen_count,
    (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.reaction_type = 'praying') AS praying_count,
    (SELECT COUNT(*) FROM post_comments pc WHERE pc.post_id = p.id) AS comment_count
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE ($1::uuid IS NULL OR p.group_id = $1)
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1 AND user_id = $2;

-- name: AddReaction :exec
INSERT INTO post_reactions (post_id, user_id, reaction_type)
VALUES ($1, $2, $3)
ON CONFLICT (post_id, user_id, reaction_type) DO NOTHING;

-- name: RemoveReaction :exec
DELETE FROM post_reactions
WHERE post_id = $1 AND user_id = $2 AND reaction_type = $3;

-- name: GetReactionsByPost :many
SELECT reaction_type, COUNT(*) AS count
FROM post_reactions
WHERE post_id = $1
GROUP BY reaction_type;

-- name: CreateComment :one
INSERT INTO post_comments (post_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListComments :many
SELECT pc.*, u.name AS author_name, u.avatar_url AS author_avatar
FROM post_comments pc
JOIN users u ON u.id = pc.user_id
WHERE pc.post_id = $1
ORDER BY pc.created_at ASC;

-- name: DeleteComment :exec
DELETE FROM post_comments WHERE id = $1 AND user_id = $2;
