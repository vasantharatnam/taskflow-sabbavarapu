INSERT INTO users (id, name, email, password_hash)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Test User',
    'test@example.com',
    '$2a$12$AAXF4MejDyJoNBxEWbkggO5dlJnv8.U/WpJIpgOXzWQ6mlYuDrPye'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'Seed Project',
    'Initial seeded project',
    '11111111-1111-1111-1111-111111111111'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, creator_id, due_date)
VALUES
(
    '33333333-3333-3333-3333-333333333331',
    'Setup database',
    'Create initial schema',
    'todo',
    'high',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-20'
),
(
    '33333333-3333-3333-3333-333333333332',
    'Build auth API',
    'Implement JWT login and register',
    'in_progress',
    'medium',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-21'
),
(
    '33333333-3333-3333-3333-333333333333',
    'Write README',
    'Document setup and architecture',
    'done',
    'low',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-22'
) ON CONFLICT (id) DO NOTHING;
