-- +goose Up
INSERT INTO projects (id, name, created_at, updated_at)
VALUES
    (
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Demo Project',
        '2025-01-01T00:00:00Z',
        '2025-01-01T00:00:00Z'
    );

INSERT INTO
    tasks (
        id,
        project_id,
        title,
        description,
        status,
        created_at,
        updated_at
    )
VALUES
    (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Morning routine',
        'A basic overview',
        'IN_PROGRESS',
        '2025-01-01T01:00:00Z',
        '2025-01-01T02:00:00Z'
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Eat food',
        'descriptions of eating',
        'TODO',
        '2025-01-01T01:30:00Z',
        '2025-01-01T01:30:00Z'
    );

-- +goose Down
DELETE FROM
    tasks
WHERE
    id IN (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'cccccccc-cccc-cccc-cccc-cccccccccccc'
    );

DELETE FROM
    projects
WHERE
    id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';