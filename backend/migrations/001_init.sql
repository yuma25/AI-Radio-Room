-- AI Radio Room - Issue #1: Database Initial Schema
-- Generated with assistance from GitHub Copilot CLI

-- 拡張機能の有効化 (UUID生成用)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 放送テーブル
CREATE TABLE IF NOT EXISTS broadcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

-- 台本・発言ログテーブル
CREATE TABLE IF NOT EXISTS transcripts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broadcast_id UUID NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    speaker_id TEXT NOT NULL, -- mc, guest-a, guest-b
    text TEXT NOT NULL,
    emotion TEXT NOT NULL DEFAULT 'normal', -- normal, excited, calm
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 検索を高速化するためのインデックス
CREATE INDEX IF NOT EXISTS idx_transcripts_broadcast_id ON transcripts(broadcast_id);
CREATE INDEX IF NOT EXISTS idx_transcripts_created_at ON transcripts(created_at);
