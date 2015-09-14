package utils

import (
	"testing"
)

func TestLookupServiceName(t *testing.T) {
	tests := []struct {
		labels map[string]string
		want   string
	}{
		{labels: map[string]string{"com.intent.service": "foo"}, want: "foo"},
		{labels: map[string]string{"com.docker.compose.service": "foo"}, want: "foo"},
		{labels: map[string]string{"com.intent.my.service": "foo"}, want: ""},
	}

	for _, tt := range tests {
		if got := LookupServiceName(tt.labels); got != tt.want {
			t.Errorf("invalid service name:\ngot  %s\nwant %s", got, tt.want)
		}
	}
}
