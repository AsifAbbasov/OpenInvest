-- Roll back AUTH-ADV-01 only in a disposable environment before accepted use.
BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';

DROP INDEX audit.auth_security_event_deduplications_uidx;
DROP TABLE audit.auth_security_event_deduplications;

COMMIT;
