package minion

import "testing"

func Test_isKeyValid(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "invalidKeyLength", args: args{key: "newrelic"}, want: false},
		{name: "EmptyKeyLength", args: args{key: ""}, want: false},
		{name: "validKeyLength", args: args{key: "123456789012345678901234567890123456"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isKeyValid(tt.args.key); got != tt.want {
				t.Errorf("isKeyValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
