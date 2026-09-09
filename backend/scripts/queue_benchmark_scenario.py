#!/usr/bin/env python3
"""Run state-driven queue benchmark scenarios.

A scenario is a JSON document that combines submit steps, rejudge steps, and
wait conditions. This is useful when the next action should be triggered only
after the system reaches a specific judging or rejudge state.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import subprocess
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any

SCRIPT_DIR = Path(__file__).resolve().parent
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

import queue_benchmark_analyze as analyze_mod
import queue_benchmark_rejudge as rejudge_mod
import queue_benchmark_submit as submit_mod

DEFAULT_OUTPUT_DIR = "scripts/output/queue-benchmark-scenarios"
DEFAULT_SCENARIO = "scripts/scenarios/priority_order_validation.json"
DEFAULT_POLL_INTERVAL_SEC = 2.0
DEFAULT_ANALYZE_UNTIL = "finished"


@dataclass
class ScenarioAttempt:
    attempt_id: str
    username: str
    password: str
    sequence_for_user: int
    step_index: int
    step_name: str
    scheduled_at: float


class ScenarioError(RuntimeError):
    pass


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Run a state-driven queue benchmark scenario")
    parser.add_argument("--scenario", default=DEFAULT_SCENARIO)
    parser.add_argument("--output-dir", default=DEFAULT_OUTPUT_DIR)
    parser.add_argument("--base-url", default=submit_mod.DEFAULT_BASE_URL)
    parser.add_argument("--api-prefix", default=submit_mod.DEFAULT_API_PREFIX)
    parser.add_argument("--users-json", default=submit_mod.DEFAULT_USERS_JSON)
    parser.add_argument("--contest-name", default=submit_mod.DEFAULT_CONTEST_NAME)
    parser.add_argument("--contest-project-name", default=submit_mod.DEFAULT_CONTEST_PROJECT_NAME)
    parser.add_argument("--project-id", type=int, default=None)
    parser.add_argument("--project-file", default=None)
    parser.add_argument("--project-filename", default=submit_mod.DEFAULT_PROJECT_FILENAME)
    parser.add_argument("--timeout-sec", type=float, default=submit_mod.DEFAULT_TIMEOUT_SEC)
    parser.add_argument("--submission-poll-sec", type=float, default=submit_mod.DEFAULT_SUBMISSION_POLL_SEC)
    parser.add_argument("--submission-poll-timeout-sec", type=float, default=submit_mod.DEFAULT_SUBMISSION_POLL_TIMEOUT_SEC)
    parser.add_argument("--admin-username", default=analyze_mod.DEFAULT_ADMIN_USERNAME)
    parser.add_argument("--admin-password", default=analyze_mod.DEFAULT_ADMIN_PASSWORD)
    parser.add_argument("--poll-interval-sec", type=float, default=DEFAULT_POLL_INTERVAL_SEC)
    parser.add_argument("--analyze-until", choices=["started", "finished"], default=DEFAULT_ANALYZE_UNTIL)
    return parser


def load_scenario(path: str) -> dict[str, Any]:
    return json.loads(Path(path).read_text(encoding="utf-8"))


def build_user_map(users_json: str) -> dict[str, dict[str, Any]]:
    users = submit_mod.read_users(users_json)
    return {user["username"]: user for user in users}


def discover_project_id(args: argparse.Namespace, scenario: dict[str, Any]) -> int:
    if args.project_id:
        return args.project_id
    if scenario.get("project_id"):
        return int(scenario["project_id"])
    return submit_mod.discover_project_id(
        args.base_url,
        args.api_prefix,
        scenario.get("contest_name", args.contest_name),
        scenario.get("contest_project_name", args.contest_project_name),
        args.timeout_sec,
    )


def create_output_dir(base_dir: str) -> Path:
    timestamp = time.strftime("%Y%m%d-%H%M%S", time.localtime())
    path = Path(base_dir) / timestamp
    path.mkdir(parents=True, exist_ok=True)
    return path


def make_attempt(username: str, password: str, sequence_for_user: int, step_index: int, step_name: str, scheduled_at: float) -> ScenarioAttempt:
    return ScenarioAttempt(
        attempt_id=submit_mod.uuid.uuid4().hex,
        username=username,
        password=password,
        sequence_for_user=sequence_for_user,
        step_index=step_index,
        step_name=step_name,
        scheduled_at=scheduled_at,
    )


def run_submit_batch(
    step: dict[str, Any],
    step_index: int,
    user_map: dict[str, dict[str, Any]],
    sequence_tracker: dict[str, int],
    base_url: str,
    api_prefix: str,
    project_id: int,
    payload_filename: str,
    payload_bytes: bytes,
    timeout_sec: float,
    submission_poll_sec: float,
    submission_poll_timeout_sec: float,
) -> list[dict[str, Any]]:
    usernames = step.get("usernames") or []
    if not usernames:
        raise ScenarioError(f"submit_batch step {step_index} has no usernames")

    schedule = step.get("schedule", "sequential")
    interval_sec = float(step.get("interval_sec", 0.0))
    window_sec = float(step.get("window_sec", 0.0))
    start_delay_sec = float(step.get("start_delay_sec", 0.0))
    step_name = step.get("name", f"step_{step_index}")

    now = time.time() + start_delay_sec
    attempts: list[ScenarioAttempt] = []
    for offset, username in enumerate(usernames):
        if username not in user_map:
            raise ScenarioError(f"Unknown username in scenario: {username}")
        sequence_tracker[username] = sequence_tracker.get(username, 0) + 1
        scheduled_at = now
        if schedule == "sequential":
            scheduled_at = now + offset * interval_sec
        attempts.append(make_attempt(
            username=username,
            password=user_map[username]["password"],
            sequence_for_user=sequence_tracker[username],
            step_index=step_index,
            step_name=step_name,
            scheduled_at=scheduled_at,
        ))

    if schedule in {"burst", "spread"}:
        plans = [submit_mod.AttemptPlan(a.attempt_id, a.username, a.password, 0.0, a.sequence_for_user) for a in attempts]
        if schedule == "burst" or len(plans) == 1:
            for plan in plans:
                plan.scheduled_at = now
        else:
            interval = window_sec / max(len(plans) - 1, 1)
            for idx, plan in enumerate(plans):
                plan.scheduled_at = now + idx * interval
        for attempt, plan in zip(attempts, plans):
            attempt.scheduled_at = plan.scheduled_at

    results: list[dict[str, Any]] = []
    max_workers = len(attempts) if schedule in {"burst", "spread"} else 1
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        future_map = {
            executor.submit(
                submit_mod.submit_once,
                submit_mod.AttemptPlan(a.attempt_id, a.username, a.password, a.scheduled_at, a.sequence_for_user),
                base_url,
                api_prefix,
                project_id,
                payload_filename,
                payload_bytes,
                timeout_sec,
                submission_poll_sec,
                submission_poll_timeout_sec,
            ): a
            for a in attempts
        }
        for future in concurrent.futures.as_completed(future_map):
            attempt = future_map[future]
            result = future.result()
            result["step_index"] = attempt.step_index
            result["step_name"] = attempt.step_name
            results.append(result)

    results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("username", ""), item.get("attempt_id", "")))
    return results


def fetch_submission_status(admin_opener: Any, args: argparse.Namespace, submission_id: str) -> dict[str, Any]:
    return analyze_mod.fetch_submission_detail(admin_opener, args.base_url, args.api_prefix, submission_id, args.timeout_sec)


def fetch_rejudge_status(admin_opener: Any, args: argparse.Namespace, rejudge_id: str) -> dict[str, Any]:
    status, payload = rejudge_mod.json_request(
        admin_opener,
        "GET",
        rejudge_mod.api_url(args.base_url, args.api_prefix, f"/rejudges/{rejudge_id}"),
        None,
        args.timeout_sec,
    )
    if status != 200 or not payload.get("success"):
        raise ScenarioError(f"Failed to fetch rejudge detail: status={status}, payload={payload}")
    return payload.get("data", {})


def resolve_submission_attempt(attempts: list[dict[str, Any]], target: dict[str, Any]) -> dict[str, Any]:
    submission_id = target.get("submission_id")
    if submission_id:
        row = next((row for row in attempts if row.get("submission_id") == submission_id), None)
        if row is None:
            raise ScenarioError(f"Unknown submission_id in scenario target: {submission_id}")
        return row

    username = target.get("username")
    sequence_for_user = int(target.get("sequence_for_user", 0))
    row = next(
        (
            item for item in attempts
            if item.get("username") == username and item.get("sequence_for_user") == sequence_for_user and item.get("success") and item.get("submission_id")
        ),
        None,
    )
    if row is None:
        raise ScenarioError(f"Could not resolve submission target for username={username} sequence_for_user={sequence_for_user}")
    return row


def discover_rejudges_by_reason_prefix(admin_opener: Any, args: argparse.Namespace, reason_prefix: str, created_at_before: float, expected_count: int, poll_interval_sec: float, poll_timeout_sec: float) -> list[dict[str, Any]]:
    deadline = time.time() + poll_timeout_sec
    discovered: list[dict[str, Any]] = []
    while time.time() <= deadline:
        rows = rejudge_mod.fetch_rejudges(admin_opener, args.base_url, args.api_prefix, args.timeout_sec)
        discovered = [
            row for row in rows
            if str(row.get("reason", "")).startswith(reason_prefix)
            and rejudge_mod.parse_timestamp_safe(row.get("createdAt")) is not None
            and rejudge_mod.parse_timestamp_safe(row.get("createdAt")) >= created_at_before - 1
        ]
        if len(discovered) >= expected_count:
            break
        time.sleep(poll_interval_sec)
    discovered.sort(key=lambda item: (item.get("createdAt", ""), item.get("id", "")))
    return discovered


def approve_rejudges(admin_opener: Any, csrf_token: str, args: argparse.Namespace, rejudges: list[dict[str, Any]]) -> list[dict[str, Any]]:
    approve_results: list[dict[str, Any]] = []
    for row in rejudges:
        result = rejudge_mod.approve_rejudge(admin_opener, csrf_token, args.base_url, args.api_prefix, row["id"], args.timeout_sec)
        result["step_rejudge_reason"] = row.get("reason")
        approve_results.append(result)
    approve_results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("rejudge_id", "")))
    return approve_results


def trigger_contest_project_rejudge(admin_opener: Any, csrf_token: str, args: argparse.Namespace, project_id: int, reason: str, grading_item_ids: list[int] | None) -> dict[str, Any]:
    request_sent_at = time.time()
    payload: dict[str, Any] = {"reason": reason}
    if grading_item_ids is not None:
        payload["gradingItemIds"] = grading_item_ids
    status, response = rejudge_mod.json_request(
        admin_opener,
        "POST",
        rejudge_mod.api_url(args.base_url, args.api_prefix, f"/contest-projects/{project_id}/rejudge"),
        payload,
        args.timeout_sec,
        extra_headers={"csrf-token": csrf_token},
    )
    response_received_at = time.time()
    return {
        "attempt_id": submit_mod.uuid.uuid4().hex,
        "kind": "contest_project_rejudge_request",
        "contest_project_id": project_id,
        "reason": reason,
        "grading_item_ids": grading_item_ids,
        "request_sent_at": request_sent_at,
        "response_received_at": response_received_at,
        "latency_ms": round((response_received_at - request_sent_at) * 1000, 3),
        "status_code": status,
        "success": status == 200 and bool(response.get("success")),
        "response": response,
    }


def run_rejudge_submissions(step: dict[str, Any], step_index: int, step_name: str, attempts: list[dict[str, Any]], admin_opener: Any, csrf_token: str, args: argparse.Namespace) -> dict[str, Any]:
    targets = step.get("targets") or []
    if not targets:
        raise ScenarioError(f"rejudge_submissions step {step_index} has no targets")

    resolved_targets = [resolve_submission_attempt(attempts, target) for target in targets]
    reason_prefix = step.get("reason_prefix") or f"scenario-rejudge-submissions-{int(time.time())}"
    created_at_before = time.time()
    start_delay_sec = float(step.get("start_delay_sec", 0.0))
    window_sec = float(step.get("window_sec", 0.0))
    schedule = step.get("schedule", "burst")
    approve_mode = step.get("approve", "none")
    poll_interval_sec = float(step.get("poll_interval_sec", args.poll_interval_sec))
    poll_timeout_sec = float(step.get("poll_timeout_sec", 60.0))

    schedules = rejudge_mod.build_schedule(len(resolved_targets), start_delay_sec, window_sec, schedule)
    request_results: list[dict[str, Any]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=max(1, len(resolved_targets))) as executor:
        future_map = {}
        for index, target in enumerate(resolved_targets):
            reason = f"{reason_prefix}-{index + 1:03d}"
            future = executor.submit(
                rejudge_mod.trigger_rejudge,
                admin_opener,
                csrf_token,
                args.base_url,
                args.api_prefix,
                target["submission_id"],
                reason,
                schedules[index],
                args.timeout_sec,
            )
            future_map[future] = (target, reason)
        for future in concurrent.futures.as_completed(future_map):
            target, reason = future_map[future]
            result = future.result()
            result["username"] = target.get("username")
            result["sequence_for_user"] = target.get("sequence_for_user")
            result["target_submission_id"] = target.get("submission_id")
            result["target_submission_created_at"] = target.get("submission_created_at")
            result["step_index"] = step_index
            result["step_name"] = step_name
            result["reason"] = reason
            request_results.append(result)
    request_results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("target_submission_id", "")))

    successful_requests = [row for row in request_results if row.get("success")]
    discovered_rejudges = discover_rejudges_by_reason_prefix(
        admin_opener,
        args,
        reason_prefix,
        created_at_before,
        len(successful_requests),
        poll_interval_sec,
        poll_timeout_sec,
    )

    approve_results: list[dict[str, Any]] = []
    if approve_mode == "auto":
        approve_results = approve_rejudges(admin_opener, csrf_token, args, discovered_rejudges)

    return {
        "step_index": step_index,
        "step_name": step_name,
        "type": "rejudge_submissions",
        "reason_prefix": reason_prefix,
        "request_results": request_results,
        "discovered_rejudges": discovered_rejudges,
        "approve_results": approve_results,
    }


def run_rejudge_contest_project(step: dict[str, Any], step_index: int, step_name: str, admin_opener: Any, csrf_token: str, args: argparse.Namespace, project_id: int) -> dict[str, Any]:
    reason = step.get("reason") or f"scenario-rejudge-contest-project-{int(time.time())}"
    grading_item_ids = step.get("grading_item_ids")
    approve_mode = step.get("approve", "none")
    poll_interval_sec = float(step.get("poll_interval_sec", args.poll_interval_sec))
    poll_timeout_sec = float(step.get("poll_timeout_sec", 60.0))
    created_at_before = time.time()

    request_result = trigger_contest_project_rejudge(admin_opener, csrf_token, args, project_id, reason, grading_item_ids)
    discovered_rejudges = discover_rejudges_by_reason_prefix(
        admin_opener,
        args,
        reason,
        created_at_before,
        1 if request_result.get("success") else 0,
        poll_interval_sec,
        poll_timeout_sec,
    )

    approve_results: list[dict[str, Any]] = []
    if approve_mode == "auto":
        approve_results = approve_rejudges(admin_opener, csrf_token, args, discovered_rejudges)

    return {
        "step_index": step_index,
        "step_name": step_name,
        "type": "rejudge_contest_project",
        "reason_prefix": reason,
        "request_results": [request_result],
        "discovered_rejudges": discovered_rejudges,
        "approve_results": approve_results,
    }


def run_approve_rejudge(step: dict[str, Any], step_index: int, step_name: str, admin_opener: Any, csrf_token: str, args: argparse.Namespace, rejudge_registry: dict[str, list[dict[str, Any]]]) -> dict[str, Any]:
    source_step = step.get("source_step")
    if not source_step:
        raise ScenarioError(f"approve_rejudge step {step_index} requires source_step")
    rejudges = list(rejudge_registry.get(source_step, []))
    if not rejudges:
        raise ScenarioError(f"approve_rejudge step {step_index} could not find rejudges from step {source_step}")
    approve_results = approve_rejudges(admin_opener, csrf_token, args, rejudges)
    return {
        "step_index": step_index,
        "step_name": step_name,
        "type": "approve_rejudge",
        "source_step": source_step,
        "approve_results": approve_results,
    }


def condition_matches(detail: dict[str, Any], condition: dict[str, Any]) -> bool:
    judgments = detail.get("judgments", [])
    expected_status = condition.get("judgment_status")
    if expected_status:
        return any(judgment.get("status") == expected_status for judgment in judgments)
    if condition.get("judging_started") is True:
        return any(j.get("judgingStartedAt") for j in judgments)
    if condition.get("judging_finished") is True:
        return any(j.get("judgingFinishedAt") for j in judgments)
    return False


def rejudge_condition_matches(detail: dict[str, Any], condition: dict[str, Any]) -> bool:
    expected_status = condition.get("rejudge_status")
    if expected_status:
        return detail.get("status") == expected_status
    return False


def run_wait_until_submission(step: dict[str, Any], attempts: list[dict[str, Any]], admin_opener: Any, args: argparse.Namespace) -> dict[str, Any]:
    username = step.get("username")
    sequence_for_user = int(step.get("sequence_for_user", 0))
    timeout_sec = float(step.get("timeout_sec", 180.0))
    poll_interval_sec = float(step.get("poll_interval_sec", args.poll_interval_sec))
    deadline = time.time() + timeout_sec

    while time.time() <= deadline:
        target = next(
            (
                row for row in attempts
                if row.get("username") == username and row.get("sequence_for_user") == sequence_for_user and row.get("success") and row.get("submission_id")
            ),
            None,
        )
        if target is not None:
            detail = fetch_submission_status(admin_opener, args, target["submission_id"])
            if condition_matches(detail, step):
                return {
                    "wait_condition": {
                        "username": username,
                        "sequence_for_user": sequence_for_user,
                        "submission_id": target["submission_id"],
                    },
                    "matched_submission_detail": detail,
                }
        time.sleep(poll_interval_sec)

    raise ScenarioError(f"Timed out waiting for condition: {step}")


def run_wait_until_rejudge(step: dict[str, Any], rejudge_registry: dict[str, list[dict[str, Any]]], admin_opener: Any, args: argparse.Namespace) -> dict[str, Any]:
    source_step = step.get("source_step")
    timeout_sec = float(step.get("timeout_sec", 180.0))
    poll_interval_sec = float(step.get("poll_interval_sec", args.poll_interval_sec))
    deadline = time.time() + timeout_sec
    expected_status = step.get("rejudge_status")
    if not source_step:
        raise ScenarioError(f"wait_until_rejudge requires source_step: {step}")

    while time.time() <= deadline:
        rejudges = rejudge_registry.get(source_step, [])
        if rejudges:
            details = [fetch_rejudge_status(admin_opener, args, row["id"]) for row in rejudges]
            if expected_status:
                if all(rejudge_condition_matches(detail, step) for detail in details):
                    return {
                        "wait_condition": {
                            "source_step": source_step,
                            "rejudge_status": expected_status,
                            "rejudge_ids": [row.get("id") for row in rejudges],
                        },
                        "matched_rejudge_details": details,
                    }
            else:
                return {
                    "wait_condition": {
                        "source_step": source_step,
                        "rejudge_ids": [row.get("id") for row in rejudges],
                    },
                    "matched_rejudge_details": details,
                }
        time.sleep(poll_interval_sec)

    raise ScenarioError(f"Timed out waiting for rejudge condition: {step}")


def run_analyze(attempts_path: Path) -> dict[str, Any]:
    cmd = [sys.executable, str(SCRIPT_DIR / "queue_benchmark_analyze.py"), "--run-log", str(attempts_path)]
    completed = subprocess.run(cmd, capture_output=True, text=True)
    if completed.returncode != 0:
        raise ScenarioError(f"Analyze failed\nstdout:\n{completed.stdout}\nstderr:\n{completed.stderr}")
    return json.loads(completed.stdout)


def load_analysis(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def build_order_analysis(analysis: dict[str, Any], scenario: dict[str, Any]) -> dict[str, Any]:
    actual_records = [row for row in analysis.get("records", []) if row.get("judging_start_rank")]
    actual_records.sort(key=lambda row: row["judging_start_rank"])
    actual_order = [f"{row['username']}#{row.get('sequence_for_user')}" for row in actual_records]
    expected_order = scenario.get("expected_judging_order", [])

    first_mismatch = None
    for idx, expected in enumerate(expected_order):
        if idx >= len(actual_order):
            first_mismatch = {"index": idx + 1, "expected": expected, "actual": None}
            break
        if actual_order[idx] != expected:
            first_mismatch = {"index": idx + 1, "expected": expected, "actual": actual_order[idx]}
            break

    missing_from_actual = [item for item in expected_order if item not in actual_order]
    extra_in_actual = [item for item in actual_order if item not in expected_order]

    return {
        "expected_judging_order": expected_order,
        "actual_judging_order": actual_order,
        "first_mismatch": first_mismatch,
        "missing_from_actual": missing_from_actual,
        "extra_in_actual": extra_in_actual,
    }


def count_ready_records(records: list[dict[str, Any]], mode: str) -> int:
    if mode == "started":
        return sum(1 for row in records if row.get("judging_started_at"))
    return sum(1 for row in records if row.get("judging_finished_at"))


def wait_for_analysis_completion(attempts_path: Path, expected_success_count: int, mode: str, poll_interval_sec: float) -> dict[str, Any]:
    while True:
        analyze_result = run_analyze(attempts_path)
        analysis = load_analysis(Path(analyze_result["output"]))
        if analysis.get("matched_count", 0) >= expected_success_count and count_ready_records(analysis.get("records", []), mode) >= expected_success_count:
            return analysis
        time.sleep(poll_interval_sec)


def main() -> int:
    args = build_parser().parse_args()
    scenario = load_scenario(args.scenario)
    user_map = build_user_map(args.users_json)
    output_dir = create_output_dir(args.output_dir)
    project_id = discover_project_id(args, scenario)
    payload_filename, payload_bytes = submit_mod.create_project_payload(args.project_file or scenario.get("project_file"), args.project_filename)

    run_manifest = {
        "generated_at": time.time(),
        "experiment_type": scenario.get("experiment_type", "scenario_benchmark"),
        "scenario_name": scenario.get("name", Path(args.scenario).stem),
        "scenario_file": args.scenario,
        "base_url": args.base_url,
        "api_prefix": args.api_prefix,
        "contest_name": scenario.get("contest_name", args.contest_name),
        "contest_project_name": scenario.get("contest_project_name", args.contest_project_name),
        "project_id": project_id,
        "project_payload_filename": payload_filename,
        "steps": scenario.get("steps", []),
        "expected_judging_order": scenario.get("expected_judging_order", []),
    }
    (output_dir / "run.json").write_text(json.dumps(run_manifest, ensure_ascii=False, indent=2), encoding="utf-8")

    admin_opener, csrf_token = rejudge_mod.login(args.base_url, args.api_prefix, args.admin_username, args.admin_password, args.timeout_sec)
    attempts: list[dict[str, Any]] = []
    sequence_tracker: dict[str, int] = {}
    step_results: list[dict[str, Any]] = []
    rejudge_registry: dict[str, list[dict[str, Any]]] = {}

    for step_index, step in enumerate(scenario.get("steps", []), start=1):
        step_type = step.get("type")
        step_name = step.get("name", f"step_{step_index}")
        if step_type == "submit_batch":
            results = run_submit_batch(
                step,
                step_index,
                user_map,
                sequence_tracker,
                args.base_url,
                args.api_prefix,
                project_id,
                payload_filename,
                payload_bytes,
                args.timeout_sec,
                args.submission_poll_sec,
                args.submission_poll_timeout_sec,
            )
            attempts.extend(results)
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "result_count": len(results),
                "success_count": sum(1 for row in results if row.get("success")),
            })
            submit_mod.write_jsonl(output_dir / "attempts.jsonl", sorted(attempts, key=lambda item: (item.get("request_sent_at", 0.0), item.get("username", ""), item.get("attempt_id", ""))))
        elif step_type == "wait_until":
            result = run_wait_until_submission(step, attempts, admin_opener, args)
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "wait_condition": result["wait_condition"],
            })
        elif step_type == "rejudge_submissions":
            result = run_rejudge_submissions(step, step_index, step_name, attempts, admin_opener, csrf_token, args)
            rejudge_registry[step_name] = list(result["discovered_rejudges"])
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "requested": len(result["request_results"]),
                "successful_requests": sum(1 for row in result["request_results"] if row.get("success")),
                "discovered_rejudges": len(result["discovered_rejudges"]),
                "approved_rejudges": sum(1 for row in result["approve_results"] if row.get("success")),
            })
        elif step_type == "rejudge_contest_project":
            result = run_rejudge_contest_project(step, step_index, step_name, admin_opener, csrf_token, args, project_id)
            rejudge_registry[step_name] = list(result["discovered_rejudges"])
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "successful_requests": sum(1 for row in result["request_results"] if row.get("success")),
                "discovered_rejudges": len(result["discovered_rejudges"]),
                "approved_rejudges": sum(1 for row in result["approve_results"] if row.get("success")),
            })
        elif step_type == "approve_rejudge":
            result = run_approve_rejudge(step, step_index, step_name, admin_opener, csrf_token, args, rejudge_registry)
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "source_step": step.get("source_step"),
                "approved_rejudges": sum(1 for row in result["approve_results"] if row.get("success")),
            })
        elif step_type == "wait_until_rejudge":
            result = run_wait_until_rejudge(step, rejudge_registry, admin_opener, args)
            step_results.append({
                "step_index": step_index,
                "step_name": step_name,
                "type": step_type,
                "wait_condition": result["wait_condition"],
            })
        else:
            raise ScenarioError(f"Unsupported step type: {step_type}")

    attempts.sort(key=lambda item: (item.get("request_sent_at", 0.0), item.get("username", ""), item.get("attempt_id", "")))
    submit_mod.write_jsonl(output_dir / "attempts.jsonl", attempts)
    (output_dir / "step_results.json").write_text(json.dumps(step_results, ensure_ascii=False, indent=2), encoding="utf-8")
    (output_dir / "rejudge_registry.json").write_text(json.dumps(rejudge_registry, ensure_ascii=False, indent=2), encoding="utf-8")

    expected_success_count = sum(1 for row in attempts if row.get("success"))
    analysis = wait_for_analysis_completion(output_dir / "attempts.jsonl", expected_success_count, args.analyze_until, args.poll_interval_sec)
    order_analysis = build_order_analysis(analysis, scenario)
    (output_dir / "order_analysis.json").write_text(json.dumps(order_analysis, ensure_ascii=False, indent=2), encoding="utf-8")

    print(json.dumps({
        "output_dir": str(output_dir),
        "scenario_name": run_manifest["scenario_name"],
        "attempt_count": len(attempts),
        "success_count": expected_success_count,
        "analysis": str(output_dir / "analysis.json"),
        "order_analysis": str(output_dir / "order_analysis.json"),
        "step_results": str(output_dir / "step_results.json"),
        "rejudge_registry": str(output_dir / "rejudge_registry.json"),
        "summary": analysis.get("summary"),
        "order_summary": {
            "first_mismatch": order_analysis.get("first_mismatch"),
            "missing_from_actual": order_analysis.get("missing_from_actual"),
            "extra_in_actual": order_analysis.get("extra_in_actual"),
        },
    }, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
