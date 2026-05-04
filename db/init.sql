DO $$ 
BEGIN 
  -- Optionally drop existing tables if running iteratively (Caution: for development only)
  DROP TABLE IF EXISTS permissions, events, teams, organizations, user_schedules, schedules, users CASCADE;
END $$;

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  is_admin BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (organization_id, name)
);

CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(255) NOT NULL,
  start_date_time TIMESTAMP NOT NULL,
  end_date_time TIMESTAMP,
  location VARCHAR(255),
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  created_by_id UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  resource_type VARCHAR(50) NOT NULL,
  resource_id UUID NOT NULL,
  permission VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, resource_type, resource_id, permission)
);

CREATE INDEX IF NOT EXISTS idx_events_start_date_time ON events(start_date_time);
CREATE INDEX IF NOT EXISTS idx_events_team_id ON events(team_id);
CREATE INDEX IF NOT EXISTS idx_events_created_by_id ON events(created_by_id);

-- Seed default admin user (id: 00000000-0000-0000-0000-000000000001)
-- Username: admin, Password: admin123
INSERT INTO users (id, username, password_hash, is_admin)
VALUES ('00000000-0000-0000-0000-000000000001'::uuid, 'admin', '$2a$10$yKRN3CG0VMgdJZB6b//7ieP8yC/TtpGigdtAilNw1.ZZ8IXZzgVgi', TRUE)
ON CONFLICT (username) DO NOTHING;

-- Seed organization
INSERT INTO organizations (id, name)
VALUES ('11111111-1111-1111-1111-111111111111'::uuid, 'Shakopee Wrestling')
ON CONFLICT (name) DO NOTHING;

-- Seed teams
INSERT INTO teams (id, organization_id, name)
VALUES 
  ('22222222-2222-2222-2222-222222222221'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Boy''s Varsity'),
  ('22222222-2222-2222-2222-222222222222'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Boys JV'),
  ('22222222-2222-2222-2222-222222222223'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Girls'),
  ('22222222-2222-2222-2222-222222222224'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Ninth Grade')
ON CONFLICT (organization_id, name) DO NOTHING;

-- Admins is_admin so they have access everywhere, no need for specific permission row.

-- Seed sample events into Boy's Varsity
INSERT INTO events (title, start_date_time, end_date_time, location, team_id, created_by_id)
VALUES
  ('Team Standup', '2026-04-06 09:00:00', '2026-04-06 09:30:00', 'Conference Room A', '22222222-2222-2222-2222-222222222221'::uuid, '00000000-0000-0000-0000-000000000001'::uuid),
  ('Design Review', '2026-04-06 11:00:00', '2026-04-06 12:00:00', 'Design Studio', '22222222-2222-2222-2222-222222222221'::uuid, '00000000-0000-0000-0000-000000000001'::uuid),
  ('Sprint Planning', '2026-04-07 10:00:00', '2026-04-07 11:30:00', 'War Room', '22222222-2222-2222-2222-222222222221'::uuid, '00000000-0000-0000-0000-000000000001'::uuid);
