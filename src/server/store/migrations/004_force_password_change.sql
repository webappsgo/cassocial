-- Migration 004: add force_password_change column to users
-- Allows admins to require a user to change their password on next login.
ALTER TABLE users ADD COLUMN force_password_change BOOLEAN NOT NULL DEFAULT 0;
