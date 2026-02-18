package brownfield

import (
	"strings"
	"testing"
)

func TestBuildLLMClassificationPrompt_IncludesRubricAndContract(t *testing.T) {
	prompt := buildLLMClassificationPrompt("docs/core/ARCHITECTURE.md", "# Architecture\n")

	for _, want := range []string{
		"Prompt version: " + llmPromptVersion,
		"Category rubric:",
		"Canonical mapping context:",
		"Confidence calibration:",
		`prefer category "unknown"`,
		"Output contract:",
		`{"category":"<allowed-category>","confidence":<0_to_1>,"rationale":"<brief evidence>"}`,
		"Path: docs/core/ARCHITECTURE.md",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q\nprompt:\n%s", want, prompt)
		}
	}
}

func TestBuildLLMClassificationPrompt_TruncatesLongContent(t *testing.T) {
	long := strings.Repeat("A", llmClassificationLimit+500)
	prompt := buildLLMClassificationPrompt("misc/notes.md", long)

	if !strings.Contains(prompt, "\n...\n") {
		t.Fatalf("expected truncated marker in prompt")
	}
	if !strings.Contains(prompt, "Path: misc/notes.md") {
		t.Fatalf("expected path in prompt")
	}
}
