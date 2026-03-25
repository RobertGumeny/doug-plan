package prompt

import (
	"bytes"
	"strings"
	"testing"
)

// ---- SelectOne ----

func TestSelectOne_NonTTY_ReturnsDefault(t *testing.T) {
	idx, val, err := SelectOne(new(bytes.Buffer), strings.NewReader(""), false, "Pick one", []string{"a", "b", "c"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 || val != "b" {
		t.Errorf("want idx=1 val=b; got idx=%d val=%s", idx, val)
	}
}

func TestSelectOne_TTY_ValidChoice(t *testing.T) {
	var out bytes.Buffer
	idx, val, err := SelectOne(&out, strings.NewReader("2\n"), true, "Pick one", []string{"claude", "codex", "gemini"}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 || val != "codex" {
		t.Errorf("want idx=1 val=codex; got idx=%d val=%s", idx, val)
	}
	if !strings.Contains(out.String(), "Pick one") {
		t.Error("expected question text in output")
	}
}

func TestSelectOne_TTY_EmptyInputReturnsDefault(t *testing.T) {
	idx, val, err := SelectOne(new(bytes.Buffer), strings.NewReader("\n"), true, "Pick", []string{"x", "y"}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 0 || val != "x" {
		t.Errorf("want default; got idx=%d val=%s", idx, val)
	}
}

func TestSelectOne_TTY_OutOfRangeReturnsDefault(t *testing.T) {
	idx, val, err := SelectOne(new(bytes.Buffer), strings.NewReader("99\n"), true, "Pick", []string{"x", "y"}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 0 || val != "x" {
		t.Errorf("want default; got idx=%d val=%s", idx, val)
	}
}

func TestSelectOne_TTY_NonNumericReturnsDefault(t *testing.T) {
	idx, val, err := SelectOne(new(bytes.Buffer), strings.NewReader("abc\n"), true, "Pick", []string{"x", "y"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 || val != "y" {
		t.Errorf("want default idx=1 val=y; got idx=%d val=%s", idx, val)
	}
}

func TestSelectOne_EmptyOptions_ReturnsError(t *testing.T) {
	_, _, err := SelectOne(new(bytes.Buffer), strings.NewReader(""), true, "Pick", []string{}, 0)
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestSelectOne_TTY_MarkerShownForDefault(t *testing.T) {
	var out bytes.Buffer
	SelectOne(&out, strings.NewReader("\n"), true, "Pick", []string{"a", "b", "c"}, 1)
	output := out.String()
	if !strings.Contains(output, "[x]") {
		t.Errorf("expected [x] marker for default item, got: %s", output)
	}
}

func TestSelectOne_NonTTY_DefaultIdxOutOfRange_ClampsToZero(t *testing.T) {
	idx, val, err := SelectOne(new(bytes.Buffer), strings.NewReader(""), false, "Pick", []string{"a", "b"}, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 0 || val != "a" {
		t.Errorf("want idx=0 val=a; got idx=%d val=%s", idx, val)
	}
}

// ---- SelectMulti ----

func TestSelectMulti_NonTTY_ReturnsDefault(t *testing.T) {
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader(""), false, "Pick", []string{"a", "b", "c"}, []int{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "b" {
		t.Errorf("want [b]; got %v", result)
	}
}

func TestSelectMulti_NonTTY_EmptyDefaultIdxs_ReturnsFirst(t *testing.T) {
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader(""), false, "Pick", []string{"a", "b"}, []int{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("want [a]; got %v", result)
	}
}

func TestSelectMulti_TTY_ToggleAndConfirm(t *testing.T) {
	var out bytes.Buffer
	// Toggle item 2 (codex), then confirm with empty line.
	result, err := SelectMulti(&out, strings.NewReader("2\n\n"), true, "Pick", []string{"claude", "codex", "gemini"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "codex" {
		t.Errorf("want [codex]; got %v", result)
	}
}

func TestSelectMulti_TTY_SelectMultiple(t *testing.T) {
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader("1\n2\n\n"), true, "Pick", []string{"claude", "codex", "gemini"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != "claude" || result[1] != "codex" {
		t.Errorf("want [claude codex]; got %v", result)
	}
}

func TestSelectMulti_TTY_EmptyInputReturnsDefault(t *testing.T) {
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader("\n"), true, "Pick", []string{"a", "b"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("want [a]; got %v", result)
	}
}

func TestSelectMulti_TTY_ToggleOnThenOff(t *testing.T) {
	// Toggle item 2 on, then off again, then confirm -> falls back to default.
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader("2\n2\n\n"), true, "Pick", []string{"a", "b"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("want [a] (default); got %v", result)
	}
}

func TestSelectMulti_TTY_EOFConfirms(t *testing.T) {
	// Reader EOF after a toggle should confirm with that selection.
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader("2\n"), true, "Pick", []string{"a", "b", "c"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "b" {
		t.Errorf("want [b]; got %v", result)
	}
}

func TestSelectMulti_TTY_OutOfRangeIgnored(t *testing.T) {
	result, err := SelectMulti(new(bytes.Buffer), strings.NewReader("99\n\n"), true, "Pick", []string{"a", "b"}, []int{0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No valid toggle -> falls back to default.
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("want [a]; got %v", result)
	}
}

func TestSelectMulti_EmptyOptions_ReturnsError(t *testing.T) {
	_, err := SelectMulti(new(bytes.Buffer), strings.NewReader(""), true, "Pick", []string{}, []int{})
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestSelectMulti_TTY_QuestionShownInOutput(t *testing.T) {
	var out bytes.Buffer
	SelectMulti(&out, strings.NewReader("\n"), true, "Choose providers", []string{"a", "b"}, []int{0})
	if !strings.Contains(out.String(), "Choose providers") {
		t.Errorf("expected question in output; got: %s", out.String())
	}
}

// ---- Text ----

func TestText_NonTTY_ReturnsDefault(t *testing.T) {
	got, err := Text(new(bytes.Buffer), strings.NewReader(""), false, "Name?", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "alice" {
		t.Errorf("want alice; got %s", got)
	}
}

func TestText_TTY_UserInput(t *testing.T) {
	got, err := Text(new(bytes.Buffer), strings.NewReader("bob\n"), true, "Name?", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "bob" {
		t.Errorf("want bob; got %s", got)
	}
}

func TestText_TTY_EmptyInputReturnsDefault(t *testing.T) {
	got, err := Text(new(bytes.Buffer), strings.NewReader("\n"), true, "Name?", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "alice" {
		t.Errorf("want alice; got %s", got)
	}
}

func TestText_TTY_DefaultShownInPrompt(t *testing.T) {
	var out bytes.Buffer
	_, _ = Text(&out, strings.NewReader("\n"), true, "Name?", "alice")
	if !strings.Contains(out.String(), "[alice]") {
		t.Errorf("expected default value in prompt, got: %s", out.String())
	}
}

func TestText_TTY_NoDefaultNoBrackets(t *testing.T) {
	var out bytes.Buffer
	_, _ = Text(&out, strings.NewReader("\n"), true, "Name?", "")
	if strings.Contains(out.String(), "[") {
		t.Errorf("no brackets expected when no default, got: %s", out.String())
	}
}

func TestText_NonTTY_EmptyDefault_ReturnsEmpty(t *testing.T) {
	got, err := Text(new(bytes.Buffer), strings.NewReader(""), false, "Name?", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("want empty string; got %q", got)
	}
}
