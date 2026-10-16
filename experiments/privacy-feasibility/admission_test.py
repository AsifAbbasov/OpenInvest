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
    canonical(value)
    return value


class Journal:
    def __init__(self, directory):
        self.root = Path(directory)
        self.path = self.root / "independent-journal.sqlite"
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

    def append(self, intent, kind, now, digest="d" * 64, fail_before_commit=False):
        with self.connect() as db:
            db.execute("BEGIN IMMEDIATE")
            rows = db.execute("SELECT seq,payload,root,signature FROM events ORDER BY seq").fetchall()
            own = [strict_decode(r[1]) for r in rows if strict_decode(r[1])["intent"] == intent]
            if own and own[0]["digest"] != digest:
                raise ValueError("immutable intent conflict")
            for r in rows:
                item = strict_decode(r[1])
                if item["intent"] == intent and item["kind"] == kind:
                    db.rollback()
                    return r  # immutable idempotent reply; no new deadline
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
                        kind=kind, digest=digest, at=now,
                        deadline=now + GRACE if kind == "ACTIVE" else (own[-1]["deadline"] if own else 0))
            if rows and now < strict_decode(rows[-1][1])["at"]:
                raise ValueError("time regression")
            payload = canonical(item)
            seq = len(rows) + 1
            prior = rows[-1][2] if rows else "0" * 64
            root = hashlib.sha256(bytes.fromhex(prior) + payload).hexdigest()
            signature = self.key.sign(canonical(dict(domain="oi.test.checkpoint", epoch=1, seq=seq, root=root)))
            db.execute("INSERT INTO events VALUES(?,?,?,?)", (seq, payload, root, signature))
            if fail_before_commit:
                db.rollback()
                raise OSError("injected pre-commit failure")
            db.commit()
            return seq, payload, root, signature

    def verify(self, rows, pinned):
        # pinned=(seq,root) must come from independent trusted state, not this backup.
        if not rows or (rows[-1][0], rows[-1][2]) != pinned:
            raise ValueError("coverage/quarantine")
        previous = "0" * 64
        result = {}
        for expected, (seq, payload, root, signature) in enumerate(rows, 1):
            item = strict_decode(payload)
            if seq != expected or item["epoch"] != 1 or item["domain"] != "oi.test.admission":
                raise ValueError("sequence/domain")
            if hashlib.sha256(bytes.fromhex(previous) + payload).hexdigest() != root:
                raise ValueError("chain")
            self.key.public_key().verify(signature, canonical(dict(domain="oi.test.checkpoint", epoch=1, seq=seq, root=root)))
            previous = root
            result[item["intent"]] = item["kind"]
        return result


class AdmissionTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.journal = Journal(self.temp.name)
        self.primary = sqlite3.connect(":memory:")
        self.addCleanup(self.primary.close)
        self.primary.execute("CREATE TABLE pending(intent TEXT PRIMARY KEY)")

    def local_pending(self, intent, journal_available=True):
        if not journal_available:
            raise ConnectionError("witness unavailable; no local pending")
        self.journal.append(intent, "PREPARE", 100)
        self.primary.execute("INSERT INTO pending VALUES(?)", (intent,))
        self.primary.commit()

    def test_prepare_precedes_pending_and_external_ack(self):
        self.local_pending("i1")
        self.assertEqual(strict_decode(self.journal.records()[0][1])["kind"], "PREPARE")
        active = self.journal.append("i1", "ACTIVE", 101)
        self.assertEqual(strict_decode(active[1])["deadline"], 101 + GRACE)
        self.assertEqual(self.primary.execute("SELECT count(*) FROM pending").fetchone()[0], 1)

    def test_lost_ack_process_exit_after_commit(self):
        code = "import runpy,os,sys;m=runpy.run_path(sys.argv[1]);m['Journal'](sys.argv[2]).append('lost','PREPARE',100);os._exit(23)"
        done = subprocess.run([sys.executable, "-c", code, str(Path(__file__).resolve()), self.temp.name], timeout=10)
        self.assertEqual(done.returncode, 23)
        reopened = Journal(self.temp.name)
        first = reopened.records()[0]
        self.assertEqual(reopened.append("lost", "PREPARE", 200), first)
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
        self.assertEqual(self.primary.execute("SELECT count(*) FROM pending").fetchone()[0], 0)

    def test_primary_rollback_leaves_prepared_quarantine(self):
        self.journal.append("i1", "PREPARE", 100)
        self.primary.execute("INSERT INTO pending VALUES('i1')")
        self.primary.rollback()
        self.assertEqual(self.primary.execute("SELECT count(*) FROM pending").fetchone()[0], 0)
        rows = self.journal.records()
        self.assertEqual(self.journal.verify(rows, (rows[-1][0], rows[-1][2])), {"i1": "PREPARE"})

    def test_old_primary_snapshot_recovers_pending_from_journal(self):
        self.local_pending("i1")
        self.journal.append("i1", "ACTIVE", 101)
        self.primary.execute("DELETE FROM pending")  # simulate an older empty PRIMARY image only
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
        active = self.journal.append("i1", "ACTIVE", 101)
        self.assertEqual(self.journal.append("i1", "ACTIVE", 200), active)
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


if __name__ == "__main__":
    unittest.main()
