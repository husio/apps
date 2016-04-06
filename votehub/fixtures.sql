BEGIN;

INSERT INTO accounts VALUES
    (1, 'baldar', 'github', now()),
    (2, 'Freya', 'github', now()),
    (3, 'Frigg', 'github', now()),
    (4, 'odin', 'github', now());
    (4, 'odin', 'github', now());

INSERT INTO counters (owner_id, created, url, description) VALUES
    (1, now(), 'https://github.com/optiopay/kafka/issues/49', 'Issue 49: Deadlock where Broker.mu is never released'),
    (2, now(), 'https://github.com/optiopay/kafka/issues/39', 'Issue 39: Document best practices for high level consumer functionality'),

    (3, now(), 'https://github.com/optiopay/kafka/issues/47', 'Issue 47:  Using proto.RequiredAcksNone causes panic #47 '),
    (4, now(), 'https://github.com/optiopay/kafka/issues/38', 'Issue 38: TravisCI test runs unstable'),
    (1, now(), 'https://github.com/optiopay/kafka/issues/27', 'Issue 27: Consumer does not transparently fail over when the leader changes'),
    (2, now(), 'https://github.com/optiopay/kafka/issues/26', 'Issue 26: Abstract away connection error handling');

INSERT INTO votes (counter_id, account_id, created) VALUES
    (1, 1, now()),
    (1, 2, now()),
    (1, 3, now()),
    (1, 4, now()),
    (2, 1, now()),
    (2, 2, now()),
    (2, 3, now()),
    (3, 1, now()),
    (3, 3, now()),
    (3, 4, now()),
    (4, 2, now()),
    (4, 3, now()),
    (4, 4, now());

COMMIT;
