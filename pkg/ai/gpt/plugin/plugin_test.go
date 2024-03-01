package plugin

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"testing"
)

func TestClient_ConvertInput(t *testing.T) {
	type args struct {
		input any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "Test with string input",
			args: args{
				input: "test",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test with int input",
			args: args{
				input: 123,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got, err := c.ConvertInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertInput() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ConvertOutput(t *testing.T) {
	type args struct {
		response dto.Message
	}
	tests := []struct {
		name    string
		args    args
		want    *ConvertedResponse
		wantErr bool
	}{
		{
			name: "Test with message",
			args: args{
				response: dto.Message{
					Content: "test",
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got, err := c.ConvertOutput(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertOutput() got = %v, want %v", got, tt.want)
			}
		})
	}
}
