#!/usr/bin/env python3
"""Run submit + analyze in one command until the benchmark reaches the target state."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

DEFAULT_OUTPUT_DIR = "scripts/output/queue-benchmark"
DEFAULT_POLL_INTERVAL_SEC = 5.0
DEFAULT_ANALYZE_UNTIL = "finished"
DEFAULT_STABLE_POLLS = 1


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Run queue benchmark submit + analyze until completion")
    parser.add_argument("--output-dir", default=DEFAULT_OUTPUT_DIR)
    parser.add_argument("--poll-interval-sec", type=float, default=DEFAULT_POLL_INTERVAL_SEC)
    parser.add_argument("--analyze-until", choices=["started", "finished"], default=DEFAULT_ANALYZE_UNTIL)
    parser.add_argument("--stable-polls", type=int, default=DEFAULT_STABLE_POLLS)
    return parser


def parse_json_from_stdout(stdout: str) -> dict[str, Any]:
    stdout = stdout.strip()
    if not stdout:
        raise RuntimeError("Command returned empty stdout")
    return json.loads(stdout)


def run_submit(output_dir: str, submit_extra_args: list[str]) -> dict[str, Any]:
    cmd = [
        sys.executable,
        "scripts/queue_benchmark_submit.py",
        "--output-dir",
        output_dir,
        *submit_extra_args,
    ]
    completed = subprocess.run(cmd, capture_output=True, text=True)
    if completed.returncode != 0:
        raise RuntimeError(
            "Submit failed\n"
            f"stdout:\n{completed.stdout}\n"
            f"stderr:\n{completed.stderr}"
        )
    result = parse_json_from_stdout(completed.stdout)
    result["_stdout"] = completed.stdout
    result["_stderr"] = completed.stderr
    return result


def run_analyze(attempts_path: Path) -> dict[str, Any]:
    cmd = [
        sys.executable,
        "scripts/queue_benchmark_analyze.py",
        "--run-log",
        str(attempts_path),
    ]
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
    parser = build_parser()
    args, submit_extra_args = parser.parse_known_args()
    if submit_extra_args and submit_extra_args[0] == "--":
        submit_extra_args = submit_extra_args[1:]

    submit_result = run_submit(args.output_dir, submit_extra_args)
    output_dir = Path(submit_result["output_dir"])
    attempts_path = output_dir / "attempts.jsonl"
    expected_success_count = int(submit_result.get("success_count", 0))

    print(json.dumps({
        "phase": "submit_done",
        "output_dir": str(output_dir),
        "attempt_count": submit_result.get("attempt_count"),
        "success_count": expected_success_count,
        "resolved_submission_count": submit_result.get("resolved_submission_count"),
        "analyze_until": args.analyze_until,
    }, ensure_ascii=False, indent=2))

    stable_hits = 0
    poll_count = 0
    while True:
        poll_count += 1
        analyze_result = run_analyze(attempts_path)
        analysis_path = Path(analyze_result["output"])
        analysis = load_analysis(analysis_path)
        done, reason = is_done(analysis, expected_success_count, args.analyze_until)

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
