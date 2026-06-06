\set ON_ERROR_STOP on

CREATE EXTENSION IF NOT EXISTS pgcrypto;

\i orm/example/sql/fdatabase/admin_demo.sql
\i orm/example/sql/fdatabase/admin_log_demo.sql
\i orm/example/sql/fdatabase/admin_role_demo.sql
\i orm/example/sql/fdatabase/admin_to_role_demo.sql
\i orm/example/sql/fdatabase/data_type_demo.sql
\i orm/example/sql/fdatabase/user_demo.sql

INSERT INTO public.user_demo (
    id,
    uid,
    username,
    password,
    nickname,
    status,
    tenant_id,
    created_at,
    updated_at
) VALUES
    (
        '182a65a0-ee20-4fe0-a0e8-ba30edcf402b',
        'user-1',
        'a',
        'password',
        'user-a',
        1,
        1,
        NOW(),
        NOW()
    ),
    (
        '2cc31ef9-7d6b-438b-874c-01d84a332b57',
        'user-2',
        'b',
        'password',
        'user-b',
        1,
        2,
        NOW(),
        NOW()
    );
