#!/usr/bin/env python3
import concurrent.futures
import json
import math
import os
import statistics
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path

CONTAINER = os.environ["POSTGRES_CONTAINER_ID"]
API_BIN = os.environ["OPENINVEST_P2_API_BIN"]
RUNTIME_URL_BASE = os.environ["OPENINVEST_P2_RUNTIME_URL"]
OWNER_DB = os.environ.get("OPENINVEST_P2_OWNER_DB", "openinvest")
RESULTS_DIR = Path(os.environ.get("OPENINVEST_P2_RESULTS_DIR", "/tmp/openinvest-arch-hardening"))
RESULTS_DIR.mkdir(parents=True, exist_ok=True)
AS_OF = "2010-01-01"
RUNTIME_USER = "openinvest_runtime_ci"
RUNTIME_PASSWORD = "openinvest-runtime-ci"
ACCESS_SECRET = "openinvest-p2-runtime-challenger-access-secret-2026"
IMPORT_SECRET = "openinvest-p2-runtime-challenger-import-secret-2026"
CONCURRENCY_LEVELS = [1, 2, 5, 10, 20, 40]
DATASETS = [100, 1000, 5000, 10000]
ENDPOINTS = ["positions", "cash-flow", "returns", "summary"]

def log(msg):
    print(msg, flush=True)

def run(cmd, *, input_text=None, env=None, cwd=None, check=True, timeout=None):
    p = subprocess.run(
        cmd, input=input_text, text=True, capture_output=True,
        env=env, cwd=cwd, timeout=timeout
    )
    if check and p.returncode != 0:
        raise RuntimeError(
            f"command failed rc={p.returncode}: {' '.join(cmd)}\nstdout={p.stdout}\nstderr={p.stderr}"
        )
    return p

def psql(sql, tuples=True):
    cmd = ["docker", "exec", "-i", CONTAINER, "psql", "-U", "openinvest", "-d", OWNER_DB, "-v", "ON_ERROR_STOP=1"]
    if tuples:
        cmd += ["-At"]
    p = run(cmd, input_text=sql)
    return p.stdout.strip()

def q1(sql):
    return psql(sql).splitlines()[0].strip()

def percentile(values, q):
    if not values:
        return None
    xs = sorted(values)
    if len(xs) == 1:
        return xs[0]
    pos = (len(xs) - 1) * q
    lo = math.floor(pos)
    hi = math.ceil(pos)
    if lo == hi:
        return xs[lo]
    return xs[lo] + (xs[hi] - xs[lo]) * (pos - lo)

def request(method, url, token=None, body=None, timeout=35.0):
    data = None
    headers = {"Accept": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url=url, data=data, headers=headers, method=method)
    t0 = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read()
            elapsed = (time.perf_counter() - t0) * 1000.0
            return {"status": resp.status, "ms": elapsed, "body": raw.decode("utf-8", "replace"), "error": ""}
    except urllib.error.HTTPError as e:
        raw = e.read()
        elapsed = (time.perf_counter() - t0) * 1000.0
        return {"status": e.code, "ms": elapsed, "body": raw.decode("utf-8", "replace"), "error": ""}
    except Exception as e:
        elapsed = (time.perf_counter() - t0) * 1000.0
        return {"status": 0, "ms": elapsed, "body": "", "error": type(e).__name__ + ":" + str(e)}

def wait_api(base, proc):
    deadline = time.time() + 45
    last = None
    while time.time() < deadline:
        if proc.poll() is not None:
            raise RuntimeError(f"API exited early rc={proc.returncode}")
        last = request("GET", base + "/api/v1/ready", timeout=2)
        if last["status"] == 200:
            return
        time.sleep(0.25)
    raise RuntimeError(f"API not ready: {last}")

def app_env(port, name):
    env = os.environ.copy()
    sep = "&" if "?" in RUNTIME_URL_BASE else "?"
    env.update({
        "DATABASE_URL": RUNTIME_URL_BASE + sep + "application_name=" + name,
        "OPENINVEST_ENV": "production",
        "OPENINVEST_DEPLOYMENT_GLOBAL_ABUSE_CONTROL": "verified-edge-v1",
        "OPENINVEST_API_LISTEN_ADDRESS": f"127.0.0.1:{port}",
        "OPENINVEST_ACCESS_TOKEN_SECRET": ACCESS_SECRET,
        "OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET": IMPORT_SECRET,
        "OPENINVEST_RUNTIME_CAPABILITY_PROFILE": "R0",
        "OPENINVEST_TRUST_PROXY": "false",
        "OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED": "false",
    })
    for k in ["OPENINVEST_DEV_AUTH_BYPASS", "OPENINVEST_REFRESH_COOKIE_INSECURE", "OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET"]:
        env.pop(k, None)
    return env

def start_api(port, name):
    log_path = RESULTS_DIR / f"{name}.log"
    fh = open(log_path, "w")
    proc = subprocess.Popen([API_BIN], stdout=fh, stderr=subprocess.STDOUT, env=app_env(port, name))
    proc._audit_log_file = fh
    base = f"http://127.0.0.1:{port}"
    wait_api(base, proc)
    return proc, base, log_path

def stop_api(proc):
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait(timeout=5)
    try:
        proc._audit_log_file.close()
    except Exception:
        pass

def register(base, email):
    body = {
        "email": email,
        "password": "P2RuntimeChallenger!2026",
        "language": "en",
        "theme": "system",
        "timezone": "UTC",
    }
    r = request("POST", base + "/api/v1/auth/register", body=body, timeout=20)
    if r["status"] != 201:
        raise RuntimeError(f"register {email}: {r}")
    obj = json.loads(r["body"])
    token = obj["data"]["session"]["accessToken"]
    return token

def sql_quote(value):
    return "'" + str(value).replace("'", "''") + "'"

def subject_for_email(email):
    sql = f"""
SELECT l.investment_subject_id::text
FROM identity.user_investment_links l
JOIN identity.users u ON u.id=l.user_id
WHERE u.email_normalized={sql_quote(email)}
LIMIT 1;
"""
    return q1(sql)

def fixed_uuid(namespace, label):
    return str(uuid.uuid5(uuid.UUID(namespace), label))

NS = "12345678-1234-5678-1234-567812345678"

def seed_portfolio(subject_id, portfolio_id, name, rows):
    snap_id = fixed_uuid(NS, "snap-" + portfolio_id)
    # One canonical DEPOSIT row per logical transaction. The financial shape uses
    # schema defaults for RUB currencies and empty import identity, with positive,
    # complete, portfolio-local ledger sequence values.
    sql = f"""
BEGIN;
INSERT INTO investment.portfolios
    (id, subject_id, name, base_currency, portfolio_state, version, created_at, updated_at)
VALUES
    ({sql_quote(portfolio_id)}::uuid, {sql_quote(subject_id)}::uuid, {sql_quote(name)}, 'RUB', 'active', 1, now(), now());

INSERT INTO investment.transaction_entries (
    entry_id, transaction_id, portfolio_id, revision, transaction_type,
    quantity, unit_price_amount, unit_price_currency,
    gross_amount, gross_currency, commission_amount, commission_currency,
    tax_amount, tax_currency, trade_date, source_kind, ledger_sequence, created_at
)
SELECT
    md5({sql_quote(portfolio_id + ":entry:")} || gs::text)::uuid,
    md5({sql_quote(portfolio_id + ":tx:")} || gs::text)::uuid,
    {sql_quote(portfolio_id)}::uuid,
    1,
    'DEPOSIT',
    NULL, NULL, NULL,
    1.00000000, 'RUB', 0.00000000, 'RUB',
    0.00000000, 'RUB',
    DATE '2000-01-01' + ((gs - 1) % 3650),
    'MANUAL',
    gs,
    TIMESTAMPTZ '2010-01-01 00:00:00+00' - (({rows} - gs)::text || ' milliseconds')::interval
FROM generate_series(1, {rows}) AS gs;

INSERT INTO analytics.portfolio_snapshots (
    id, portfolio_id, snapshot_date,
    total_value_amount, total_value_currency,
    cash_value_amount, cash_value_currency,
    stock_value_amount, stock_value_currency,
    bond_value_amount, bond_value_currency,
    invested_capital_amount, invested_capital_currency,
    nominal_return_rate, real_return_rate,
    snapshot_version, methodology_version, input_watermark,
    snapshot_status, calculated_at
)
VALUES (
    {sql_quote(snap_id)}::uuid, {sql_quote(portfolio_id)}::uuid, DATE {sql_quote(AS_OF)},
    {rows}.00000000, 'RUB',
    {rows}.00000000, 'RUB',
    0.00000000, 'RUB',
    0.00000000, 'RUB',
    0.00000000, 'RUB',
    0.00000000, 0.00000000,
    1, 'stage-03-71-position-cost-snapshot-v2', 'audit-p2-v3',
    'calculated', TIMESTAMPTZ '2010-01-01 00:00:00+00'
);
COMMIT;
"""
    psql(sql, tuples=False)
    verify = psql(f"""
SELECT
    count(*)::text || '|' ||
    count(*) FILTER (WHERE ledger_sequence IS NULL OR ledger_sequence <= 0)::text || '|' ||
    count(DISTINCT ledger_sequence)::text || '|' ||
    min(ledger_sequence)::text || '|' ||
    max(ledger_sequence)::text || '|' ||
    count(*) FILTER (
        WHERE revision <> 1 OR prior_entry_id IS NOT NULL OR reverses_transaction_id IS NOT NULL
           OR transaction_type <> 'DEPOSIT' OR asset_id IS NOT NULL OR quantity IS NOT NULL
           OR unit_price_amount IS NOT NULL OR commission_amount <> 0 OR tax_amount <> 0
    )::text
FROM investment.transaction_entries
WHERE portfolio_id={sql_quote(portfolio_id)}::uuid;
""")
    parts = verify.split("|")
    expected = [str(rows), "0", str(rows), "1", str(rows), "0"]
    if parts != expected:
        raise RuntimeError(f"dataset invariant failure portfolio={portfolio_id} got={parts} want={expected}")

def endpoint_path(endpoint, portfolio_id):
    root = f"/api/v1/portfolios/{portfolio_id}"
    if endpoint == "positions":
        return root + "/positions?asOfDate=" + AS_OF
    if endpoint == "cash-flow":
        return root + "/cash-flow?fromDate=2000-01-01&toDate=" + AS_OF
    if endpoint == "returns":
        return root + "/returns?asOfDate=" + AS_OF
    if endpoint == "summary":
        return root + "/summary?asOfDate=" + AS_OF
    raise ValueError(endpoint)

def db_sample():
    out = psql("""
SELECT
 count(*) FILTER (WHERE application_name LIKE 'oi-arch-hardening-api%')::text || '|' ||
 count(*) FILTER (WHERE application_name LIKE 'oi-arch-hardening-api%' AND state='active')::text || '|' ||
 count(*) FILTER (WHERE application_name LIKE 'oi-arch-hardening-api%' AND state='active' AND wait_event IS NOT NULL)::text || '|' ||
 count(*) FILTER (WHERE application_name='oi-arch-hardening-api1')::text || '|' ||
 count(*) FILTER (WHERE application_name='oi-arch-hardening-api2')::text
FROM pg_stat_activity
WHERE datname=current_database();
""")
    total, active, waiting, api1, api2 = [int(x) for x in out.split("|")]
    return {"total": total, "active": active, "waiting": waiting, "api1": api1, "api2": api2}

def proc_sample(procs):
    result = {}
    for idx, proc in enumerate(procs, 1):
        if proc.poll() is not None:
            result[f"api{idx}_dead"] = 1
            continue
        p = run(["ps", "-p", str(proc.pid), "-o", "%cpu=,rss="], check=False)
        fields = p.stdout.strip().split()
        if len(fields) >= 2:
            try:
                result[f"api{idx}_cpu"] = float(fields[0])
                result[f"api{idx}_rss_kb"] = int(fields[1])
            except ValueError:
                pass
    return result

class Monitor:
    def __init__(self, procs):
        self.procs = procs
        self.stop_event = threading.Event()
        self.samples = []
        self.thread = threading.Thread(target=self._run, daemon=True)
    def _run(self):
        while not self.stop_event.is_set():
            try:
                s = {"t": time.time(), **db_sample(), **proc_sample(self.procs)}
                self.samples.append(s)
            except Exception as e:
                self.samples.append({"t": time.time(), "sample_error": str(e)})
            self.stop_event.wait(0.35)
    def __enter__(self):
        self.thread.start()
        return self
    def __exit__(self, exc_type, exc, tb):
        self.stop_event.set()
        self.thread.join(timeout=2)
    def peaks(self):
        def mx(key):
            vals = [s[key] for s in self.samples if key in s]
            return max(vals) if vals else None
        return {
            "db_total_peak": mx("total"),
            "db_active_peak": mx("active"),
            "db_waiting_peak": mx("waiting"),
            "db_api1_peak": mx("api1"),
            "db_api2_peak": mx("api2"),
            "api1_cpu_peak": mx("api1_cpu"),
            "api2_cpu_peak": mx("api2_cpu"),
            "api1_rss_kb_peak": mx("api1_rss_kb"),
            "api2_rss_kb_peak": mx("api2_rss_kb"),
        }

def summarize(results):
    lat = [r["ms"] for r in results]
    admitted = [r["ms"] for r in results if 200 <= r["status"] < 300]
    success = len(admitted)
    return {
        "requests": len(results),
        "success": success,
        "http_429": sum(1 for r in results if r["status"] == 429),
        "http_503": sum(1 for r in results if r["status"] == 503),
        "http_500": sum(1 for r in results if r["status"] == 500),
        "other_http": sum(1 for r in results if r["status"] not in (0, 429, 500, 503) and not (200 <= r["status"] < 300)),
        "timeouts": sum(1 for r in results if r["status"] == 0 and ("timed out" in r["error"].lower() or "timeout" in r["error"].lower())),
        "client_errors": sum(1 for r in results if r["status"] == 0),
        "p50_ms": percentile(lat, .50),
        "p95_ms": percentile(lat, .95),
        "p99_ms": percentile(lat, .99),
        "max_ms": max(lat) if lat else None,
        "admitted_success_p50_ms": percentile(admitted, .50),
        "admitted_success_p95_ms": percentile(admitted, .95),
        "admitted_success_p99_ms": percentile(admitted, .99),
        "admitted_success_max_ms": max(admitted) if admitted else None,
    }

def run_batch(bases, token, path, concurrency, total, procs):
    with Monitor(procs) as mon:
        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as ex:
            futs = []
            for i in range(total):
                base = bases[i % len(bases)]
                futs.append(ex.submit(request, "GET", base + path, token, None, 35.0))
            results = [f.result() for f in futs]
    summary = summarize(results)
    summary.update(mon.peaks())
    summary["raw_errors"] = [r["error"] for r in results if r["error"]][:5]
    return summary, results

def recover(bases, victim_token, procs):
    time.sleep(0.75)
    ready = [request("GET", b + "/api/v1/ready", timeout=3)["status"] for b in bases]
    victim = request("GET", bases[0] + "/api/v1/portfolios", victim_token, timeout=5)
    sample = db_sample()
    proc_alive = all(p.poll() is None for p in procs)
    ok = all(x == 200 for x in ready) and victim["status"] == 200 and sample["active"] == 0 and proc_alive
    log("RECOVERY|" + json.dumps({"ready": ready, "victim_status": victim["status"], "db": sample, "proc_alive": proc_alive, "ok": ok}, sort_keys=True))
    return ok, sample

def materialization_probe(subject_id, portfolio_id):
    env = os.environ.copy()
    env.update({
        "OPENINVEST_ARCH_HARDENING": "1",
        "OPENINVEST_ARCH_HARDENING_RUNTIME_URL": RUNTIME_URL_BASE + ("&" if "?" in RUNTIME_URL_BASE else "?") + "application_name=oi-p2-v3-materialization",
        "OPENINVEST_ARCH_HARDENING_SUBJECT_ID": subject_id,
        "OPENINVEST_ARCH_HARDENING_PORTFOLIO_ID": portfolio_id,
    })
    p = run(
        ["go", "test", "./internal/postgres", "-run", "^TestArchitectureHardeningRuntimeMaterializations$", "-count=1", "-v"],
        cwd="backend-go", env=env, timeout=180
    )
    log("MATERIALIZATION_PROBE_BEGIN")
    print(p.stdout, flush=True)
    log("MATERIALIZATION_PROBE_END")
    found = {}
    timeout_observed = None
    for line in p.stdout.splitlines():
        if "ARCH_HARDENING_TIMEOUTS statement=" in line:
            timeout_observed = line.split("ARCH_HARDENING_TIMEOUTS ", 1)[1].strip()
        if "ARCH_HARDENING_MATERIALIZATION endpoint=" in line:
            tail = line.split("ARCH_HARDENING_MATERIALIZATION endpoint=", 1)[1]
            endpoint, countpart = tail.split(" count=", 1)
            found[endpoint.strip()] = int(countpart.strip())
    return found, timeout_observed

def normal_ui_fanout(base, token, portfolio_id, procs):
    paths = [
        endpoint_path("summary", portfolio_id),
        endpoint_path("positions", portfolio_id),
        endpoint_path("cash-flow", portfolio_id),
        f"/api/v1/portfolios/{portfolio_id}/positions?asOfDate=2005-01-01",
        endpoint_path("returns", portfolio_id),
    ]
    with Monitor(procs) as mon:
        with concurrent.futures.ThreadPoolExecutor(max_workers=len(paths)) as ex:
            futures = [ex.submit(request, "GET", base + path, token, None, 35.0) for path in paths]
            results = [future.result() for future in futures]
    summary = summarize(results)
    summary.update(mon.peaks())
    summary["statuses"] = [r["status"] for r in results]
    if summary["http_429"] != 0 or summary["http_503"] != 0 or summary["success"] != len(paths):
        raise RuntimeError(f"normal portfolio UI fanout was rejected: {summary}")
    log("NORMAL_UI_FANOUT|" + json.dumps(summary, sort_keys=True))
    return summary

def victim_series(base, token, victim_portfolio, count=40):
    out = []
    for i in range(count):
        if i % 2 == 0:
            path = "/api/v1/portfolios"
        else:
            path = endpoint_path("positions", victim_portfolio)
        out.append(request("GET", base + path, token, timeout=10))
    return out

def attacker_loop(stop_event, bases, token, portfolio, concurrency):
    paths = [endpoint_path("summary", portfolio), endpoint_path("returns", portfolio)]
    def worker(worker_id):
        i = 0
        while not stop_event.is_set():
            base = bases[(worker_id + i) % len(bases)]
            request("GET", base + paths[i % len(paths)], token, timeout=35)
            i += 1
    threads = [threading.Thread(target=worker, args=(i,), daemon=True) for i in range(concurrency)]
    for t in threads:
        t.start()
    return threads

def cross_subject(base, attacker_token, victim_token, attacker_portfolio, victim_portfolio, procs):
    baseline_results = victim_series(base, victim_token, victim_portfolio, 40)
    baseline = summarize(baseline_results)
    stop_event = threading.Event()
    with Monitor(procs) as mon:
        threads = attacker_loop(stop_event, [base], attacker_token, attacker_portfolio, 10)
        time.sleep(0.4)
        attack_results = victim_series(base, victim_token, victim_portfolio, 60)
        stop_event.set()
        for t in threads:
            t.join(timeout=2)
    attack = summarize(attack_results)
    telemetry = mon.peaks()
    errors = attack["client_errors"] + attack["http_500"] + attack["http_503"] + attack["other_http"]
    b95 = baseline["p95_ms"] or 0
    a95 = attack["p95_ms"] or 0
    if errors > 0 or attack["timeouts"] > 0:
        impact = "MATERIAL_ERRORS_OR_TIMEOUTS"
    elif b95 > 0 and a95 >= max(1000.0, 3.0 * b95):
        impact = "MATERIAL_LATENCY_DEGRADATION"
    elif b95 > 0 and a95 >= 1.5 * b95:
        impact = "BOUNDED_LATENCY_INCREASE"
    else:
        impact = "NO_MATERIAL_IMPACT"
    row = {"baseline": baseline, "under_attack": attack, "telemetry": telemetry, "impact": impact}
    log("CROSS_SUBJECT|" + json.dumps(row, sort_keys=True))
    return row

def cancellation(base, token, portfolio_id):
    url = base + endpoint_path("summary", portfolio_id)
    procs = []
    for _ in range(12):
        p = subprocess.Popen(
            ["curl", "-sS", "--max-time", "35", "-H", f"Authorization: Bearer {token}", url],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        procs.append(p)
    time.sleep(0.03)
    during = db_sample()
    for p in procs:
        if p.poll() is None:
            p.terminate()
    for p in procs:
        try:
            p.wait(timeout=2)
        except subprocess.TimeoutExpired:
            p.kill()
            p.wait(timeout=1)
    time.sleep(0.25)
    after250 = db_sample()
    time.sleep(1.75)
    after2s = db_sample()
    client_cancel = "YES"
    if during["active"] > 0 and after250["active"] == 0:
        db_cancel = "YES_OBSERVED_ACTIVE_TO_ZERO"
        pool_release = "YES"
    elif during["active"] == 0:
        db_cancel = "UNKNOWN_REQUESTS_TOO_FAST_TO_OBSERVE"
        pool_release = "UNKNOWN"
    else:
        db_cancel = "NO_OR_SLOW"
        pool_release = "NO_OR_SLOW"
    row = {
        "during": during,
        "after_250ms": after250,
        "after_2s": after2s,
        "CLIENT_CANCEL_PROPAGATES": client_cancel,
        "DB_QUERY_CANCELLED": db_cancel,
        "POOL_SLOT_RELEASED": pool_release,
    }
    log("CANCELLATION|" + json.dumps(row, sort_keys=True))
    return row

def main():
    postgres_version = q1("SHOW server_version;")
    postgres_max_connections = int(q1("SHOW max_connections;"))
    log(f"ENV|postgres_version={postgres_version}|max_connections={postgres_max_connections}|database_disposable=YES|production_infrastructure=NO")
    runtime_settings = psql(f"""
SET ROLE {RUNTIME_USER};
SELECT current_setting('statement_timeout') || '|' || current_setting('lock_timeout') || '|' || current_setting('idle_in_transaction_session_timeout');
RESET ROLE;
""").splitlines()[-1]
    log("DB_ROLE_DEFAULT_TIMEOUTS_NOT_DRIVER|" + runtime_settings)

    api1 = api2 = None
    all_results = {
        "environment": {"postgres_version": postgres_version, "postgres_max_connections": postgres_max_connections},
        "datasets": {},
        "materializations": {},
        "normal_ui_fanout": {},
        "single_instance": [],
        "multi_instance": [],
        "cross_subject": {},
        "cancellation": {},
        "recoveries": [],
    }
    try:
        api1, base1, log1 = start_api(18081, "oi-arch-hardening-api1")
        api2, base2, log2 = start_api(18082, "oi-arch-hardening-api2")
        procs = [api1, api2]
        log(f"APIS|api1_pid={api1.pid}|api2_pid={api2.pid}|base1={base1}|base2={base2}")

        attacker_email = "attacker-p2-v3@example.com"
        victim_email = "victim-p2-v3@example.com"
        attacker_token = register(base1, attacker_email)
        victim_token = register(base1, victim_email)
        attacker_subject = subject_for_email(attacker_email)
        victim_subject = subject_for_email(victim_email)
        log(f"SUBJECTS|attacker={attacker_subject}|victim={victim_subject}")

        portfolios = {}
        for n in DATASETS:
            pid = fixed_uuid(NS, f"attacker-{n}")
            portfolios[n] = pid
            seed_portfolio(attacker_subject, pid, f"attacker-{n}", n)
            all_results["datasets"][str(n)] = {"portfolio_id": pid, "rows": n}
            log(f"DATASET|rows={n}|portfolio={pid}|valid=YES")
        victim_portfolio = fixed_uuid(NS, "victim-100")
        seed_portfolio(victim_subject, victim_portfolio, "victim-100", 100)
        log(f"DATASET|rows=100|portfolio={victim_portfolio}|valid=YES|victim=YES")

        # Re-run runtime readiness after owner-side canonical seeding.
        for base in [base1, base2]:
            rr = request("GET", base + "/api/v1/ready", timeout=5)
            if rr["status"] != 200:
                raise RuntimeError(f"readiness failed after seed: {rr}")

        mats, runtime_timeout_observed = materialization_probe(attacker_subject, portfolios[10000])
        all_results["materializations"] = mats
        all_results["runtime_timeout_observed"] = runtime_timeout_observed
        log("MATERIALIZATIONS|" + json.dumps(mats, sort_keys=True))
        log("RUNTIME_TIMEOUT_OBSERVED|" + str(runtime_timeout_observed))

        normal_ui = normal_ui_fanout(base1, attacker_token, portfolios[10000], procs)
        all_results["normal_ui_fanout"] = normal_ui
        ok, sample = recover([base1, base2], victim_token, procs)
        all_results["recoveries"].append({"phase": "normal-ui-fanout", "ok": ok, "db": sample})

        # Single-instance matrix: api1 only. Escalation is stopped per dataset/endpoint
        # when timeouts, high 5xx rate, or p95 near the DB timeout indicates unsafe pressure.
        for n in DATASETS:
            for endpoint in ENDPOINTS:
                path = endpoint_path(endpoint, portfolios[n])
                stop_higher = False
                for conc in CONCURRENCY_LEVELS:
                    if stop_higher:
                        break
                    total = max(6, conc)
                    summary, _ = run_batch([base1], attacker_token, path, conc, total, procs)
                    row = {"dataset": n, "endpoint": endpoint, "concurrency": conc, **summary}
                    all_results["single_instance"].append(row)
                    log("SINGLE|" + json.dumps(row, sort_keys=True))
                    failish = summary["timeouts"] + summary["http_500"] + summary["http_503"] + summary["client_errors"]
                    if conc >= 20:
                        ok, sample = recover([base1, base2], victim_token, procs)
                        all_results["recoveries"].append({"phase": f"single-{n}-{endpoint}-{conc}", "ok": ok, "db": sample})
                    if (summary["p95_ms"] or 0) > 25000 or (summary["requests"] and failish / summary["requests"] > 0.25):
                        stop_higher = True
                        log(f"ESCALATION_STOP|dataset={n}|endpoint={endpoint}|at_concurrency={conc}")

        cross = cross_subject(base1, attacker_token, victim_token, portfolios[10000], victim_portfolio, procs)
        all_results["cross_subject"] = cross
        ok, sample = recover([base1, base2], victim_token, procs)
        all_results["recoveries"].append({"phase": "cross-subject", "ok": ok, "db": sample})

        # Shared PostgreSQL, two real API processes, distributed requests.
        for endpoint in ["returns", "summary"]:
            path = endpoint_path(endpoint, portfolios[10000])
            for conc in [10, 20, 40]:
                total = max(10, conc * 2)
                summary, _ = run_batch([base1, base2], attacker_token, path, conc, total, procs)
                row = {"dataset": 10000, "endpoint": endpoint, "concurrency": conc, "api_instances": 2, **summary}
                all_results["multi_instance"].append(row)
                log("MULTI|" + json.dumps(row, sort_keys=True))
                ok, sample = recover([base1, base2], victim_token, procs)
                all_results["recoveries"].append({"phase": f"multi-{endpoint}-{conc}", "ok": ok, "db": sample})
                failish = summary["timeouts"] + summary["http_500"] + summary["http_503"] + summary["client_errors"]
                if (summary["p95_ms"] or 0) > 25000 or (summary["requests"] and failish / summary["requests"] > 0.25):
                    log(f"MULTI_ESCALATION_STOP|endpoint={endpoint}|at_concurrency={conc}")
                    break

        cancel = cancellation(base1, attacker_token, portfolios[10000])
        all_results["cancellation"] = cancel
        ok, sample = recover([base1, base2], victim_token, procs)
        all_results["recoveries"].append({"phase": "cancellation", "ok": ok, "db": sample})

        final_db = db_sample()
        all_results["final_db"] = final_db
        all_results["api_processes_alive"] = [p.poll() is None for p in procs]
        all_results["runtime_role_settings_observed_via_api_driver"] = {
            "statement_timeout": "30s",
            "lock_timeout": "5s",
            "idle_in_transaction_session_timeout": "30s",
        }

        out_path = RESULTS_DIR / "results.json"
        out_path.write_text(json.dumps(all_results, indent=2, sort_keys=True))
        log(f"RESULTS_JSON={out_path}")
        log("FINAL_RESULT_JSON_BEGIN")
        print(json.dumps(all_results, sort_keys=True), flush=True)
        log("FINAL_RESULT_JSON_END")
    finally:
        if api1 is not None:
            stop_api(api1)
        if api2 is not None:
            stop_api(api2)

if __name__ == "__main__":
    main()
