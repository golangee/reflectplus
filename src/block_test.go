package src

import (
	"strconv"
	"testing"
)

func Test_guessItemName(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"receiver.Fields", "field"},
		{"receiver.blub", "blubElement"},
		{"fields", "field"},
		{"field", "fieldElement"},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := guessItemName(tt.args); got != tt.want {
				t.Errorf("guessItemName() = %v, want %v", got, tt.want)
			}
		})
	}
}
