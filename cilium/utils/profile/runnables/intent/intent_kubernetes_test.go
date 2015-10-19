package intent

import (
	"testing"
)

func TestConvertMapTo(t *testing.T) {
	type foo struct {
		Bar    string
		Foobar int
		Barbar *int
	}
	myMap := map[string]interface{}{
		"bar":    "bar",
		"foobar": 4,
		"barbar": 5,
	}
	var obj foo
	err := convertMapTo(myMap, &obj)
	if err != nil {
		t.Fatal("Error while converting map to object: %s")
	}
	if obj.Bar != "bar" {
		t.Errorf("invalid obj.Bar:\ngot  %s\nwant %s",
			obj.Bar,
			"bar")
	}
	if obj.Foobar != 4 {
		t.Errorf("invalid obj.Foobar:\ngot  %d\nwant %d",
			obj.Foobar,
			4)
	}
	if obj.Barbar == nil || *obj.Barbar != 5 {
		t.Errorf("invalid *obj.Barbar:\ngot  %+v\nwant %d",
			obj.Barbar,
			5)
	}
}
