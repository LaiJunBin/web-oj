package resultparser_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"testing"
)

func makeResult(exitCode int, stdout string) *execresult.ExecResult {
	return execresult.NewExecResult(&execresult.ExecResultConfig{
		ExitCode: exitCode,
		Stdout:   stdout,
	})
}

// ── ExitCodeParser ──────────────────────────────────────────────────────────

func TestExitCodeParser_Pass(t *testing.T) {
	s, score, stdout := resultparser.ExitCodeParser.Parse(makeResult(0, "build log"), 10)
	if s != status.PASSED || score != 10 {
		t.Errorf("expected PASSED/10, got %v/%v", s, score)
	}
	if stdout != "build log" {
		t.Errorf("exitcode parser must return stdout unchanged, got %q", stdout)
	}
}

func TestExitCodeParser_Fail(t *testing.T) {
	s, score, stdout := resultparser.ExitCodeParser.Parse(makeResult(1, "crash log"), 10)
	if s != status.FAILED || score != 10 {
		t.Errorf("expected FAILED/10, got %v/%v", s, score)
	}
	if stdout != "crash log" {
		t.Errorf("exitcode parser must return stdout unchanged, got %q", stdout)
	}
}

func TestExitCodeParser_String(t *testing.T) {
	if resultparser.ExitCodeParser.String() != "exitcode" {
		t.Errorf("unexpected string: %s", resultparser.ExitCodeParser.String())
	}
}

// ── StdoutParser ─────────────────────────────────────────────────────────────

func TestStdoutParser_Pass(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 8.5"), 10)
	if s != status.PASSED || score != 8.5 {
		t.Errorf("expected PASSED/8.5, got %v/%v", s, score)
	}
}

func TestStdoutParser_Fail(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "fail 3.0"), 10)
	if s != status.FAILED || score != 3.0 {
		t.Errorf("expected FAILED/3.0, got %v/%v", s, score)
	}
}

func TestStdoutParser_CaseInsensitive(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "PASS 10"), 5)
	if s != status.PASSED || score != 10 {
		t.Errorf("expected PASSED/10, got %v/%v", s, score)
	}
}

func TestStdoutParser_IntegerScore(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 0"), 10)
	if s != status.PASSED || score != 0 {
		t.Errorf("expected PASSED/0, got %v/%v", s, score)
	}
}

func TestStdoutParser_IgnoresConfiguredScore(t *testing.T) {
	_, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 7"), 99)
	if score != 7 {
		t.Errorf("expected score 7 from stdout, got %v", score)
	}
}

func TestStdoutParser_MultilineUsesFirstLine(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 5\nsome other output"), 10)
	if s != status.PASSED || score != 5 {
		t.Errorf("expected PASSED/5, got %v/%v", s, score)
	}
}

func TestStdoutParser_InvalidFormat_Error(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "ok 10"), 10)
	if s != status.ERROR || score != 0 {
		t.Errorf("expected ERROR/0, got %v/%v", s, score)
	}
}

func TestStdoutParser_EmptyStdout_Error(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, ""), 10)
	if s != status.ERROR || score != 0 {
		t.Errorf("expected ERROR/0, got %v/%v", s, score)
	}
}

func TestStdoutParser_PassNoNumber_UsesConfiguredScore(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass"), 10)
	if s != status.PASSED || score != 10 {
		t.Errorf("expected PASSED/10, got %v/%v", s, score)
	}
}

func TestStdoutParser_FailNoNumber_UsesConfiguredScore(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "fail"), 10)
	if s != status.FAILED || score != 10 {
		t.Errorf("expected FAILED/10, got %v/%v", s, score)
	}
}

func TestStdoutParser_String(t *testing.T) {
	if resultparser.StdoutParser.String() != "stdout" {
		t.Errorf("unexpected string: %s", resultparser.StdoutParser.String())
	}
}

// ── StdoutParser: the verdict line must not survive into public stdout ──────
// Sample-testcase stdout is contestant-visible; whatever the parser consumed
// must be stripped, and everything the grading script printed AFTER the
// verdict must be preserved byte-for-byte.

func TestStdoutParser_StripsVerdictOnlyOutput(t *testing.T) {
	_, _, stdout := resultparser.StdoutParser.Parse(makeResult(0, "pass 0.75"), 1)
	if stdout != "" {
		t.Errorf("verdict-only stdout must become empty, got %q", stdout)
	}
}

func TestStdoutParser_StripsVerdictKeepsDiagnostics(t *testing.T) {
	_, _, stdout := resultparser.StdoutParser.Parse(makeResult(0, "fail 0\nexpected <div id=\"app\">\n  got <span>"), 1)
	if stdout != "expected <div id=\"app\">\n  got <span>" {
		t.Errorf("diagnostics after the verdict must be preserved verbatim, got %q", stdout)
	}
}

func TestStdoutParser_StripsVerdictAfterLeadingBlankLines(t *testing.T) {
	_, _, stdout := resultparser.StdoutParser.Parse(makeResult(0, "\n  \npass 5\ndetail"), 10)
	if stdout != "detail" {
		t.Errorf("leading blank lines must be consumed with the verdict, got %q", stdout)
	}
}

func TestStdoutParser_StripsVerdictWithCRLF(t *testing.T) {
	s, score, stdout := resultparser.StdoutParser.Parse(makeResult(0, "pass 5\r\ndetail"), 10)
	if s != status.PASSED || score != 5 {
		t.Errorf("CRLF verdict line must parse, got %v/%v", s, score)
	}
	if stdout != "detail" {
		t.Errorf("expected rest after CRLF verdict, got %q", stdout)
	}
}

func TestStdoutParser_ErrorKeepsRawStdout(t *testing.T) {
	_, _, stdout := resultparser.StdoutParser.Parse(makeResult(0, "traceback: boom"), 10)
	if stdout != "traceback: boom" {
		t.Errorf("unparsed stdout must stay raw for diagnosis, got %q", stdout)
	}
}

// ── StdoutParser: verdict must be the entire first line ─────────────────────
// Prefix-only matching once scored "passing tests..." as PASSED with full
// configured score; these pin the anchored behavior.

func TestStdoutParser_PrefixLookalike_Error(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "passing tests..."), 10)
	if s != status.ERROR || score != 0 {
		t.Errorf("'passing tests...' must not parse as a verdict, got %v/%v", s, score)
	}
}

func TestStdoutParser_TrailingJunkOnVerdictLine_Error(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 10 extra"), 10)
	if s != status.ERROR || score != 0 {
		t.Errorf("a malformed verdict line must not parse, got %v/%v", s, score)
	}
}

func TestStdoutParser_TrailingWhitespaceOnVerdictLine_OK(t *testing.T) {
	s, score, _ := resultparser.StdoutParser.Parse(makeResult(0, "pass 5   \nrest"), 10)
	if s != status.PASSED || score != 5 {
		t.Errorf("trailing whitespace after the verdict must be tolerated, got %v/%v", s, score)
	}
}

// ── AutoParser ───────────────────────────────────────────────────────────────

func TestAutoParser_StdoutMatchPriority(t *testing.T) {
	// exit code is non-zero but stdout says pass — stdout wins
	s, score, _ := resultparser.AutoParser.Parse(makeResult(1, "pass 9"), 10)
	if s != status.PASSED || score != 9 {
		t.Errorf("expected PASSED/9, got %v/%v", s, score)
	}
}

func TestAutoParser_FallbackToExitCode_Pass(t *testing.T) {
	s, score, _ := resultparser.AutoParser.Parse(makeResult(0, "no score here"), 10)
	if s != status.PASSED || score != 10 {
		t.Errorf("expected PASSED/10 via exitcode fallback, got %v/%v", s, score)
	}
}

func TestAutoParser_FallbackToExitCode_Fail(t *testing.T) {
	s, score, _ := resultparser.AutoParser.Parse(makeResult(1, "some log output"), 10)
	if s != status.FAILED || score != 10 {
		t.Errorf("expected FAILED/10 via exitcode fallback, got %v/%v", s, score)
	}
}

func TestAutoParser_StdoutBranch_StripsVerdict(t *testing.T) {
	_, _, stdout := resultparser.AutoParser.Parse(makeResult(1, "pass 9\ndebug"), 10)
	if stdout != "debug" {
		t.Errorf("auto's stdout branch must strip the verdict line, got %q", stdout)
	}
}

func TestAutoParser_ExitCodeBranch_KeepsStdout(t *testing.T) {
	_, _, stdout := resultparser.AutoParser.Parse(makeResult(0, "no score here"), 10)
	if stdout != "no score here" {
		t.Errorf("auto's exitcode branch must keep stdout unchanged, got %q", stdout)
	}
}

func TestAutoParser_PrefixLookalike_FallsBackToExitCode(t *testing.T) {
	// pre-hardening this was hijacked into the stdout branch as PASSED/10
	s, score, stdout := resultparser.AutoParser.Parse(makeResult(1, "passing tests..."), 10)
	if s != status.FAILED || score != 10 {
		t.Errorf("expected exitcode fallback FAILED/10, got %v/%v", s, score)
	}
	if stdout != "passing tests..." {
		t.Errorf("fallback branch must keep stdout unchanged, got %q", stdout)
	}
}

func TestAutoParser_String(t *testing.T) {
	if resultparser.AutoParser.String() != "auto" {
		t.Errorf("unexpected string: %s", resultparser.AutoParser.String())
	}
}

// ── Parsers map registration ──────────────────────────────────────────────────

func TestParsers_AllRegistered(t *testing.T) {
	for _, key := range []string{"exitcode", "stdout", "auto"} {
		if _, ok := resultparser.Parsers[key]; !ok {
			t.Errorf("parser %q not registered in Parsers map", key)
		}
	}
}
