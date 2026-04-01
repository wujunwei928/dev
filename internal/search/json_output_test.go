package search

import (
	"encoding/json"
	"testing"
)

func TestKeyValToResultItems(t *testing.T) {
	input := [][]KeyVal{
		{
			{Key: "标题", Val: "Test Title"},
			{Key: "链接", Val: "https://example.com"},
			{Key: "描述", Val: "Test Description"},
		},
	}

	result := KeyValToResultItems("bing", "test query", input)

	if result.Engine != "bing" {
		t.Errorf("expected engine 'bing', got %q", result.Engine)
	}
	if result.Query != "test query" {
		t.Errorf("expected query 'test query', got %q", result.Query)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", result.Results[0].Title)
	}
	if result.Results[0].URL != "https://example.com" {
		t.Errorf("expected url 'https://example.com', got %q", result.Results[0].URL)
	}
}

func TestKeyValToResultItems_DoubanKeyName(t *testing.T) {
	input := [][]KeyVal{
		{
			{Key: "名称", Val: "Douban Title"},
			{Key: "链接", Val: "https://douban.com/xxx"},
			{Key: "描述", Val: "Douban Desc"},
		},
	}

	result := KeyValToResultItems("douban", "test", input)

	if result.Results[0].Title != "Douban Title" {
		t.Errorf("expected '名称' mapped to title, got %q", result.Results[0].Title)
	}
}

func TestSearchResult_ToJSON(t *testing.T) {
	sr := SearchResult{
		Engine: "bing",
		Query:  "test",
		Results: []ResultItem{
			{Title: "Title", URL: "https://example.com", Description: "Desc"},
		},
	}

	jsonStr, err := sr.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed SearchResult
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed.Engine != "bing" {
		t.Error("JSON roundtrip failed")
	}
}

func TestKeyValToResultItems_EmptyInput(t *testing.T) {
	result := KeyValToResultItems("bing", "test", nil)
	if len(result.Results) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(result.Results))
	}
}

func TestKeyValToResultItems_IgnoresUnknownKeys(t *testing.T) {
	input := [][]KeyVal{
		{
			{Key: "评分", Val: "9.0"},
			{Key: "标题", Val: "Title"},
		},
	}

	result := KeyValToResultItems("douban", "test", input)
	if result.Results[0].Title != "Title" {
		t.Errorf("expected title 'Title', got %q", result.Results[0].Title)
	}
}
