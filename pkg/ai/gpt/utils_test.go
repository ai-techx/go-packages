package gpt

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"reflect"
	"testing"
)

func TestIsOpenAIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Test with OpenAI endpoint",
			url:  "https://api.openai.com/v1/engines",
			want: true,
		},
		{
			name: "Test with non-OpenAI endpoint",
			url:  "https://example.com",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOpenAIEndpoint(tt.url); got != tt.want {
				t.Errorf("isOpenAIEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterOutUserMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []dto.Message
		want     []dto.Message
	}{
		{
			name: "Test with user messages",
			messages: []dto.Message{
				{Role: dto.RoleUser},
				{Role: dto.RoleAssistant},
				{Role: dto.RoleUser},
			},
			want: []dto.Message{
				{Role: dto.RoleAssistant},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterOutUserMessages(tt.messages); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterOutUserMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertFunctionContentToString(t *testing.T) {
	tests := []struct {
		name    string
		content interface{}
		want    string
	}{
		{
			name:    "Test with string content",
			content: "test",
			want:    "test",
		},
		{
			name:    "Test with *string content",
			content: stringPtr("test"),
			want:    "test",
		},
		{
			name:    "Test with int content",
			content: 123,
			want:    "123",
		},
		{
			name:    "Test with int content",
			content: nil,
			want:    "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertFunctionContentToString(tt.content); got != tt.want {
				t.Errorf("convertFunctionContentToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []dto.Message
		want     []dto.Message
	}{
		{
			name: "Test with usage not nil",
			messages: []dto.Message{
				{Usage: &dto.Usage{}},
				{Usage: &dto.Usage{}},
			},
			want: []dto.Message{
				{Usage: nil},
				{Usage: nil},
			},
		},
		{
			name: "Test with usage nil",
			messages: []dto.Message{
				{Usage: nil},
				{Usage: nil},
			},
			want: []dto.Message{
				{Usage: nil},
				{Usage: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanMessages(tt.messages); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
