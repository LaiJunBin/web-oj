# Queue Benchmark Scripts

Internal benchmark scripts for queue and scheduling experiments. They are kept
for reference and are NOT wired to the shipped demo seed — using them takes
manual setup (below).

Main scripts (`queue_benchmark_run.py` / `queue_benchmark_scenario.py`
orchestrate the others; the shell scripts alongside are the key generators and
`detect-docker-gateway.sh`, unrelated to benchmarking):

- `queue_benchmark_submit.py`
- `queue_benchmark_analyze.py`
- `queue_benchmark_rejudge.py`

All Python scripts use the standard library only.

## Setup you must do first

- Have the backend running (default `http://localhost:8080`).
- Create a benchmark contest and project yourself (the defaults look for a
  contest named `Queue Benchmark Contest` with project
  `Queue Benchmark Short Task`; pass `--project-id` to point anywhere else).
- Supply a contestant roster JSON and pass it via `--users-json`
  (list of objects with `username` / `password`; the default path
  `database/seeder/seeders/users.json` is not shipped).
- Have an admin account that can access admin submission and rejudge APIs.

## Script 1: Submit Benchmark Workload

Run a basic 10-user benchmark, one submission per user, spread over 10 seconds:

```bash
python3 scripts/queue_benchmark_submit.py \
  --user-count 10 \
  --submissions-per-user 1 \
  --window-sec 10
```

Run a priority/fairness benchmark where a few users submit more often:

```bash
python3 scripts/queue_benchmark_submit.py \
  --user-count 10 \
  --submissions-per-user 1 \
  --extra-usernames A01,A02 \
  --window-sec 30 \
  --experiment-type priority_fairness
```

### Useful Submit Parameters

- `--user-count`: how many accounts to use from `users.json`
- `--usernames`: explicit comma-separated usernames instead of first N users
- `--submissions-per-user`: base submission count per selected user
- `--extra-usernames`: usernames that should submit one extra time
- `--start-delay-sec`: how long to wait before the workload starts
- `--window-sec`: total submission time window
- `--schedule`: `spread` or `burst`
- `--project-id`: explicit contest project ID
- `--project-file`: use a real zip instead of the generated placeholder zip
- `--experiment-type`: label written into `run.json`
- `--output-dir`: where run logs should be stored

### Submit Output

Each run creates a timestamped directory under `scripts/output/queue-benchmark/` containing:

- `run.json`
- `attempts.jsonl`

The `run.json` file now includes:

- `heavy_users`
- `light_users`
- `per_user_attempt_count`
- `experiment_type`

This makes it easier to evaluate whether frequent submitters are deprioritized as intended.

## Script 2: Analyze Queue Waiting Time

Analyze one run log using admin APIs:

```bash
python3 scripts/queue_benchmark_analyze.py \
  --run-log scripts/output/queue-benchmark/<timestamp>/attempts.jsonl
```

The analyzer treats queue waiting time as the primary metric:

- `waiting_ms = judgingStartedAt - submission.createdAt`

It also provides supporting metrics:

- `execution_ms = judgingFinishedAt - judgingStartedAt`
- `total_ms = judgingFinishedAt - submission.createdAt`

## Main Metric

For queue and scheduling experiments, the primary metric is:

- `queue waiting time = judgingStartedAt - submission.createdAt`

This best matches the current experiment goal, where:

- `submission.createdAt` = accepted submission time
- `judgingStartedAt` = actual judging start time

## Secondary Metrics

The analyzer also outputs two secondary metrics for reference:

- `execution_ms`
  - full judgment lifecycle duration after `judgingStartedAt`
  - not pure testcase execution time
  - may include container setup, file copy, lifecycle hooks, persistence, and cleanup
- `total_ms`
  - end-to-end duration from submission accepted to judgment finished

These secondary metrics should be treated as supporting observations rather than the main queue benchmark conclusion.

### Analyze Output

By default, the analysis result is written next to the run log as `analysis.json`.

It contains:

- `primary_metric`
- `summary`
  - summary of `waiting_ms`
- `execution_summary`
- `total_summary`
- `per_user_waiting`
- `first_submission_per_user_waiting`
- `heavy_user_group`
- `light_user_group`
- `records`

Each `record` now also contains ordering fields:

- `request_rank`
- `submission_rank`
- `judging_start_rank`

These are useful for checking whether the actual judging order matches the intended priority strategy.

## Script 3: Trigger Rejudge Storm

Use this script to create a batch of rejudge requests and optionally auto-approve them.

Example: rejudge the latest 20 submissions in one burst:

```bash
python3 scripts/queue_benchmark_rejudge.py \
  --limit 20 \
  --schedule burst \
  --approve auto \
  --experiment-type rejudge_isolation
```

Example: rejudge only specific users' submissions:

```bash
python3 scripts/queue_benchmark_rejudge.py \
  --usernames A01,A02,B01 \
  --limit 12 \
  --schedule burst \
  --approve auto
```

### Useful Rejudge Parameters

- `--submission-ids`: explicit submission IDs to rejudge
- `--usernames`: restrict rejudge targets to submissions by specific users
- `--limit`: only use the latest N submissions after filtering
- `--start-delay-sec`: how long to wait before starting the storm
- `--window-sec`: burst/spread window
- `--schedule`: `spread` or `burst`
- `--reason-prefix`: prefix used to identify benchmark-generated rejudges
- `--approve`: `auto` or `none`
- `--poll-interval-sec`
- `--poll-timeout-sec`
- `--experiment-type`

### Rejudge Output

Each run creates a timestamped `*-rejudge` directory containing:

- `run.json`
- `rejudge_attempts.jsonl`
- `rejudges.json`
- `rejudge_approve_attempts.jsonl` when `--approve auto`

## Script 4: Run Submit + Analyze Together

If you do not want to manually look up the timestamped output directory, use the wrapper script:

```bash
python3 scripts/queue_benchmark_run.py --   --user-count 10   --submissions-per-user 1   --extra-usernames A01,A02   --window-sec 30   --experiment-type priority_fairness
```

This script will:

1. Run `queue_benchmark_submit.py`
2. Capture the generated output directory automatically
3. Re-run `queue_benchmark_analyze.py` every few seconds
4. Stop only when all records reach the requested state

Useful parameters:

- `--analyze-until started`
  - stop when every matched record has `judging_started_at`
- `--analyze-until finished`
  - stop when every matched record has `judging_finished_at`
- `--poll-interval-sec`
  - how often to rerun analysis
- `--stable-polls`
  - require the completion condition to be true for N consecutive polls before exiting

Example: wait until all records are fully finished:

```bash
python3 scripts/queue_benchmark_run.py   --analyze-until finished   --poll-interval-sec 5   -- --user-count 10 --window-sec 30
```

## Recommended Experiment Flow

### 1. Priority Fairness Experiment

Goal:
- Verify that users who submit repeatedly in a short time are not favored over users who submit only occasionally.

Suggested workload:

```bash
python3 scripts/queue_benchmark_submit.py \
  --user-count 10 \
  --submissions-per-user 1 \
  --extra-usernames A01,A02 \
  --window-sec 30 \
  --experiment-type priority_fairness

python3 scripts/queue_benchmark_analyze.py \
  --run-log scripts/output/queue-benchmark/<timestamp>/attempts.jsonl
```

What to inspect:

- `heavy_user_group.waiting_summary`
- `light_user_group.waiting_summary`
- `heavy_user_group.first_submission_waiting_summary`
- `light_user_group.first_submission_waiting_summary`
- `records[*].judging_start_rank`

### 2. Rejudge Isolation Experiment

Goal:
- Verify that a sudden burst of rejudges does not excessively delay fresh submissions.

Suggested procedure:

1. Start a normal submission benchmark:

```bash
python3 scripts/queue_benchmark_submit.py \
  --user-count 10 \
  --window-sec 30 \
  --experiment-type rejudge_isolation_normal
```

2. In parallel, trigger a rejudge storm against recent submissions:

```bash
python3 scripts/queue_benchmark_rejudge.py \
  --limit 20 \
  --schedule burst \
  --approve auto \
  --experiment-type rejudge_isolation_storm
```

3. After the normal submission run finishes, analyze the normal submission run:

```bash
python3 scripts/queue_benchmark_analyze.py \
  --run-log scripts/output/queue-benchmark/<normal-timestamp>/attempts.jsonl
```

Compare this normal-run result against a baseline submission-only run. The key question is whether `waiting_ms` for fresh submissions remains stable when rejudges surge into the queue.

## Notes

- The analyzer uses existing admin APIs instead of direct MongoDB access.
- If a run contains old submissions from the same users, use a fresh time window and run log so matching stays reliable.
- If you want stricter matching, reduce `--late-tolerance-sec`.
- For thesis writing, prefer `waiting_ms` as the main queue benchmark result and treat `execution_ms` / `total_ms` as supplementary context.


## Script 5: Run State-Driven Scenario

Use this script when the next action should only happen after the system reaches a specific state.

Example: run the built-in priority order validation scenario:

```bash
python3 scripts/queue_benchmark_scenario.py   --scenario scripts/scenarios/priority_order_validation.json   --analyze-until finished   --poll-interval-sec 2
```

This script will:

1. Execute scenario steps in order
2. Submit batches according to each step definition
3. Support both fresh submissions and rejudge flows
4. Wait until target submissions or rejudges reach specific states
5. Run the normal analyzer until all scenario submissions reach the requested state
6. Write `order_analysis.json` so you can compare expected and actual judging order

Supported scenario step types:

- `submit_batch`
- `wait_until`
- `rejudge_submissions`
- `rejudge_contest_project`
- `approve_rejudge`
- `wait_until_rejudge`

Scenario output directory files:

- `run.json`
- `attempts.jsonl`
- `analysis.json`
- `step_results.json`
- `rejudge_registry.json`
- `order_analysis.json`

Built-in scenarios:

- `scripts/scenarios/priority_order_validation.json`
  - `A01~A10` submit once, `A01/A02` repeat, then a later `A10` submission tests ordering behavior
- `scripts/scenarios/rejudge_isolation.json`
  - a contest-project rejudge storm is triggered while a fresh submission arrives later

Example: run the rejudge isolation scenario:

```bash
python3 scripts/queue_benchmark_scenario.py   --scenario scripts/scenarios/rejudge_isolation.json   --analyze-until finished   --poll-interval-sec 2
```
