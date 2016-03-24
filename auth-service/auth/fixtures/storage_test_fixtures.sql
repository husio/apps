

INSERT INTO accounts
    (id, login, password_hash, role, valid_till, created_at)
VALUES
    (1, 'bob@example.com', 'xxx', 'admin', now() + interval '90 days', now()),
    (2, 'rob@example.com', 'xxx', 'user', now() + interval '10 days', now() - interval '20 days'),
    (3, 'dick@example.com', 'xxx', 'user', now(), now() - interval '90 days'),
    (4, 'email-service', 'xxx', 'service', now() + interval '10 years', now() - interval '1 day');

---

SELECT setval('accounts_id_seq', 100);
