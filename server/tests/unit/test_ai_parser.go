package ai_test

import (
	"testing"

	"flowtask-server/internal/ai"
)

func TestParseAIResponse_DirectJSON(t *testing.T) {
	parser := ai.NewParser()

	input := `[{"title":"Task 1","description":"Desc 1"},{"title":"Task 2","description":"Desc 2"}]`
	var tasks []ai.GeneratedTask

	err := parser.ParseAIResponse(input, &tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "Task 1" {
		t.Errorf("Expected title 'Task 1', got '%s'", tasks[0].Title)
	}
}

func TestParseAIResponse_MarkdownCodeBlock(t *testing.T) {
	parser := ai.NewParser()

	input := "Here are the tasks:\n```json\n[{\"title\":\"Task 1\"},{\"title\":\"Task 2\"}]\n```"
	var tasks []ai.GeneratedTask

	err := parser.ParseAIResponse(input, &tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestParseAIResponse_MarkdownCodeBlockNoLang(t *testing.T) {
	parser := ai.NewParser()

	input := "Here are the tasks:\n```\n[{\"title\":\"Task 1\"}]\n```"
	var tasks []ai.GeneratedTask

	err := parser.ParseAIResponse(input, &tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
}

func TestParseAIResponse_MixedContent(t *testing.T) {
	parser := ai.NewParser()

	input := "I'll help you create a learning plan. Here are the tasks:\n[{\"title\":\"Learn Go\",\"description\":\"Study Go basics\"}]\nLet me know if you need more!"
	var tasks []ai.GeneratedTask

	err := parser.ParseAIResponse(input, &tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Learn Go" {
		t.Errorf("Expected title 'Learn Go', got '%s'", tasks[0].Title)
	}
}

func TestParseAIResponse_InvalidJSON(t *testing.T) {
	parser := ai.NewParser()

	input := "This is not JSON at all"
	var tasks []ai.GeneratedTask

	err := parser.ParseAIResponse(input, &tasks)
	if err == nil {
		t.Fatal("Expected error for invalid input, got nil")
	}
}

func TestExtractMarkdownCodeBlock(t *testing.T) {
	parser := ai.NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "json code block",
			input:    "text\n```json\n{\"key\":\"value\"}\n```",
			expected: "{\"key\":\"value\"}",
		},
		{
			name:     "plain code block",
			input:    "text\n```\n[{\"key\":\"value\"}]\n```",
			expected: "[{\"key\":\"value\"}]",
		},
		{
			name:     "no code block",
			input:    "no code block here",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractMarkdownCodeBlock(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFindJSONArray(t *testing.T) {
	parser := ai.NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "array in text",
			input:    "Here is [1,2,3] array",
			expected: "[1,2,3]",
		},
		{
			name:     "no array",
			input:    "no array here",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.FindJSONArray(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetParsingStrategy(t *testing.T) {
	parser := ai.NewParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "direct JSON",
			input:    `[{"title":"Task 1"}]`,
			expected: "direct_json",
		},
		{
			name:     "markdown codeblock",
			input:    "text\n```json\n[{\"title\":\"Task 1\"}]\n```",
			expected: "markdown_codeblock",
		},
		{
			name:     "mixed content",
			input:    "Here are tasks: [{\"title\":\"Task 1\"}] done",
			expected: "json_array_extraction",
		},
		{
			name:     "invalid",
			input:    "no json here",
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetParsingStrategy(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
