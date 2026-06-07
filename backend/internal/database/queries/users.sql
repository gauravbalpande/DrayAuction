-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserProgression :exec
UPDATE users
SET coins = coins + $2, xp = xp + $3, rank_points = rank_points + $4,
    wins = wins + $5, losses = losses + $6, rank_tier = $7, updated_at = NOW()
WHERE id = $1;

-- name: CreateAuction :one
INSERT INTO auctions (user_id, difficulty, budget, player_pool_size, ai_opponents, seed)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAuctionByID :one
SELECT * FROM auctions WHERE id = $1;

-- name: UpdateAuctionStatus :exec
UPDATE auctions SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: CreateBid :one
INSERT INTO bids (auction_id, auction_player_id, participant_id, amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLeaderboard :many
SELECT l.*, u.username, u.rank_tier
FROM leaderboards l
JOIN users u ON u.id = l.user_id
WHERE l.period = $1
ORDER BY l.total_score DESC
LIMIT $2;
