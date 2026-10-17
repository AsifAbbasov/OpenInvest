"""Synthetic G1/A3 feasibility, not a service or a production cryptographic profile.
Run: python experiments/privacy-feasibility/admission_test.py -v
Only temporary SQLite databases and generated test signing keys are used.
"""
import hashlib
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import sys
import tempfile
import unittest
from contextlib import contextmanager
from concurrent.futures import ThreadPoolExecutor

from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey
from cryptography.hazmat.primitives.serialization import Encoding, PrivateFormat, NoEncryption

GRACE = 30 * 86400


def canonical(value):
    # Restricted ASCII/string/integer profile; NOT an implementation of all RFC 8785.
    def check(v):
        if isinstance(v, str):
            if not v.isascii():
                raise ValueError("non-ASCII profile value")
        elif type(v) is int:
            if not 0 <= v <= 2**53 - 1:
                raise ValueError("integer range")
        elif isinstance(v, dict):
            for k, item in v.items():
                if not isinstance(k, str):
                    raise ValueError("object key")
                check(k)
                check(item)
        else:
            raise ValueError("unsupported profile type")
    check(value)
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode()


def strict_decode(raw):
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result:
                raise ValueError("duplicate field")
            result[key] = value
        return result
    value = json.loads(raw, object_pairs_hook=pairs)
    original = raw.encode("utf-8") if isinstance(raw, str) else raw
    if not isinstance(original, bytes) or original != canonical(value):
        raise ValueError("non-canonical JSON bytes")
    return value


class SyntheticCheckpointDB:
    """Separate committed-read SQLite authority; not a distributed attestation service."""
    def __init__(self, root):
        self.path = Path(root) / "synthetic-primary.sqlite"
        with self.connect() as db:
            db.execute("CREATE TABLE IF NOT EXISTS pending(intent TEXT PRIMARY KEY, evidence BLOB NOT NULL)")

    @contextmanager
    def connect(self):
        db = sqlite3.connect(self.path, timeout=5, isolation_level=None)
        db.execute("PRAGMA synchronous=FULL")
        try:
            yield db
        finally:
            db.close()

    def proposed(self, journal, intent, generation=1, db_version=1):
        prepare = journal.operation_receipt(intent, "PREPARE", generation)
        return dict(v=1, domain="oi.test.db-checkpoint", intent=intent,
                    generation=generation, prepare_seq=prepare[0], prepare_root=prepare[2],
                    db_version=db_version, phase="PENDING")

    def committed(self, intent):
        with self.connect() as db:
            row = db.execute("SELECT evidence FROM pending WHERE intent=?", (intent,)).fetchone()
        return strict_decode(row[0]) if row else None

    def commit_pending(self, journal, intent, generation=1, db_version=1):
        evidence = self.proposed(journal, intent, generation, db_version)
        if type(db_version) is not int or db_version < 1:
            raise ValueError("checkpoint version")
        with self.connect() as db:
            db.execute("BEGIN IMMEDIATE")
            row = db.execute("SELECT evidence FROM pending WHERE intent=?", (intent,)).fetchone()
            old = strict_decode(row[0]) if row else None
            if old and (old["generation"] != generation or old["db_version"] > db_version):
                raise ValueError("checkpoint regression")
            if old and old["db_version"] == db_version and old != evidence:
                raise ValueError("checkpoint conflict")
            db.execute("INSERT INTO pending VALUES(?,?) ON CONFLICT(intent) DO UPDATE SET evidence=excluded.evidence",
                       (intent, canonical(evidence)))
            db.commit()
        return evidence

    @staticmethod
    def validate_binding(evidence, prepare, intent, generation):
        fields = {"v", "domain", "intent", "generation", "prepare_seq", "prepare_root", "db_version", "phase"}
        if not isinstance(evidence, dict) or set(evidence) != fields:
            raise ValueError("missing or ambiguous checkpoint")
        canonical(evidence)
        expected = dict(v=1, domain="oi.test.db-checkpoint", intent=intent, generation=generation,
                        prepare_seq=prepare[0], prepare_root=prepare[2], db_version=evidence["db_version"], phase="PENDING")
        if evidence != expected or type(evidence["db_version"]) is not int or evidence["db_version"] < 1:
            raise ValueError("checkpoint binding")
    def validate(self, evidence, prepare, intent, generation):
        self.validate_binding(evidence, prepare, intent, generation)
        if self.committed(intent) != evidence:
            raise ValueError("checkpoint not committed or stale")


def signing_bytes(seq, payload, root):
    strict_decode(payload)  # digest and signature must bind the exact same canonical bytes
    return canonical(dict(domain="oi.test.checkpoint", epoch=1, seq=seq, root=root,
                          payload_digest=hashlib.sha256(payload).hexdigest()))


class Journal:
    def __init__(self, directory):
        self.root = Path(directory)
        self.path = self.root / "independent-journal.sqlite"
        self.primary = SyntheticCheckpointDB(directory)
        key_path = self.root / "synthetic-signing-key"
        if not key_path.exists():
            key = Ed25519PrivateKey.generate()
            key_path.write_bytes(key.private_bytes(Encoding.Raw, PrivateFormat.Raw, NoEncryption()))
            key_path.chmod(0o600)
        self.key = Ed25519PrivateKey.from_private_bytes(key_path.read_bytes())
        with self.connect() as db:
            db.execute("CREATE TABLE IF NOT EXISTS events(seq INTEGER PRIMARY KEY, payload BLOB NOT NULL, root TEXT NOT NULL, signature BLOB NOT NULL)")

    @contextmanager
    def connect(self):
        db = sqlite3.connect(self.path, timeout=5, isolation_level=None)
        db.execute("PRAGMA synchronous=FULL")
        db.execute("PRAGMA journal_mode=DELETE")
        try:
            yield db
        finally:
            db.close()

    def records(self):
        with self.connect() as db:
            return db.execute("SELECT seq,payload,root,signature FROM events ORDER BY seq").fetchall()

    def operation_receipt(self, intent, kind, generation=1):
        rows = self.records()
        if not rows:
            raise ValueError("operation absent")
        self.verify(rows, (rows[-1][0], rows[-1][2]))
        for row in rows:
            item = strict_decode(row[1])
            if (item["intent"], item["kind"], item["generation"]) == (intent, kind, generation):
                return row
        raise ValueError("operation absent")

    def reply(self, row, rows, intent, generation):
        state = self.verify(rows, (rows[-1][0], rows[-1][2]))[intent]
        phase = {"PREPARE": "PREPARED", "ACTIVE": "PENDING", "CANCEL": "CANCELLED", "CLAIM": "CLAIMED"}[state]
        return dict(operation_receipt=row, authoritative_current_state=dict(
            phase=phase, generation=generation, frontier_seq=rows[-1][0], frontier_root=rows[-1][2]))

    def append(self, intent, kind, now, digest="d" * 64, fail_before_commit=False,
               generation=1, checkpoint=None):
        if type(generation) is not int or generation < 1:
            raise ValueError("generation")
        with self.connect() as db:
            db.execute("BEGIN IMMEDIATE")
            rows = db.execute("SELECT seq,payload,root,signature FROM events ORDER BY seq").fetchall()
            own = [strict_decode(r[1]) for r in rows if strict_decode(r[1])["intent"] == intent]
            if own and (own[0]["digest"] != digest or own[0]["generation"] != generation):
                raise ValueError("immutable intent conflict")
            if kind == "ACTIVE":
                prepare = next((r for r in rows if strict_decode(r[1])["intent"] == intent
                                and strict_decode(r[1])["kind"] == "PREPARE"), None)
                if prepare is None:
                    raise ValueError("activation requires PREPARE")
                self.primary.validate(checkpoint, prepare, intent, generation)
            for r in rows:
                item = strict_decode(r[1])
                if item["intent"] == intent and item["kind"] == kind:
                    db.rollback()
                    return self.reply(r, rows, intent, generation)  # historical receipt, current frontier
            last = own[-1]["kind"] if own else None
            if kind == "PREPARE" and last is not None:
                raise ValueError("generation already exists")
            if kind == "ACTIVE" and last != "PREPARE":
                raise ValueError("activation requires PREPARE")
            if kind in ("CANCEL", "CLAIM"):
                if last != "ACTIVE":
                    raise ValueError("transition authority")
                deadline = own[-1]["deadline"]
                if (kind == "CANCEL" and now >= deadline) or (kind == "CLAIM" and now < deadline):
                    raise ValueError("grace boundary")
            if kind not in ("PREPARE", "ACTIVE", "CANCEL", "CLAIM"):
                raise ValueError("kind")
            item = dict(v=1, domain="oi.test.admission", epoch=1, intent=intent,
                        kind=kind, digest=digest, at=now, generation=generation,
                        deadline=now + GRACE if kind == "ACTIVE" else (own[-1]["deadline"] if own else 0))
            if kind == "ACTIVE":
                item["db_checkpoint"] = checkpoint
            if rows and now < strict_decode(rows[-1][1])["at"]:
                raise ValueError("time regression")
            payload = canonical(item)
            seq = len(rows) + 1
            prior = rows[-1][2] if rows else "0" * 64
            root = hashlib.sha256(bytes.fromhex(prior) + payload).hexdigest()
            signature = self.key.sign(signing_bytes(seq, payload, root))
            db.execute("INSERT INTO events VALUES(?,?,?,?)", (seq, payload, root, signature))
            if fail_before_commit:
                db.rollback()
                raise OSError("injected pre-commit failure")
            db.commit()
            row = (seq, payload, root, signature)
            return self.reply(row, rows + [row], intent, generation)

    def verify(self, rows, pinned):
        # pinned=(seq,root) must come from independent trusted state, not this backup.
        if not rows or (rows[-1][0], rows[-1][2]) != pinned:
            raise ValueError("coverage/quarantine")
        previous = "0" * 64
        result = {}
        prepares = {}
        for expected, (seq, payload, root, signature) in enumerate(rows, 1):
            item = strict_decode(payload)
            if seq != expected or item["epoch"] != 1 or item["domain"] != "oi.test.admission":
                raise ValueError("sequence/domain")
            if hashlib.sha256(bytes.fromhex(previous) + payload).hexdigest() != root:
                raise ValueError("chain")
            self.key.public_key().verify(signature, signing_bytes(seq, payload, root))
            if type(item["generation"]) is not int or item["generation"] < 1:
                raise ValueError("history generation")
            if item["kind"] == "PREPARE":
                prepares[item["intent"]] = (seq, payload, root, signature)
            if item["kind"] == "ACTIVE":
                prepare = prepares.get(item["intent"])
                if prepare is None or strict_decode(prepare[1])["generation"] != item["generation"]:
                    raise ValueError("history activation without matching PREPARE")
                # Historical verification checks the signed binding, not a claim that the primary still exists.
                SyntheticCheckpointDB.validate_binding(item.get("db_checkpoint"), prepare, item["intent"], item["generation"])
            previous = root
            result[item["intent"]] = item["kind"]
        return result


class AdmissionTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.journal = Journal(self.temp.name)
        self.primary = self.journal.primary

    def local_pending(self, intent, journal_available=True):
        if not journal_available:
            raise ConnectionError("witness unavailable; no local pending")
        self.journal.append(intent, "PREPARE", 100)
        return self.primary.commit_pending(self.journal, intent)

    def test_prepare_precedes_pending_and_external_ack(self):
        self.local_pending("i1")
        self.assertEqual(strict_decode(self.journal.records()[0][1])["kind"], "PREPARE")
        checkpoint = self.primary.committed("i1") or self.primary.commit_pending(self.journal, "i1")
        active = self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
        self.assertEqual(strict_decode(active["operation_receipt"][1])["deadline"], 101 + GRACE)
        self.assertIsNotNone(self.primary.committed("i1"))

    def test_lost_ack_process_exit_after_commit(self):
        code = "import runpy,os,sys;m=runpy.run_path(sys.argv[1]);m['Journal'](sys.argv[2]).append('lost','PREPARE',100);os._exit(23)"
        done = subprocess.run([sys.executable, "-c", code, str(Path(__file__).resolve()), self.temp.name], timeout=10)
        self.assertEqual(done.returncode, 23)
        reopened = Journal(self.temp.name)
        first = reopened.records()[0]
        self.assertEqual(reopened.append("lost", "PREPARE", 200)["operation_receipt"], first)
        self.assertEqual(len(reopened.records()), 1)

    def test_duplicate_concurrent_requests(self):
        with ThreadPoolExecutor(max_workers=4) as pool:
            replies = list(pool.map(lambda _: self.journal.append("i1", "PREPARE", 100), range(16)))
        self.assertTrue(all(reply == replies[0] for reply in replies))
        self.assertEqual(len(self.journal.records()), 1)

    def test_partial_journal_transaction_failure(self):
        with self.assertRaises(OSError):
            self.journal.append("i1", "PREPARE", 100, fail_before_commit=True)
        self.assertEqual(self.journal.records(), [])

    def test_outage_does_not_commit_primary_pending(self):
        with self.assertRaises(ConnectionError):
            self.local_pending("i1", journal_available=False)
        self.assertIsNone(self.primary.committed("i1"))

    def test_primary_rollback_leaves_prepared_quarantine(self):
        self.journal.append("i1", "PREPARE", 100)
        evidence = self.primary.proposed(self.journal, "i1")
        with self.primary.connect() as db:
            db.execute("BEGIN")
            db.execute("INSERT INTO pending VALUES(?,?)", ("i1", canonical(evidence)))
            db.rollback()
        self.assertIsNone(self.primary.committed("i1"))
        rows = self.journal.records()
        self.assertEqual(self.journal.verify(rows, (rows[-1][0], rows[-1][2])), {"i1": "PREPARE"})

    def test_old_primary_snapshot_recovers_pending_from_journal(self):
        self.local_pending("i1")
        self.journal.append("i1", "ACTIVE", 101, checkpoint=self.primary.committed("i1"))
        with self.primary.connect() as db:
            db.execute("DELETE FROM pending")  # simulate an older empty PRIMARY image only
        rows = self.journal.records()
        recovered = self.journal.verify(rows, (rows[-1][0], rows[-1][2]))
        self.assertEqual(recovered["i1"], "ACTIVE")  # admission ACTIVE means deletion-pending, not account-active
        self.assertNotIn(recovered["i1"], ("CANCEL",))

    def test_signed_prefix_cannot_prove_complete_history(self):
        self.journal.append("i1", "PREPARE", 100)
        self.journal.append("i2", "PREPARE", 101)
        rows = self.journal.records()
        with self.assertRaises(ValueError):
            self.journal.verify(rows[:1], (rows[-1][0], rows[-1][2]))

    def test_grace_retry_and_irreversible_boundary(self):
        self.journal.append("i1", "PREPARE", 100)
        checkpoint = self.primary.committed("i1") or self.primary.commit_pending(self.journal, "i1")
        active = self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
        self.assertEqual(self.journal.append("i1", "ACTIVE", 200, checkpoint=checkpoint), active)
        with self.assertRaises(ValueError):
            self.journal.append("i1", "CLAIM", 101 + GRACE - 1)
        self.journal.append("i1", "CLAIM", 101 + GRACE)
        with self.assertRaises(ValueError):
            self.journal.append("i1", "CANCEL", 101 + GRACE)

    def test_intent_conflict_and_clock_regression(self):
        self.journal.append("i1", "PREPARE", 100)
        with self.assertRaises(ValueError):
            self.journal.append("i1", "PREPARE", 101, digest="a" * 64)
        with self.assertRaises(ValueError):
            self.journal.append("i2", "PREPARE", 99)

    def test_binding_profile_negative_controls(self):
        for raw in ['{"v":1,"v":2}', '{"v":1.0}', '{"v":true}', '{"v":-1}', '{"v":null}']:
            with self.assertRaises(ValueError):
                strict_decode(raw)
        self.assertEqual(canonical({"b": 2, "a": "x"}), b'{"a":"x","b":2}')
        self.journal.append("i1", "PREPARE", 100)
        rows = self.journal.records()
        changed = [(rows[0][0], rows[0][1].replace(b"i1", b"i2"), rows[0][2], rows[0][3])]
        with self.assertRaises(ValueError):
            self.journal.verify(changed, (rows[0][0], rows[0][2]))


    def test_canonical_positive_roundtrip(self):
        for value in [{"a": "x", "b": 2}, {"nested": {"a": 1}}, {"a": "quote\"\\"}]:
            with self.subTest(value=value):
                raw = canonical(value)
                self.assertEqual(canonical(strict_decode(raw)), raw)
                self.assertEqual(strict_decode(raw.decode()), value)

    def test_noncanonical_byte_representations_rejected(self):
        for raw in [b'{"b":2,"a":1}', b'{ "a":1}', b'{"a":1}\n', br'{"a":"\u0078"}']:
            with self.subTest(raw=raw), self.assertRaises(ValueError):
                strict_decode(raw)

    def test_unsupported_profile_values_rejected(self):
        for raw in [b'{"a":1e0}', b'{"a":1.0}', b'{"a":-0}', b'{"a":9007199254740992}',
                    b'{"a":true}', b'{"a":null}', b'{"a":[]}', b'{"a":NaN}', b'[]', b'{"a":1,"a":1}']:
            with self.subTest(raw=raw), self.assertRaises(ValueError):
                strict_decode(raw)

    def test_verifier_rejects_even_signed_noncanonical_payload(self):
        self.journal.append("i1", "PREPARE", 100)
        seq, payload, _, _ = self.journal.records()[0]
        altered = b" " + payload
        root = hashlib.sha256(bytes(32) + altered).hexdigest()
        # Sign the altered bytes deliberately; canonical rejection must precede signature acceptance.
        message = canonical(dict(domain="oi.test.checkpoint", epoch=1, seq=seq, root=root,
                                 payload_digest=hashlib.sha256(altered).hexdigest()))
        signature = self.journal.key.sign(message)
        with self.assertRaises(ValueError):
            self.journal.verify([(seq, altered, root, signature)], (seq, root))

    def test_activation_requires_committed_checkpoint(self):
        self.journal.append("i1", "PREPARE", 100)
        with self.assertRaises(ValueError):
            self.journal.append("i1", "ACTIVE", 101)
        checkpoint = self.primary.commit_pending(self.journal, "i1")
        result = self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
        self.assertEqual(result["authoritative_current_state"]["phase"], "PENDING")
        self.assertEqual(strict_decode(result["operation_receipt"][1])["db_checkpoint"], checkpoint)

    def test_checkpoint_binding_failures(self):
        self.journal.append("i1", "PREPARE", 100)
        checkpoint = self.primary.commit_pending(self.journal, "i1")
        for key, bad in [("intent", "other"), ("generation", 2), ("prepare_seq", 2),
                         ("prepare_root", "0" * 64), ("db_version", 0), ("phase", "UNKNOWN")]:
            with self.subTest(key=key), self.assertRaises(ValueError):
                self.journal.append("i1", "ACTIVE", 101, checkpoint=dict(checkpoint, **{key: bad}))
        for bad in [None, {"outcome": "UNKNOWN"}, [checkpoint, checkpoint], dict(checkpoint, extra=1)]:
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                self.journal.append("i1", "ACTIVE", 101, checkpoint=bad)
        with self.assertRaises(ValueError):
            self.journal.append("i1", "ACTIVE", 101, generation=2, checkpoint=checkpoint)
        self.assertEqual(len(self.journal.records()), 1)

    def test_uncommitted_and_rolled_back_checkpoint_cannot_activate(self):
        self.journal.append("i1", "PREPARE", 100)
        checkpoint = self.primary.proposed(self.journal, "i1")
        with self.primary.connect() as db:
            db.execute("BEGIN")
            db.execute("INSERT INTO pending VALUES(?,?)", ("i1", canonical(checkpoint)))
            with self.assertRaises(ValueError):
                self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
            db.rollback()
        reopened = Journal(self.temp.name)
        with self.assertRaises(ValueError):
            reopened.append("i1", "ACTIVE", 102, checkpoint=checkpoint)

    def test_stale_checkpoint_and_unknown_ack_retry(self):
        self.journal.append("i1", "PREPARE", 100)
        first = self.primary.commit_pending(self.journal, "i1")
        second = self.primary.commit_pending(self.journal, "i1", db_version=2)
        with self.assertRaises(ValueError):
            Journal(self.temp.name).append("i1", "ACTIVE", 101, checkpoint=first)
        with self.assertRaises(ValueError):
            Journal(self.temp.name).append("i1", "ACTIVE", 101, checkpoint={"outcome": "UNKNOWN"})
        # Resolve uncertainty by a separate committed read; a guessed/unknown acknowledgement is insufficient.
        reopened = Journal(self.temp.name)
        self.assertEqual(reopened.primary.committed("i1"), second)
        reply = reopened.append("i1", "ACTIVE", 102, checkpoint=second)
        self.assertEqual(reopened.append("i1", "ACTIVE", 103, checkpoint=second), reply)

    def test_no_checkpoint_can_precede_prepare(self):
        with self.assertRaises(ValueError):
            self.primary.commit_pending(self.journal, "absent")
        with self.assertRaises(ValueError):
            self.journal.append("absent", "ACTIVE", 100, checkpoint={"outcome": "UNKNOWN"})
        self.assertEqual(self.journal.records(), [])

    def test_active_retry_after_cancel_separates_receipt_and_state(self):
        checkpoint = self.local_pending("i1")
        active = self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
        self.journal.append("i1", "CANCEL", 102)
        before = self.journal.records()
        for journal in [self.journal, Journal(self.temp.name)]:
            reply = journal.append("i1", "ACTIVE", 103, checkpoint=checkpoint)
            self.assertEqual(reply["operation_receipt"], active["operation_receipt"])
            self.assertEqual(strict_decode(reply["operation_receipt"][1])["kind"], "ACTIVE")
            self.assertEqual(reply["authoritative_current_state"]["phase"], "CANCELLED")
            self.assertEqual(reply["authoritative_current_state"]["frontier_seq"], 3)
            self.assertEqual(journal.records(), before)

    def test_restart_active_retry_cannot_ignore_lost_checkpoint(self):
        checkpoint = self.local_pending("i1")
        self.journal.append("i1", "ACTIVE", 101, checkpoint=checkpoint)
        with self.primary.connect() as db:
            db.execute("DELETE FROM pending")
        reopened = Journal(self.temp.name)
        with self.assertRaises(ValueError):
            reopened.append("i1", "ACTIVE", 102, checkpoint=checkpoint)
        self.assertEqual(len(reopened.records()), 2)


if __name__ == "__main__":
    unittest.main()
