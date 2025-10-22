-- +goose Up
INSERT INTO
    todos (id, title, completed, created_at)
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'Read the spec',
        0,
        '2025-01-01T00:00:00Z'
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'Implement handlers',
        1,
        '2025-01-02T00:00:00Z'
    );

-- +goose Down
DELETE FROM
    todos
WHERE
    id IN (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222'
    );