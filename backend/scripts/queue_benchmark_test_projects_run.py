#!/usr/bin/env python3
"""Submit one per-user test project zip and analyze until completion.

This wrapper mirrors the benchmark flow, but instead of reusing one shared zip,
it scans scripts/test-projects/*.zip and submits each user's matching zip once.
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

import queue_benchmark_submit as submit_mod

DEFAULT_PROJECTS_DIR = "scripts/test-projects"
DEFAULT_OUTPUT_DIR = "scripts/output/queue-benchmark-test-projects"
DEFAULT_POLL_INTERVAL_SEC = 5.0
DEFAULT_ANALYZE_UNTIL = "finished"
DEFAULT_STABLE_POLLS = 1
DEFAULT_START_DELAY_SEC = 3.0
DEFAULT_WINDOW_SEC = 10.0
DEFAULT_SCHEDULE = "spread"
DEFAULT_EXPERIMENT_TYPE = "test_projects_once"


@dataclass
class UserProjectPlan:
    attempt_id: str
    username: str
    password: str
    scheduled_at: float
    project_file: Path


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Submit each available per-user test project zip once, then analyze until completion"
    )
    parser.add_argument("--projects-dir", default=DEFAULT_PROJECTS_DIR)
    parser.add_argument("--users-json", default=submit_mod.DEFAULT_USERS_JSON)
    parser.add_argument("--output-dir", default=DEFAULT_OUTPUT_DIR)
    parser.add_argument("--base-url", default=submit_mod.DEFAULT_BASE_URL)
    parser.add_argument("--api-prefix", default=submit_mod.DEFAULT_API_PREFIX)
    parser.add_argument("--contest-name", default=submit_mod.DEFAULT_CONTEST_NAME)
    parser.add_argument("--contest-project-name", default=submit_mod.DEFAULT_CONTEST_PROJECT_NAME)
    parser.add_argument("--project-id", type=int, default=None)
    parser.add_argument("--usernames", default="")
    parser.add_argument("--start-delay-sec", type=float, default=DEFAULT_START_DELAY_SEC)
    parser.add_argument("--window-sec", type=float, default=DEFAULT_WINDOW_SEC)
    parser.add_argument("--schedule", choices=["spread", "burst"], default=DEFAULT_SCHEDULE)
    parser.add_argument("--timeout-sec", type=float, default=submit_mod.DEFAULT_TIMEOUT_SEC)
    parser.add_argument("--submission-poll-sec", type=float, default=submit_mod.DEFAULT_SUBMISSION_POLL_SEC)
    parser.add_argument("--submission-poll-timeout-sec", type=float, default=submit_mod.DEFAULT_SUBMISSION_POLL_TIMEOUT_SEC)
    parser.add_argument("--analyze-until", choices=["started", "finished"], default=DEFAULT_ANALYZE_UNTIL)
    parser.add_argument("--poll-interval-sec", type=float, default=DEFAULT_POLL_INTERVAL_SEC)
    parser.add_argument("--stable-polls", type=int, default=DEFAULT_STABLE_POLLS)
    parser.add_argument("--experiment-type", default=DEFAULT_EXPERIMENT_TYPE)
    parser.add_argument("--admin-username", default="admin")
    parser.add_argument("--admin-password", default="1234")
    return parser


def parse_json_from_stdout(stdout: str) -> dict[str, Any]:
    stdout = stdout.strip()
    if not stdout:
        raise RuntimeError("Command returned empty stdout")
    return json.loads(stdout)


def discover_available_user_projects(projects_dir: str, usernames: str, users_json: str) -> list[UserProjectPlan]:
    users = submit_mod.read_users(users_json)
    selected_users = submit_mod.select_users(users, usernames, None)
    user_map = {user["username"]: user for user in selected_users}

    wanted = set(user_map.keys()) if usernames else None
    plans: list[UserProjectPlan] = []
    for path in sorted(Path(projects_dir).glob("*.zip")):
        username = path.stem
        if wanted is not None and username not in wanted:
            continue
        user = user_map.get(username)
        if user is None:
            continue
        plans.append(
            UserProjectPlan(
                attempt_id=submit_mod.uuid.uuid4().hex,
                username=username,
                password=user["password"],
                scheduled_at=0.0,
                project_file=path,
            )
        )

    missing_requested: list[str] = []
    if wanted is not None:
        found = {plan.username for plan in plans}
        missing_requested = sorted(wanted - found)
        if missing_requested:
            raise submit_mod.ApiError(
                "Requested usernames are missing zip files in projects dir: " + ", ".join(missing_requested)
            )

    if not plans:
        raise submit_mod.ApiError(f"No user project zip files found in {projects_dir}")
    return plans


def schedule_plans(plans: list[UserProjectPlan], start_delay_sec: float, window_sec: float, schedule: str) -> None:
    start_at = time.time() + start_delay_sec
    if schedule == "burst" or len(plans) == 1:
        for plan in plans:
            plan.scheduled_at = start_at
        return

    interval = window_sec / max(len(plans) - 1, 1)
    for index, plan in enumerate(plans):
        plan.scheduled_at = start_at + index * interval


def submit_user_project(
    plan: UserProjectPlan,
    base_url: str,
    api_prefix: str,
    project_id: int,
    timeout_sec: float,
    submission_poll_sec: float,
    submission_poll_timeout_sec: float,
) -> dict[str, Any]:
    payload_bytes = plan.project_file.read_bytes()
    return submit_mod.submit_once(
        submit_mod.AttemptPlan(
            attempt_id=plan.attempt_id,
            username=plan.username,
            password=plan.password,
            scheduled_at=plan.scheduled_at,
            sequence_for_user=1,
        ),
        base_url,
        api_prefix,
        project_id,
        plan.project_file.name,
        payload_bytes,
        timeout_sec,
        submission_poll_sec,
        submission_poll_timeout_sec,
    )


def write_run_manifest(output_dir: Path, args: argparse.Namespace, project_id: int, plans: list[UserProjectPlan]) -> None:
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
        "selected_usernames": [plan.username for plan in plans],
        "per_user_attempt_count": {plan.username: 1 for plan in plans},
        "heavy_users": [],
        "light_users": [plan.username for plan in plans],
        "attempt_count": len(plans),
        "projects_dir": args.projects_dir,
        "project_files": {plan.username: str(plan.project_file) for plan in plans},
    }
    (output_dir / "run.json").write_text(json.dumps(run_manifest, ensure_ascii=False, indent=2), encoding="utf-8")


def run_analyze(attempts_path: Path, args: argparse.Namespace) -> dict[str, Any]:
    cmd = [
        sys.executable,
        "scripts/queue_benchmark_analyze.py",
        "--run-log",
        str(attempts_path),
        "--base-url",
        args.base_url,
        "--api-prefix",
        args.api_prefix,
        "--admin-username",
        args.admin_username,
        "--admin-password",
        args.admin_password,
        "--contest-name",
        args.contest_name,
        "--contest-project-name",
        args.contest_project_name,
        "--timeout-sec",
        str(args.timeout_sec),
    ]
    if args.project_id is not None:
        cmd.extend(["--project-id", str(args.project_id)])
    completed = subprocess.run(cmd, capture_output=True, text=True)
    if completed.returncode != 0:
        raise RuntimeError(
            "Analyze failed\n"
            f"stdout:\n{completed.stdout}\n"
            f"stderr:\n{completed.stderr}"
        )
    result = parse_json_from_stdout(completed.stdout)
    result["_stdout"] = completed.stdout
    result["_stderr"] = completed.stderr
    return result


def load_analysis(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def count_ready_records(records: list[dict[str, Any]], mode: str) -> int:
    if mode == "started":
        return sum(1 for row in records if row.get("judging_started_at"))
    return sum(1 for row in records if row.get("judging_finished_at"))


def is_done(analysis: dict[str, Any], expected_success_count: int, mode: str) -> tuple[bool, str]:
    matched_count = analysis.get("matched_count", 0)
    if matched_count < expected_success_count:
        return False, f"matched {matched_count}/{expected_success_count}"

    records = analysis.get("records", [])
    ready_count = count_ready_records(records, mode)
    if ready_count < expected_success_count:
        field_name = "judging_started_at" if mode == "started" else "judging_finished_at"
        return False, f"{field_name} ready {ready_count}/{expected_success_count}"

    return True, f"all {expected_success_count} records reached {mode}"


def main() -> int:
    args = build_parser().parse_args()

    plans = discover_available_user_projects(args.projects_dir, args.usernames, args.users_json)
    schedule_plans(plans, args.start_delay_sec, args.window_sec, args.schedule)

    project_id = args.project_id or submit_mod.discover_project_id(
        args.base_url,
        args.api_prefix,
        args.contest_name,
        args.contest_project_name,
        args.timeout_sec,
    )

    timestamp = time.strftime("%Y%m%d-%H%M%S", time.localtime())
    output_dir = Path(args.output_dir) / timestamp
    output_dir.mkdir(parents=True, exist_ok=True)
    write_run_manifest(output_dir, args, project_id, plans)

    results: list[dict[str, Any]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=len(plans)) as executor:
        future_map = {
            executor.submit(
                submit_user_project,
                plan,
                args.base_url,
                args.api_prefix,
                project_id,
                args.timeout_sec,
                args.submission_poll_sec,
                args.submission_poll_timeout_sec,
            ): plan
            for plan in plans
        }
        for future in concurrent.futures.as_completed(future_map):
            results.append(future.result())

    results.sort(key=lambda item: (item.get("request_sent_at", 0.0), item["username"], item["attempt_id"]))
    submit_mod.write_jsonl(output_dir / "attempts.jsonl", results)

    success_count = sum(1 for item in results if item.get("success"))
    resolved_submission_count = sum(1 for item in results if item.get("submission_id"))

    print(json.dumps({
        "phase": "submit_done",
        "output_dir": str(output_dir),
        "attempt_count": len(results),
        "success_count": success_count,
        "resolved_submission_count": resolved_submission_count,
        "available_users": [plan.username for plan in plans],
        "analyze_until": args.analyze_until,
    }, ensure_ascii=False, indent=2))

    attempts_path = output_dir / "attempts.jsonl"
    stable_hits = 0
    poll_count = 0
    while True:
        poll_count += 1
        analyze_result = run_analyze(attempts_path, args)
        analysis_path = Path(analyze_result["output"])
        analysis = load_analysis(analysis_path)
        done, reason = is_done(analysis, success_count, args.analyze_until)

        print(json.dumps({
            "phase": "analyze_poll",
            "poll": poll_count,
            "analysis": str(analysis_path),
            "reason": reason,
            "summary": analysis.get("summary"),
            "execution_summary": analysis.get("execution_summary"),
            "total_summary": analysis.get("total_summary"),
        }, ensure_ascii=False, indent=2))

        if done:
            stable_hits += 1
            if stable_hits >= args.stable_polls:
                print(json.dumps({
                    "phase": "done",
                    "output_dir": str(output_dir),
                    "analysis": str(analysis_path),
                    "polls": poll_count,
                    "summary": analysis.get("summary"),
                    "execution_summary": analysis.get("execution_summary"),
                    "total_summary": analysis.get("total_summary"),
                }, ensure_ascii=False, indent=2))
                return 0
        else:
            stable_hits = 0

        time.sleep(args.poll_interval_sec)


if __name__ == "__main__":
    raise SystemExit(main())
