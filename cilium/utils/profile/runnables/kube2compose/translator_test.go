package kube2compose

import (
	"reflect"
	"testing"
)

func TestGetVars(t *testing.T) {
	m, err := getVars("/{version}/images/{name}/json", "/v1.19/images/nginx/json")
	if err != nil {
		t.Errorf("%+v", err)
	}
	want := map[string]string{
		"name":    "nginx",
		"version": "v1.19",
	}
	if !reflect.DeepEqual(m, want) {
		t.Fatalf("got %v, want %+v", m, want)
	}
	m, err = getVars("/{ip}:{port}/", "/192.168.50.1:8080/")
	if err != nil {
		t.Errorf("%+v", err)
	}
	want = map[string]string{
		"ip":   "192.168.50.1",
		"port": "8080",
	}
	if !reflect.DeepEqual(m, want) {
		t.Fatalf("got %v, want %+v", m, want)
	}

}
