-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    coins           BIGINT NOT NULL DEFAULT 1000,
    xp              BIGINT NOT NULL DEFAULT 0,
    rank_points     INT NOT NULL DEFAULT 0,
    rank_tier       VARCHAR(20) NOT NULL DEFAULT 'bronze',
    wins            INT NOT NULL DEFAULT 0,
    losses          INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_rank_points ON users(rank_points DESC);
CREATE INDEX idx_users_rank_tier ON users(rank_tier);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Player templates (for generation reference, not auction-specific)
CREATE TABLE players (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    position        VARCHAR(10) NOT NULL,
    secondary_position VARCHAR(10),
    club            VARCHAR(100) NOT NULL,
    nation          VARCHAR(100) NOT NULL,
    rating          INT NOT NULL CHECK (rating BETWEEN 60 AND 99),
    market_value    BIGINT NOT NULL,
    form            INT NOT NULL DEFAULT 75 CHECK (form BETWEEN 50 AND 99),
    age             INT NOT NULL DEFAULT 25 CHECK (age BETWEEN 17 AND 38),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_position ON players(position);
CREATE INDEX idx_players_rating ON players(rating DESC);

-- Auctions
CREATE TABLE auctions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status              VARCHAR(20) NOT NULL DEFAULT 'setup',
    difficulty          VARCHAR(20) NOT NULL,
    budget              BIGINT NOT NULL,
    player_pool_size    INT NOT NULL,
    ai_opponents        INT NOT NULL CHECK (ai_opponents BETWEEN 1 AND 5),
    seed                BIGINT NOT NULL,
    current_player_index INT NOT NULL DEFAULT 0,
    current_bid         BIGINT NOT NULL DEFAULT 0,
    highest_bidder_id   UUID,
    timer_seconds       INT NOT NULL DEFAULT 15,
    version             INT NOT NULL DEFAULT 0,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auctions_user_id ON auctions(user_id);
CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_created_at ON auctions(created_at DESC);

-- Auction participants (human + AI managers)
CREATE TABLE auction_participants (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id          UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id             UUID REFERENCES users(id),
    name                VARCHAR(100) NOT NULL,
    participant_type    VARCHAR(10) NOT NULL CHECK (participant_type IN ('human', 'ai')),
    personality         VARCHAR(30),
    remaining_budget    BIGINT NOT NULL,
    formation           VARCHAR(10) DEFAULT '4-3-3',
    has_passed          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auction_participants_auction_id ON auction_participants(auction_id);

-- Players generated for a specific auction
CREATE TABLE auction_players (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id      UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    player_index    INT NOT NULL,
    name            VARCHAR(100) NOT NULL,
    position        VARCHAR(10) NOT NULL,
    secondary_position VARCHAR(10),
    club            VARCHAR(100) NOT NULL,
    nation          VARCHAR(100) NOT NULL,
    rating          INT NOT NULL,
    market_value    BIGINT NOT NULL,
    form            INT NOT NULL,
    age             INT NOT NULL,
    sold_to_id      UUID REFERENCES auction_participants(id),
    sold_amount     BIGINT,
    sold_at         TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'available',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(auction_id, player_index)
);

CREATE INDEX idx_auction_players_auction_id ON auction_players(auction_id);
CREATE INDEX idx_auction_players_status ON auction_players(auction_id, status);

-- Bids
CREATE TABLE bids (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id      UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    auction_player_id UUID NOT NULL REFERENCES auction_players(id) ON DELETE CASCADE,
    participant_id  UUID NOT NULL REFERENCES auction_participants(id) ON DELETE CASCADE,
    amount          BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_auction_id ON bids(auction_id);
CREATE INDEX idx_bids_auction_player_id ON bids(auction_player_id);
CREATE INDEX idx_bids_participant_id ON bids(participant_id);

-- Teams (final squads after auction)
CREATE TABLE teams (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id      UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    participant_id  UUID NOT NULL REFERENCES auction_participants(id) ON DELETE CASCADE,
    formation       VARCHAR(10) NOT NULL DEFAULT '4-3-3',
    total_score     DECIMAL(8,2),
    score_breakdown JSONB,
    strengths       TEXT[],
    weaknesses      TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(auction_id, participant_id)
);

CREATE INDEX idx_teams_auction_id ON teams(auction_id);

-- Team players (squad members)
CREATE TABLE team_players (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id         UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    auction_player_id UUID NOT NULL REFERENCES auction_players(id),
    slot_type       VARCHAR(10) NOT NULL DEFAULT 'bench',
    slot_index      INT,
    purchase_price  BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_team_players_team_id ON team_players(team_id);

-- Auction results
CREATE TABLE results (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id      UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE UNIQUE,
    winner_id       UUID REFERENCES auction_participants(id),
    winner_user_id  UUID REFERENCES users(id),
    is_user_win     BOOLEAN NOT NULL DEFAULT FALSE,
    coins_reward    INT NOT NULL DEFAULT 0,
    xp_reward       INT NOT NULL DEFAULT 0,
    rank_points_reward INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_results_winner_user_id ON results(winner_user_id);
CREATE INDEX idx_results_created_at ON results(created_at DESC);

-- Leaderboards (materialized aggregation)
CREATE TABLE leaderboards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period          VARCHAR(20) NOT NULL DEFAULT 'alltime',
    total_score     BIGINT NOT NULL DEFAULT 0,
    wins            INT NOT NULL DEFAULT 0,
    losses          INT NOT NULL DEFAULT 0,
    auctions_played INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, period)
);

CREATE INDEX idx_leaderboards_period_score ON leaderboards(period, total_score DESC);
CREATE INDEX idx_leaderboards_user_id ON leaderboards(user_id);

-- Activity events (persisted feed)
CREATE TABLE activity_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id      UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    event_type      VARCHAR(30) NOT NULL,
    message         TEXT NOT NULL,
    participant_name VARCHAR(100),
    amount          BIGINT,
    player_name     VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_events_auction_id ON activity_events(auction_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS leaderboards;
DROP TABLE IF EXISTS results;
DROP TABLE IF EXISTS team_players;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS auction_players;
DROP TABLE IF EXISTS auction_participants;
DROP TABLE IF EXISTS auctions;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
