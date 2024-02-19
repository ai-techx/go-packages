package errors

import "testing"

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
