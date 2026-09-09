#!/usr/bin/env python3
"""Trigger a rejudge storm for benchmark experiments.

This script is intended to run alongside normal submission workloads to test whether
rejudges interfere with fresh submissions.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import time
import urllib.error
import urllib.request
import uuid
from http.cookiejar import CookieJar
from pathlib import Path
from typing import Any

DEFAULT_BASE_URL = "http://localhost:8080"
DEFAULT_API_PREFIX = "/api/v1"
DEFAULT_OUTPUT_DIR = "scripts/output/queue-benchmark"
DEFAULT_CONTEST_NAME = "Queue Benchmark Contest"
DEFAULT_CONTEST_PROJECT_NAME = "Queue Benchmark Short Task"
DEFAULT_ADMIN_USERNAME = "admin"
DEFAULT_ADMIN_PASSWORD = "1234"
DEFAULT_START_DELAY_SEC = 3.0
DEFAULT_WINDOW_SEC = 1.0
DEFAULT_TIMEOUT_SEC = 60.0
DEFAULT_REASON_PREFIX = "benchmark-rejudge"
DEFAULT_POLL_INTERVAL_SEC = 1.0
DEFAULT_POLL_TIMEOUT_SEC = 60.0
DEFAULT_EXPERIMENT_TYPE = "rejudge_storm"


class ApiError(RuntimeError):
    pass


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Trigger a rejudge storm for benchmark experiments")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--api-prefix", default=DEFAULT_API_PREFIX)
    parser.add_argument("--output-dir", default=DEFAULT_OUTPUT_DIR)
    parser.add_argument("--contest-name", default=DEFAULT_CONTEST_NAME)
    parser.add_argument("--contest-project-name", default=DEFAULT_CONTEST_PROJECT_NAME)
    parser.add_argument("--project-id", type=int, default=None)
    parser.add_argument("--admin-username", default=DEFAULT_ADMIN_USERNAME)
    parser.add_argument("--admin-password", default=DEFAULT_ADMIN_PASSWORD)
    parser.add_argument("--submission-ids", default="")
    parser.add_argument("--usernames", default="")
    parser.add_argument("--limit", type=int, default=None)
    parser.add_argument("--start-delay-sec", type=float, default=DEFAULT_START_DELAY_SEC)
    parser.add_argument("--window-sec", type=float, default=DEFAULT_WINDOW_SEC)
    parser.add_argument("--schedule", choices=["spread", "burst"], default="burst")
    parser.add_argument("--reason-prefix", default=DEFAULT_REASON_PREFIX)
    parser.add_argument("--approve", choices=["auto", "none"], default="auto")
    parser.add_argument("--poll-interval-sec", type=float, default=DEFAULT_POLL_INTERVAL_SEC)
    parser.add_argument("--poll-timeout-sec", type=float, default=DEFAULT_POLL_TIMEOUT_SEC)
    parser.add_argument("--timeout-sec", type=float, default=DEFAULT_TIMEOUT_SEC)
    parser.add_argument("--experiment-type", default=DEFAULT_EXPERIMENT_TYPE)
    return parser


def api_url(base_url: str, api_prefix: str, path: str) -> str:
    return base_url.rstrip("/") + api_prefix + path


def json_request(opener: urllib.request.OpenerDirector, method: str, url: str, payload: dict[str, Any] | None = None, timeout_sec: float = DEFAULT_TIMEOUT_SEC, extra_headers: dict[str, str] | None = None) -> tuple[int, dict[str, Any]]:
    data = None
    headers = dict(extra_headers or {})
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with opener.open(request, timeout=timeout_sec) as response:
            status = response.status
            body = response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        status = exc.code
        body = exc.read().decode("utf-8")
    parsed = json.loads(body) if body else {}
    return status, parsed


def fetch_csrf_token(base_url: str, api_prefix: str, opener: urllib.request.OpenerDirector, timeout_sec: float) -> str:
    status, payload = json_request(opener, "GET", api_url(base_url, api_prefix, "/csrf-token"), None, timeout_sec)
    token = payload.get("data", {}).get("token")
    if status != 200 or not payload.get("success") or not token:
        raise ApiError(f"Failed to fetch CSRF token: status={status}, payload={payload}")
    return token


def login(base_url: str, api_prefix: str, username: str, password: str, timeout_sec: float) -> tuple[urllib.request.OpenerDirector, str]:
    cookie_jar = CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookie_jar))
    csrf_token = fetch_csrf_token(base_url, api_prefix, opener, timeout_sec)
    status, payload = json_request(
        opener,
        "POST",
        api_url(base_url, api_prefix, "/auth/login"),
        {"username": username, "password": password},
        timeout_sec,
        extra_headers={"csrf-token": csrf_token},
    )
    if status != 200 or not payload.get("success"):
        raise ApiError(f"Admin login failed: status={status}, payload={payload}")
    return opener, csrf_token


def discover_project_id(base_url: str, api_prefix: str, contest_name: str, contest_project_name: str, timeout_sec: float) -> int:
    opener = urllib.request.build_opener()
    status, contests_payload = json_request(opener, "GET", api_url(base_url, api_prefix, "/public/contests"), None, timeout_sec)
    if status != 200 or not contests_payload.get("success"):
        raise ApiError(f"Failed to list public contests: status={status}, payload={contests_payload}")
    contests = contests_payload.get("data", [])
    contest = next((item for item in contests if item.get("name") == contest_name), None)
    if contest is None:
        raise ApiError(f"Contest not found by name: {contest_name}")
    status, contest_payload = json_request(opener, "GET", api_url(base_url, api_prefix, f"/public/contests/{contest['id']}"), None, timeout_sec)
    if status != 200 or not contest_payload.get("success"):
        raise ApiError(f"Failed to get contest detail: status={status}, payload={contest_payload}")
    projects = contest_payload.get("data", {}).get("projects", [])
    project = next((item for item in projects if item.get("name") == contest_project_name), None)
    if project is None:
        raise ApiError(f"Contest project not found by name: {contest_project_name}")
    return int(project["id"])


def fetch_submissions(opener: urllib.request.OpenerDirector, base_url: str, api_prefix: str, project_id: int, timeout_sec: float) -> list[dict[str, Any]]:
    status, payload = json_request(opener, "GET", api_url(base_url, api_prefix, f"/contest-projects/{project_id}/submissions"), None, timeout_sec)
    if status != 200 or not payload.get("success"):
        raise ApiError(f"Failed to fetch submissions: status={status}, payload={payload}")
    return payload.get("data", [])


def fetch_rejudges(opener: urllib.request.OpenerDirector, base_url: str, api_prefix: str, timeout_sec: float) -> list[dict[str, Any]]:
    status, payload = json_request(opener, "GET", api_url(base_url, api_prefix, "/rejudges"), None, timeout_sec)
    if status != 200 or not payload.get("success"):
        raise ApiError(f"Failed to fetch rejudges: status={status}, payload={payload}")
    return payload.get("data", [])


def select_submission_targets(submissions: list[dict[str, Any]], submission_ids: str, usernames: str, limit: int | None) -> list[dict[str, Any]]:
    rows = list(submissions)
    if submission_ids:
        wanted = {item.strip() for item in submission_ids.split(",") if item.strip()}
        rows = [row for row in rows if row.get("id") in wanted]
        missing = sorted(wanted - {row.get("id") for row in rows})
        if missing:
            raise ApiError(f"Unknown submission ids: {', '.join(missing)}")
    if usernames:
        wanted_users = {item.strip() for item in usernames.split(",") if item.strip()}
        rows = [row for row in rows if row.get("user", {}).get("username") in wanted_users]
    rows.sort(key=lambda item: item.get("createdAt", ""), reverse=True)
    if limit is not None:
        rows = rows[:limit]
    rows.sort(key=lambda item: item.get("createdAt", ""))
    return rows


def build_schedule(count: int, start_delay_sec: float, window_sec: float, schedule: str) -> list[float]:
    start_at = time.time() + start_delay_sec
    if count <= 0:
        return []
    if schedule == "burst" or count == 1:
        return [start_at for _ in range(count)]
    interval = window_sec / max(count - 1, 1)
    return [start_at + index * interval for index in range(count)]


def trigger_rejudge(opener: urllib.request.OpenerDirector, csrf_token: str, base_url: str, api_prefix: str, submission_id: str, reason: str, scheduled_at: float, timeout_sec: float) -> dict[str, Any]:
    now = time.time()
    if scheduled_at > now:
        time.sleep(scheduled_at - now)
    request_sent_at = time.time()
    status, payload = json_request(
        opener,
        "POST",
        api_url(base_url, api_prefix, f"/submissions/{submission_id}/rejudge"),
        {"reason": reason},
        timeout_sec,
        extra_headers={"csrf-token": csrf_token},
    )
    response_received_at = time.time()
    return {
        "attempt_id": uuid.uuid4().hex,
        "kind": "rejudge_request",
        "submission_id": submission_id,
        "reason": reason,
        "scheduled_at": scheduled_at,
        "request_sent_at": request_sent_at,
        "response_received_at": response_received_at,
        "latency_ms": round((response_received_at - request_sent_at) * 1000, 3),
        "status_code": status,
        "success": status == 200 and bool(payload.get("success")),
        "response": payload,
    }


def approve_rejudge(opener: urllib.request.OpenerDirector, csrf_token: str, base_url: str, api_prefix: str, rejudge_id: str, timeout_sec: float) -> dict[str, Any]:
    request_sent_at = time.time()
    status, payload = json_request(
        opener,
        "POST",
        api_url(base_url, api_prefix, f"/rejudge/{rejudge_id}/approve"),
        None,
        timeout_sec,
        extra_headers={"csrf-token": csrf_token},
    )
    response_received_at = time.time()
    return {
        "attempt_id": uuid.uuid4().hex,
        "kind": "rejudge_approve",
        "rejudge_id": rejudge_id,
        "request_sent_at": request_sent_at,
        "response_received_at": response_received_at,
        "latency_ms": round((response_received_at - request_sent_at) * 1000, 3),
        "status_code": status,
        "success": status == 200 and bool(payload.get("success")),
        "response": payload,
    }


def write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    with path.open("w", encoding="utf-8") as file:
        for row in rows:
            file.write(json.dumps(row, ensure_ascii=False) + "\n")


def main() -> int:
    args = build_parser().parse_args()
    project_id = args.project_id or discover_project_id(args.base_url, args.api_prefix, args.contest_name, args.contest_project_name, args.timeout_sec)
    opener, csrf_token = login(args.base_url, args.api_prefix, args.admin_username, args.admin_password, args.timeout_sec)

    submissions = fetch_submissions(opener, args.base_url, args.api_prefix, project_id, args.timeout_sec)
    targets = select_submission_targets(submissions, args.submission_ids, args.usernames, args.limit)
    if not targets:
        raise SystemExit("No submissions selected for rejudge storm")

    timestamp = time.strftime("%Y%m%d-%H%M%S", time.localtime())
    output_dir = Path(args.output_dir) / f"{timestamp}-rejudge"
    output_dir.mkdir(parents=True, exist_ok=True)

    schedules = build_schedule(len(targets), args.start_delay_sec, args.window_sec, args.schedule)
    created_at_before = time.time()
    reason_prefix = f"{args.reason_prefix}-{timestamp}"

    run_manifest = {
        "generated_at": time.time(),
        "experiment_type": args.experiment_type,
        "base_url": args.base_url,
        "api_prefix": args.api_prefix,
        "project_id": project_id,
        "contest_name": args.contest_name,
        "contest_project_name": args.contest_project_name,
        "schedule": args.schedule,
        "window_sec": args.window_sec,
        "start_delay_sec": args.start_delay_sec,
        "approve": args.approve,
        "reason_prefix": reason_prefix,
        "target_submission_ids": [row.get("id") for row in targets],
        "target_usernames": sorted({row.get("user", {}).get("username") for row in targets if row.get("user", {}).get("username")}),
        "target_count": len(targets),
    }
    (output_dir / "run.json").write_text(json.dumps(run_manifest, ensure_ascii=False, indent=2), encoding="utf-8")

    request_results: list[dict[str, Any]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=len(targets)) as executor:
        future_map = {}
        for index, target in enumerate(targets):
            reason = f"{reason_prefix}-{index + 1:03d}"
            future = executor.submit(
                trigger_rejudge,
                opener,
                csrf_token,
                args.base_url,
                args.api_prefix,
                target["id"],
                reason,
                schedules[index],
                args.timeout_sec,
            )
            future_map[future] = (target, reason)
        for future in concurrent.futures.as_completed(future_map):
            target, reason = future_map[future]
            result = future.result()
            result["username"] = target.get("user", {}).get("username")
            result["target_submission_created_at"] = target.get("createdAt")
            result["reason"] = reason
            request_results.append(result)

    request_results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("submission_id", "")))
    write_jsonl(output_dir / "rejudge_attempts.jsonl", request_results)

    discovered_rejudges: list[dict[str, Any]] = []
    deadline = time.time() + args.poll_timeout_sec
    successful_requests = [row for row in request_results if row.get("success")]
    expected_count = len(successful_requests)
    while time.time() <= deadline:
        rejudges = fetch_rejudges(opener, args.base_url, args.api_prefix, args.timeout_sec)
        discovered_rejudges = [
            row for row in rejudges
            if str(row.get("reason", "")).startswith(reason_prefix)
            and parse_timestamp_safe(row.get("createdAt")) is not None
            and parse_timestamp_safe(row.get("createdAt")) >= created_at_before - 1
        ]
        if len(discovered_rejudges) >= expected_count:
            break
        time.sleep(args.poll_interval_sec)

    discovered_rejudges.sort(key=lambda item: (item.get("createdAt", ""), item.get("id", "")))
    (output_dir / "rejudges.json").write_text(json.dumps(discovered_rejudges, ensure_ascii=False, indent=2), encoding="utf-8")

    approve_results: list[dict[str, Any]] = []
    if args.approve == "auto":
        for row in discovered_rejudges:
            approve_results.append(approve_rejudge(opener, csrf_token, args.base_url, args.api_prefix, row["id"], args.timeout_sec))
        approve_results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("rejudge_id", "")))
        write_jsonl(output_dir / "rejudge_approve_attempts.jsonl", approve_results)

    print(json.dumps({
        "output_dir": str(output_dir),
        "experiment_type": args.experiment_type,
        "project_id": project_id,
        "requested_rejudges": len(request_results),
        "successful_rejudge_requests": len(successful_requests),
        "discovered_rejudges": len(discovered_rejudges),
        "approved_rejudges": sum(1 for row in approve_results if row.get("success")),
    }, ensure_ascii=False, indent=2))
    return 0


def parse_timestamp_safe(value: str | None) -> float | None:
    if not value:
        return None
    normalized = value.replace("Z", "+00:00")
    dt = time.strptime(normalized[:19], "%Y-%m-%dT%H:%M:%S")
    return time.mktime(dt)


if __name__ == "__main__":
    raise SystemExit(main())
