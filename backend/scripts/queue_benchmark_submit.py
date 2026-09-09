#!/usr/bin/env python3
"""Submit queue benchmark workloads against the contestant submission API.

This script:
1. Loads contestant accounts from the seeded users.json file.
2. Optionally discovers the seeded benchmark contest project ID by public names.
3. Schedules concurrent submissions within a configurable time window.
4. Records the created submission ID for each successful attempt when possible.
5. Writes a JSONL run log that can be consumed by queue_benchmark_analyze.py.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import io
import json
import threading
import time
import urllib.error
import urllib.request
import uuid
import zipfile
from dataclasses import dataclass
from http.cookiejar import CookieJar
from pathlib import Path
from typing import Any

DEFAULT_BASE_URL = "http://localhost:8080"
DEFAULT_API_PREFIX = "/api/v1"
DEFAULT_USERS_JSON = "database/seeder/seeders/users.json"
DEFAULT_OUTPUT_DIR = "scripts/output/queue-benchmark"
DEFAULT_CONTEST_NAME = "Queue Benchmark Contest"
DEFAULT_CONTEST_PROJECT_NAME = "Queue Benchmark Short Task"
DEFAULT_START_DELAY_SEC = 3.0
DEFAULT_WINDOW_SEC = 10.0
DEFAULT_TIMEOUT_SEC = 60.0
DEFAULT_SUBMISSIONS_PER_USER = 1
DEFAULT_SCHEDULE_MODE = "spread"
DEFAULT_PROJECT_FILENAME = "benchmark_submission.zip"
DEFAULT_EXPERIMENT_TYPE = "submission_benchmark"
DEFAULT_SUBMISSION_POLL_SEC = 0.5
DEFAULT_SUBMISSION_POLL_TIMEOUT_SEC = 20.0


@dataclass
class AttemptPlan:
    attempt_id: str
    username: str
    password: str
    scheduled_at: float
    sequence_for_user: int


class ApiError(RuntimeError):
    pass


_USER_LOCKS: dict[str, threading.Lock] = {}
_USER_LOCKS_GUARD = threading.Lock()


def user_lock(username: str) -> threading.Lock:
    with _USER_LOCKS_GUARD:
        lock = _USER_LOCKS.get(username)
        if lock is None:
            lock = threading.Lock()
            _USER_LOCKS[username] = lock
        return lock


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Run queue benchmark submission workload")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--api-prefix", default=DEFAULT_API_PREFIX)
    parser.add_argument("--users-json", default=DEFAULT_USERS_JSON)
    parser.add_argument("--output-dir", default=DEFAULT_OUTPUT_DIR)
    parser.add_argument("--contest-name", default=DEFAULT_CONTEST_NAME)
    parser.add_argument("--contest-project-name", default=DEFAULT_CONTEST_PROJECT_NAME)
    parser.add_argument("--project-id", type=int, default=None)
    parser.add_argument("--project-file", default=None)
    parser.add_argument("--project-filename", default=DEFAULT_PROJECT_FILENAME)
    parser.add_argument("--user-count", type=int, default=None)
    parser.add_argument("--usernames", default="")
    parser.add_argument("--extra-usernames", default="")
    parser.add_argument("--submissions-per-user", type=int, default=DEFAULT_SUBMISSIONS_PER_USER)
    parser.add_argument("--start-delay-sec", type=float, default=DEFAULT_START_DELAY_SEC)
    parser.add_argument("--window-sec", type=float, default=DEFAULT_WINDOW_SEC)
    parser.add_argument("--schedule", choices=["spread", "burst"], default=DEFAULT_SCHEDULE_MODE)
    parser.add_argument("--timeout-sec", type=float, default=DEFAULT_TIMEOUT_SEC)
    parser.add_argument("--submission-poll-sec", type=float, default=DEFAULT_SUBMISSION_POLL_SEC)
    parser.add_argument("--submission-poll-timeout-sec", type=float, default=DEFAULT_SUBMISSION_POLL_TIMEOUT_SEC)
    parser.add_argument("--experiment-type", default=DEFAULT_EXPERIMENT_TYPE)
    return parser


def api_url(base_url: str, api_prefix: str, path: str) -> str:
    return base_url.rstrip("/") + api_prefix + path


def read_users(users_json: str) -> list[dict[str, Any]]:
    users = json.loads(Path(users_json).read_text(encoding="utf-8"))
    users.sort(key=lambda item: item["username"])
    return users


def select_users(users: list[dict[str, Any]], usernames: str, user_count: int | None) -> list[dict[str, Any]]:
    if usernames:
        wanted = {name.strip() for name in usernames.split(",") if name.strip()}
        selected = [user for user in users if user["username"] in wanted]
        missing = sorted(wanted - {user["username"] for user in selected})
        if missing:
            raise ApiError(f"Unknown usernames in users.json: {', '.join(missing)}")
        selected.sort(key=lambda item: item["username"])
        return selected

    if user_count is None:
        return users
    return users[:user_count]


def create_project_payload(project_file: str | None, project_filename: str) -> tuple[str, bytes]:
    if project_file:
        path = Path(project_file)
        return path.name, path.read_bytes()

    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, mode="w", compression=zipfile.ZIP_DEFLATED) as zf:
        zf.writestr("README.txt", "Queue benchmark placeholder submission\n")
    return project_filename, buffer.getvalue()


def encode_multipart_formdata(fields: dict[str, str], files: list[tuple[str, str, bytes, str]]) -> tuple[bytes, str]:
    boundary = f"----QueueBenchmarkBoundary{uuid.uuid4().hex}"
    lines: list[bytes] = []

    for name, value in fields.items():
        lines.extend(
            [
                f"--{boundary}".encode(),
                f'Content-Disposition: form-data; name="{name}"'.encode(),
                b"",
                value.encode(),
            ]
        )

    for field_name, filename, payload, content_type in files:
        lines.extend(
            [
                f"--{boundary}".encode(),
                f'Content-Disposition: form-data; name="{field_name}"; filename="{filename}"'.encode(),
                f"Content-Type: {content_type}".encode(),
                b"",
                payload,
            ]
        )

    lines.append(f"--{boundary}--".encode())
    lines.append(b"")
    body = b"\r\n".join(lines)
    return body, f"multipart/form-data; boundary={boundary}"


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
        raise ApiError(f"Login failed for {username}: status={status}, payload={payload}")
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

    status, contest_payload = json_request(
        opener,
        "GET",
        api_url(base_url, api_prefix, f"/public/contests/{contest['id']}"),
        None,
        timeout_sec,
    )
    if status != 200 or not contest_payload.get("success"):
        raise ApiError(f"Failed to get contest detail: status={status}, payload={contest_payload}")

    projects = contest_payload.get("data", {}).get("projects", [])
    project = next((item for item in projects if item.get("name") == contest_project_name), None)
    if project is None:
        raise ApiError(f"Contest project not found by name: {contest_project_name}")
    return int(project["id"])


def fetch_user_project_submissions(opener: urllib.request.OpenerDirector, base_url: str, api_prefix: str, project_id: int, timeout_sec: float) -> list[dict[str, Any]]:
    status, payload = json_request(opener, "GET", api_url(base_url, api_prefix, f"/user/contest-projects/{project_id}/submissions"), None, timeout_sec)
    if status != 200 or not payload.get("success"):
        raise ApiError(f"Failed to list user project submissions: status={status}, payload={payload}")
    return payload.get("data", [])


def resolve_created_submission(
    opener: urllib.request.OpenerDirector,
    base_url: str,
    api_prefix: str,
    project_id: int,
    timeout_sec: float,
    request_sent_at: float,
    before_ids: set[str],
    poll_sec: float,
    poll_timeout_sec: float,
) -> tuple[str | None, str | None]:
    deadline = time.time() + poll_timeout_sec
    best_known_id = None
    best_known_created_at = None

    while time.time() <= deadline:
        submissions = fetch_user_project_submissions(opener, base_url, api_prefix, project_id, timeout_sec)
        candidates = [row for row in submissions if row.get("id") not in before_ids]
        if candidates:
            candidates.sort(key=lambda item: item.get("createdAt", ""), reverse=True)
            best = candidates[0]
            return best.get("id"), best.get("createdAt")

        time.sleep(poll_sec)

        # fallback in case list endpoint is eventually consistent but old ids set changed unexpectedly
        submissions = fetch_user_project_submissions(opener, base_url, api_prefix, project_id, timeout_sec)
        candidates = []
        for row in submissions:
            created_at = row.get("createdAt")
            if row.get("id") in before_ids or not created_at:
                continue
            best_known_id = row.get("id")
            best_known_created_at = created_at
            candidates.append(row)
        if candidates:
            candidates.sort(key=lambda item: item.get("createdAt", ""), reverse=True)
            best = candidates[0]
            return best.get("id"), best.get("createdAt")

    return best_known_id, best_known_created_at


def build_attempts(users: list[dict[str, Any]], submissions_per_user: int, extra_usernames: str, start_delay_sec: float, window_sec: float, schedule: str) -> list[AttemptPlan]:
    start_at = time.time() + start_delay_sec
    attempts: list[AttemptPlan] = []
    per_user_sequence: dict[str, int] = {}

    for user in users:
        per_user_sequence[user["username"]] = 0
        for _ in range(submissions_per_user):
            per_user_sequence[user["username"]] += 1
            attempts.append(
                AttemptPlan(
                    attempt_id=uuid.uuid4().hex,
                    username=user["username"],
                    password=user["password"],
                    scheduled_at=0.0,
                    sequence_for_user=per_user_sequence[user["username"]],
                )
            )

    if extra_usernames:
        password_map = {user["username"]: user["password"] for user in users}
        for username in [item.strip() for item in extra_usernames.split(",") if item.strip()]:
            if username not in password_map:
                raise ApiError(f"Extra username not selected in benchmark set: {username}")
            per_user_sequence.setdefault(username, submissions_per_user)
            per_user_sequence[username] += 1
            attempts.append(
                AttemptPlan(
                    attempt_id=uuid.uuid4().hex,
                    username=username,
                    password=password_map[username],
                    scheduled_at=0.0,
                    sequence_for_user=per_user_sequence[username],
                )
            )

    attempts.sort(key=lambda item: (item.username, item.sequence_for_user, item.attempt_id))
    if not attempts:
        return attempts

    if schedule == "burst" or len(attempts) == 1:
        for attempt in attempts:
            attempt.scheduled_at = start_at
        return attempts

    interval = window_sec / max(len(attempts) - 1, 1)
    for index, attempt in enumerate(attempts):
        attempt.scheduled_at = start_at + index * interval
    return attempts


def submit_once(
    plan: AttemptPlan,
    base_url: str,
    api_prefix: str,
    project_id: int,
    payload_filename: str,
    payload_bytes: bytes,
    timeout_sec: float,
    submission_poll_sec: float,
    submission_poll_timeout_sec: float,
) -> dict[str, Any]:
    opener, csrf_token = login(base_url, api_prefix, plan.username, plan.password, timeout_sec)
    now = time.time()
    if plan.scheduled_at > now:
        time.sleep(plan.scheduled_at - now)

    with user_lock(plan.username):
        before_submissions = fetch_user_project_submissions(opener, base_url, api_prefix, project_id, timeout_sec)
        before_ids = {row.get("id") for row in before_submissions if row.get("id")}

        body, content_type = encode_multipart_formdata({}, [("project", payload_filename, payload_bytes, "application/zip")])
        request = urllib.request.Request(
            api_url(base_url, api_prefix, f"/user/contest-projects/{project_id}/submit"),
            data=body,
            headers={"Content-Type": content_type, "csrf-token": csrf_token},
            method="POST",
        )

        request_sent_at = time.time()
        status_code = None
        response_payload: dict[str, Any] = {}
        response_body = ""
        success = False

        try:
            with opener.open(request, timeout=timeout_sec) as response:
                status_code = response.status
                response_body = response.read().decode("utf-8")
        except urllib.error.HTTPError as exc:
            status_code = exc.code
            response_body = exc.read().decode("utf-8")
        except Exception as exc:
            response_received_at = time.time()
            return {
                "attempt_id": plan.attempt_id,
                "kind": "submission",
                "username": plan.username,
                "sequence_for_user": plan.sequence_for_user,
                "scheduled_at": plan.scheduled_at,
                "request_sent_at": request_sent_at,
                "response_received_at": response_received_at,
                "latency_ms": round((response_received_at - request_sent_at) * 1000, 3),
                "status_code": None,
                "success": False,
                "error": repr(exc),
            }

        response_received_at = time.time()
        if response_body:
            try:
                response_payload = json.loads(response_body)
            except json.JSONDecodeError:
                response_payload = {"raw": response_body}
        success = status_code == 200 and bool(response_payload.get("success"))

        submission_id = None
        submission_created_at = None
        if success:
            submission_id, submission_created_at = resolve_created_submission(
                opener,
                base_url,
                api_prefix,
                project_id,
                timeout_sec,
                request_sent_at,
                before_ids,
                submission_poll_sec,
                submission_poll_timeout_sec,
            )

    return {
        "attempt_id": plan.attempt_id,
        "kind": "submission",
        "username": plan.username,
        "sequence_for_user": plan.sequence_for_user,
        "scheduled_at": plan.scheduled_at,
        "request_sent_at": request_sent_at,
        "response_received_at": response_received_at,
        "latency_ms": round((response_received_at - request_sent_at) * 1000, 3),
        "status_code": status_code,
        "success": success,
        "submission_id": submission_id,
        "submission_created_at": submission_created_at,
        "response": response_payload,
    }


def write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    with path.open("w", encoding="utf-8") as file:
        for row in rows:
            file.write(json.dumps(row, ensure_ascii=False) + "\n")


def main() -> int:
    args = build_parser().parse_args()

    users = read_users(args.users_json)
    selected_users = select_users(users, args.usernames, args.user_count)
    attempts = build_attempts(
        users=selected_users,
        submissions_per_user=args.submissions_per_user,
        extra_usernames=args.extra_usernames,
        start_delay_sec=args.start_delay_sec,
        window_sec=args.window_sec,
        schedule=args.schedule,
    )
    if not attempts:
        raise SystemExit("No benchmark attempts to run. Check --user-count or --usernames.")

    project_id = args.project_id or discover_project_id(
        args.base_url,
        args.api_prefix,
        args.contest_name,
        args.contest_project_name,
        args.timeout_sec,
    )
    payload_filename, payload_bytes = create_project_payload(args.project_file, args.project_filename)

    timestamp = time.strftime("%Y%m%d-%H%M%S", time.localtime())
    output_dir = Path(args.output_dir) / timestamp
    output_dir.mkdir(parents=True, exist_ok=True)

    per_user_attempt_count: dict[str, int] = {}
    for attempt in attempts:
        per_user_attempt_count[attempt.username] = per_user_attempt_count.get(attempt.username, 0) + 1
    heavy_users = sorted([username for username, count in per_user_attempt_count.items() if count > 1])
    light_users = sorted([username for username, count in per_user_attempt_count.items() if count == 1])

    run_manifest = {
        "generated_at": time.time(),
        "experiment_type": args.experiment_type,
        "base_url": args.base_url,
        "api_prefix": args.api_prefix,
        "contest_name": args.contest_name,
        "contest_project_name": args.contest_project_name,
        "project_id": project_id,
        "schedule": args.schedule,
        "window_sec": args.window_sec,
        "start_delay_sec": args.start_delay_sec,
        "submissions_per_user": args.submissions_per_user,
        "selected_usernames": [user["username"] for user in selected_users],
        "extra_usernames": [item.strip() for item in args.extra_usernames.split(",") if item.strip()],
        "per_user_attempt_count": per_user_attempt_count,
        "heavy_users": heavy_users,
        "light_users": light_users,
        "attempt_count": len(attempts),
        "project_payload_filename": payload_filename,
    }
    (output_dir / "run.json").write_text(json.dumps(run_manifest, ensure_ascii=False, indent=2), encoding="utf-8")

    results: list[dict[str, Any]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=len(attempts)) as executor:
        future_map = {
            executor.submit(
                submit_once,
                plan,
                args.base_url,
                args.api_prefix,
                project_id,
                payload_filename,
                payload_bytes,
                args.timeout_sec,
                args.submission_poll_sec,
                args.submission_poll_timeout_sec,
            ): plan
            for plan in attempts
        }
        for future in concurrent.futures.as_completed(future_map):
            results.append(future.result())

    results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item["username"], item["attempt_id"]))
    write_jsonl(output_dir / "attempts.jsonl", results)

    success_count = sum(1 for item in results if item.get("success"))
    resolved_submission_count = sum(1 for item in results if item.get("submission_id"))
    print(json.dumps({
        "output_dir": str(output_dir),
        "experiment_type": args.experiment_type,
        "project_id": project_id,
        "attempt_count": len(results),
        "success_count": success_count,
        "resolved_submission_count": resolved_submission_count,
        "failure_count": len(results) - success_count,
        "heavy_users": heavy_users,
        "light_users": light_users,
    }, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
