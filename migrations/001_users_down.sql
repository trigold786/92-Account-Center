-- +goose Up
-- This file contains DOWN migrations for the initial schema (001)
-- Run these to reverse the initial schema migration

-- Drop indexes first
DROP INDEX IF EXISTS idx_users_phone_number;
DROP INDEX IF EXISTS idx_users_account_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables
DROP TABLE IF EXISTS enterprises CASCADE;
DROP TABLE IF EXISTS users CASCADE;
