package plugin

import (
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"reflect"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func TestStandardOutputPlugin_ConvertInput(t *testing.T) {
	type args struct {
		input any
	}
	testPtr := strPtr("test")
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "Test 1",
			args: args{
				input: "test",
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "Test 2",
			args: args{
				input: 1,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test 3",
			args: args{
				input: false,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test 4",
			args: args{
				input: testPtr,
			},
			want:    testPtr,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStandardOutputPlugin()
			got, err := s.ConvertInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertInput() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStandardOutputPlugin_Name(t *testing.T) {
	s := NewStandardOutputPlugin()
	want := "standard"
	if got := s.Name(); got != want {
		t.Errorf("Name() = %v, want %v", got, want)
	}
}

func TestStandardOutputPlugin_Description(t *testing.T) {
	s := NewStandardOutputPlugin()
	want := "Default plugin. Will return the output and as is."
	if got := s.Description(); got != want {
		t.Errorf("Description() = %v, want %v", got, want)
	}
}

func TestStandardOutputPlugin_ConvertOutput(t *testing.T) {
	type args struct {
		response dto.Message
	}
	mockMessage := dto.Message{
		Content: "test",
	}

	tests := []struct {
		name    string
		args    args
		want    *ConvertedResponse
		wantErr bool
	}{{
		name: "Test 1",
		args: args{
			response: mockMessage,
		},
		want: &ConvertedResponse{
			Action:  ContinueOutputAction,
			Message: &mockMessage,
		},
		wantErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStandardOutputPlugin()
			got, err := s.ConvertOutput(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want.Message.Content != got.Message.Content {
				t.Errorf("ConvertOutput() got = %v, want %v", got, tt.want)
			}

			if tt.want.Action != got.Action {
				t.Errorf("ConvertOutput() got = %v, want %v", got, tt.want)
			}
		})
	}
}
