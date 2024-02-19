package errors

import (
	"fmt"
	"testing"
)

func TestMapErrorCodeToHTTPStatus(t *testing.T) {
	type args struct {
		code ErrorCode
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "MissingAPIKey",
			args: args{code: ErrorMissingAPIKey},
			want: 401,
		},
		{
			name: "InvalidAPIKey",
			args: args{code: ErrorInvalidAPIKey},
			want: 401,
		},
		{
			name: "Unknown",
			args: args{code: 1},
			want: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapErrorCodeToHTTPStatus(tt.args.code); got != tt.want {
				t.Errorf("MapErrorCodeToHTTPStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapErrorToHTTPStatus(t *testing.T) {
	type args struct {
		error error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "MissingAPIKey",
			args: args{error: NewDocumentNotFound()},
			want: 404,
		},
		{
			name: "OtherError",
			args: args{error: fmt.Errorf("other error")},
			want: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapErrorToHTTPStatus(tt.args.error); got != tt.want {
				t.Errorf("MapErrorToHTTPStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
