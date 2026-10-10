#!/usr/bin/env python3
"""Bounded defensive verification; never targets a deployed or external service."""
import argparse
import json
import os
from pathlib import Path
import socket
import subprocess
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('mode', choices=['frontend', 'backend', 'safe-http', 'postgres'])
parser.add_argument('--output', default='audit-artifacts/module12-v2')
args = parser.parse_args()
OUT = ROOT / args.output / args.mode
OUT.mkdir(parents=True, exist_ok=True)
results = []


def record(name, status, evidence):
    results.append({'name': name, 'status': status, 'evidence': evidence})
    print(json.dumps(results[-1], sort_keys=True), flush=True)


def command(name, argv, cwd, env=None, timeout=360):
    try:
        run = subprocess.run(argv, cwd=cwd, env=env, stdout=subprocess.PIPE,
                             stderr=subprocess.STDOUT, timeout=timeout)
        (OUT / (name + '.log')).write_bytes(run.stdout)
        record(name, 'PASS' if run.returncode == 0 else 'FAILED',
               {'returncode': run.returncode, 'log': name + '.log'})
        if args.mode != 'frontend':
            for line in run.stdout.decode(errors='replace').splitlines():
                if any(k in line for k in ['M12_', 'M12_SAFE_', 'TOTAL_CANCELLATIONS=', 'FIRST_WAVE_REMOTE_',
                                          'HOSTILE_NONCOOPERATIVE_', 'BOUNDARY_BURST_',
                                          'DOWNSTREAM_DISCONNECTS=', 'DETACHED_OPERATION_']):
                    print(line, flush=True)
        return run.returncode == 0
    except subprocess.TimeoutExpired as error:
        (OUT / (name + '.log')).write_bytes(error.stdout or b'')
        record(name, 'FAILED', {'reason': 'command timeout', 'seconds': timeout})
        return False
    except OSError as error:
        record(name, 'FAILED', {'reason': str(error)})
        return False


if args.mode == 'backend':
    cwd = ROOT / 'backend-go'
    # Independent commands continue even when another command failed.
    commands = [
        ('rolling', ['go', 'test', '-race', '-v', '-timeout', '90s', './internal/sharedbudget', '-count=1']),
        ('shared-http', ['go', 'test', '-race', '-v', '-timeout', '90s', './internal/httpapi', '-run', 'TestCorporateActionShared', '-count=1']),
        ('shared-provider', ['go', 'test', '-race', '-v', '-timeout', '90s', './internal/provider/tinvest', '-run', 'TestSharedProvider', '-count=1']),
    ]
    for name in ['TestNoncooperativeUpstreamCannotRecycleCapacityOnCallerCancellation',
                 'TestDetachedOperationRetainsRealSocketCapacity',
                 'TestDetachedOperationHardTimeoutStillTerminatesLocalSocket',
                 'TestRawDownstreamDisconnectRetainsLocalOperationOwnership',
                 'TestCancelledProviderWorkRemainsBoundedByConcurrency',
                 'TestProviderPartialResetTimeoutStillFailClosedAndRecover']:
        commands.append((name, ['go', 'test', '-race', '-v', '-timeout', '90s',
                               './internal/provider/tinvest', '-run', '^' + name + '$', '-count=1']))
    for name, argv in commands:
        command(name, argv, cwd)
elif args.mode in ['safe-http', 'postgres']:
    cwd = ROOT / 'backend-go'
    if args.mode == 'safe-http':
        selectors = [('cors-host', '^TestModule12SafeCORSHostConformance$'), ('logs-errors', '^TestModule12SafeLogAndErrorConformance$'), ('sensitive-cache', '^TestSensitiveResponseCachePolicy'), ('supported-errors', 'TestAuthInfrastructureFailuresUseSanitizedInternalError|TestOINew03.*|TestCorporateActionProjectionReturns503WhenSourceIsUnavailable|TestAuthRateLimitedResponseIncludesRetryAfter|TestStage374HTTPMapsRevisionConflictAndRejectsMissingCorrectionSettlementField')]
        package = './internal/httpapi'
    else:
        selectors = [('role-boundaries', '^TestModule12SafeRuntimePermissionConformance$'), ('legitimate-runtime', '^TestOINew01RequiredRuntime'), ('timeout-recovery', '^TestRuntimeTimeoutFailuresRollbackAndLeavePoolUsable$')]
        package = './internal/postgres'
        if not os.getenv('OPENINVEST_DATABASE_RUNTIME_TEST_URL'):
            record('database-required', 'FAILED', {'reason': 'real PostgreSQL runtime URL missing'})
    for name, selector in selectors:
        command(name, ['go', 'test', '-race', '-v', '-timeout', '180s', package, '-run', selector, '-count=1'], cwd)
else:
    cwd = ROOT / 'frontend-next'
    env = dict(os.environ, NEXT_TELEMETRY_DISABLED='1')
    sentinels = {key: 'M12V2_SYNTHETIC_' + key + '_d6c24017'
                 for key in ['ACCESS_TOKEN', 'REFRESH_COOKIE', 'AUTHORIZATION_BEARER',
                             'PROVIDER_TOKEN', 'DATABASE_URL_PASSWORD', 'REDIS_URL_PASSWORD',
                             'IMPORT_REVIEW_TOKEN', 'FINANCIAL_PAYLOAD_SENTINEL']}
    # Synthetic private build inputs only; no production credentials or API calls.
    env.update(sentinels)
    env['OPENINVEST_TINVEST_TOKEN'] = sentinels['PROVIDER_TOKEN']
    env['OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET'] = sentinels['IMPORT_REVIEW_TOKEN']
    env['OPENINVEST_ACCESS_TOKEN_SECRET'] = sentinels['ACCESS_TOKEN']
    for name, argv in [('typecheck', ['pnpm', 'run', 'typecheck']),
                       ('tests', ['pnpm', 'test']), ('audit', ['pnpm', 'audit'])]:
        command(name, argv, cwd, env)
    built = command('build', ['pnpm', 'run', 'build'], cwd, env)
    if built:
        static = cwd / '.next' / 'static'
        files = [p for p in static.rglob('*') if p.is_file()]
        exposed = [str(p.relative_to(cwd)) for p in files
                   if any(v.encode() in p.read_bytes() for v in sentinels.values())]
        maps = [str(p.relative_to(cwd)) for p in files if p.suffix == '.map']
        record('public-artifacts', 'FAILED' if exposed else 'PASS',
               {'source_map_count': len(maps), 'sentinel_files': exposed, 'files_scanned': len(files)})
        sock = socket.socket()
        sock.bind(('127.0.0.1', 0))
        port = sock.getsockname()[1]
        sock.close()
        log = (OUT / 'runtime.log').open('wb')
        process = subprocess.Popen(['pnpm', 'exec', 'next', 'start', '-H', '127.0.0.1', '-p', str(port)],
                                   cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT,
                                   start_new_session=True)
        origin = 'http://127.0.0.1:' + str(port)
        try:
            deadline = time.monotonic() + 30
            ready = False
            while time.monotonic() < deadline:
                try:
                    with urllib.request.urlopen(origin, timeout=1) as response:
                        response.read()
                    ready = True
                    break
                except (OSError, urllib.error.URLError):
                    if process.poll() is not None:
                        break
                    time.sleep(.1)
            if not ready:
                record('runtime-ready', 'FAILED', {'reason': 'local Next runtime not ready'})
            else:
                for label, path in [('root', '/'), ('not-found', '/m12-v2-missing-page')]:
                    try:
                        try:
                            response = urllib.request.urlopen(origin + path, timeout=5)
                        except urllib.error.HTTPError as error:
                            response = error
                        with response:
                            body = response.read()
                            headers = {k.lower(): v for k, v in response.headers.items()}
                            secret_hits = [k for k, v in sentinels.items() if v.encode() in body]
                            expected = 200 if label == 'root' else 404
                            record('runtime-' + label, 'PASS' if response.status == expected and not secret_hits else 'FAILED',
                                   {'status': response.status, 'secret_hits': secret_hits,
                                    'headers': {k: headers.get(k) for k in ['content-security-policy',
                                        'strict-transport-security', 'x-content-type-options', 'x-frame-options',
                                        'referrer-policy', 'permissions-policy', 'cache-control', 'vary',
                                        'set-cookie', 'etag', 'x-powered-by', 'location']}})
                    except Exception as error:
                        record('runtime-' + label, 'FAILED', {'reason': str(error)})
        finally:
            import signal
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait(timeout=5)
            log.close()

(OUT / 'summary.json').write_text(json.dumps(results, indent=2) + '\n')
raise SystemExit(1 if any(r['status'] == 'FAILED' for r in results) else 0)
