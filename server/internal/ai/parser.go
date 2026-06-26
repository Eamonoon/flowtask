package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// Parser handles parsing AI responses in various formats
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
}

// ParseAIResponse parses AI response trying multiple strategies
// Supports: pure JSON, markdown code blocks, mixed text with JSON
func (p *Parser) ParseAIResponse(raw string, target interface{}) error {
	log.Printf("[PARSER] Attempting to parse AI response (length: %d)", len(raw))

	// Pre-process: clean up common issues
	cleaned := p.cleanJSON(raw)

	// Strategy 1: Direct JSON parsing
	if err := json.Unmarshal([]byte(cleaned), target); err == nil {
		log.Printf("[PARSER] Successfully parsed with strategy: direct_json")
		return nil
	} else {
		log.Printf("[PARSER] Strategy 1 (direct_json) failed: %v", err)
	}

	// Strategy 2: Extract from markdown code block
	jsonBlock := p.ExtractMarkdownCodeBlock(raw)
	if jsonBlock != "" {
		if err := json.Unmarshal([]byte(jsonBlock), target); err == nil {
			log.Printf("[PARSER] Successfully parsed with strategy: markdown_codeblock")
			return nil
		}
		log.Printf("[PARSER] Strategy 2 (markdown_codeblock) failed: extracted block is not valid JSON")
	} else {
		log.Printf("[PARSER] Strategy 2 (markdown_codeblock) skipped: no code block found")
	}

	// Strategy 3: Find JSON array in mixed content
	jsonContent := p.FindJSONArray(raw)
	if jsonContent != "" {
		log.Printf("[PARSER] Found JSON array (length: %d)", len(jsonContent))
		if err := json.Unmarshal([]byte(jsonContent), target); err == nil {
			log.Printf("[PARSER] Successfully parsed with strategy: json_array_extraction")
			return nil
		} else {
			log.Printf("[PARSER] Strategy 3 (json_array_extraction) failed: %v", err)
			log.Printf("[PARSER] JSON array content: %s", truncate(jsonContent, 500))
		}
	} else {
		log.Printf("[PARSER] Strategy 3 (json_array_extraction) skipped: no JSON array found")
	}

	// Strategy 4: Find JSON object in mixed content
	jsonObj := p.FindJSONObject(raw)
	if jsonObj != "" {
		if err := json.Unmarshal([]byte(jsonObj), target); err == nil {
			log.Printf("[PARSER] Successfully parsed with strategy: json_object_extraction")
			return nil
		}
		log.Printf("[PARSER] Strategy 4 (json_object_extraction) failed: extracted content is not valid JSON")
	} else {
		log.Printf("[PARSER] Strategy 4 (json_object_extraction) skipped: no JSON object found")
	}

	// All strategies failed
	log.Printf("[PARSER] ERROR: All parsing strategies failed for response (length: %d)", len(raw))
	return &ParseError{
		Code:    "PARSE_FAILED",
		Message: "unable to parse AI response: no valid JSON found",
		Details: fmt.Sprintf("response length: %d, first 100 chars: %s", len(raw), truncate(raw, 100)),
	}
}

// ParseError represents a parsing error with details
type ParseError struct {
	Code    string
	Message string
	Details string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// ExtractMarkdownCodeBlock extracts JSON from markdown code blocks
// Supports: ```json ... ```, ``` ... ```
func (p *Parser) ExtractMarkdownCodeBlock(content string) string {
	// Match ```json ... ``` or ``` ... ```
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// FindJSONArray finds the first JSON array in the content
func (p *Parser) FindJSONArray(content string) string {
	// Find first [ and last ]
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	if start != -1 && end > start {
		return content[start : end+1]
	}
	return ""
}

// FindJSONObject finds the first JSON object in the content
func (p *Parser) FindJSONObject(content string) string {
	// Find first { and last }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end > start {
		return content[start : end+1]
	}
	return ""
}

// ValidateJSON checks if a string is valid JSON
func (p *Parser) ValidateJSON(content string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(content), &js) == nil
}

// GetParsingStrategy returns which strategy successfully parsed the content
func (p *Parser) GetParsingStrategy(raw string) string {
	var temp interface{}

	// Try direct JSON
	if err := json.Unmarshal([]byte(raw), &temp); err == nil {
		return "direct_json"
	}

	// Try markdown code block
	jsonBlock := p.ExtractMarkdownCodeBlock(raw)
	if jsonBlock != "" {
		if err := json.Unmarshal([]byte(jsonBlock), &temp); err == nil {
			return "markdown_codeblock"
		}
	}

	// Try JSON array
	jsonContent := p.FindJSONArray(raw)
	if jsonContent != "" {
		if err := json.Unmarshal([]byte(jsonContent), &temp); err == nil {
			return "json_array_extraction"
		}
	}

	// Try JSON object
	jsonObj := p.FindJSONObject(raw)
	if jsonObj != "" {
		if err := json.Unmarshal([]byte(jsonObj), &temp); err == nil {
			return "json_object_extraction"
		}
	}

	return "failed"
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// cleanJSON cleans up common JSON issues from AI responses
func (p *Parser) cleanJSON(s string) string {
	// Remove BOM if present
	if len(s) >= 3 && s[0] == 0xef && s[1] == 0xbb && s[2] == 0xbf {
		s = s[3:]
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}
