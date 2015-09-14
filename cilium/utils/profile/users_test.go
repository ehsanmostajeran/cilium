package profile

import (
	"testing"
)

func TestGetUserID(t *testing.T) {
	users := []User{
		User{ID: 3, Name: "usr1"},
		User{ID: 1000, Name: "usr5"},
		User{ID: 1, Name: "usr2"},
		User{ID: 0, Name: "root"},
	}
	gotID, newID := GetUserID("root", users)
	if gotID != 0 {
		t.Errorf("value of returned ID wasn't the one expected:\ngot  %s\nwant %s", gotID, 0)
	}
	if newID {
		t.Errorf("value of returned new ID wasn't the one expected:\ngot  %d\nwant %d", newID, false)
	}

	gotID, newID = GetUserID("foo", users)
	if gotID != 1001 {
		t.Errorf("value of returned ID wasn't the one expected:\ngot  %s\nwant %s", gotID, 1001)
	}
	if !newID {
		t.Errorf("value of returned new ID wasn't the one expected:\ngot  %t\nwant %t", newID, true)
	}
	if len(users) != 4 {
		t.Errorf("the number of users should be different:\ngot  %d\nwant %d", len(users), 4)
	}
}

func TestOrderUsersByAscendingID(t *testing.T) {
	users := []User{
		User{ID: 1000, Name: "usr5"},
		User{ID: 0, Name: "root"},
		User{ID: 1001, Name: "foo"},
		User{ID: 3, Name: "usr1"},
		User{ID: 1, Name: "usr2"},
	}
	want := []User{
		User{ID: 0, Name: "root"},
		User{ID: 1, Name: "usr2"},
		User{ID: 3, Name: "usr1"},
		User{ID: 1000, Name: "usr5"},
		User{ID: 1001, Name: "foo"},
	}
	OrderUsersByAscendingID(users)
	for i := 0; i < len(users); i++ {
		if users[i] != want[i] {
			t.Errorf("users are blady sorted:\ngot  %s\nwant %s", users[i], want[i])
		}
	}
}

func TestOrderUsersByDescendingID(t *testing.T) {
	users := []User{
		User{ID: 1000, Name: "usr5"},
		User{ID: 0, Name: "root"},
		User{ID: 1001, Name: "foo"},
		User{ID: 3, Name: "usr1"},
		User{ID: 1, Name: "usr2"},
	}
	want := []User{
		User{ID: 1001, Name: "foo"},
		User{ID: 1000, Name: "usr5"},
		User{ID: 3, Name: "usr1"},
		User{ID: 1, Name: "usr2"},
		User{ID: 0, Name: "root"},
	}
	OrderUsersByDescendingID(users)
	for i := 0; i < len(users); i++ {
		if users[i] != want[i] {
			t.Errorf("users are blady sorted:\ngot  %s\nwant %s", users[i], want[i])
		}
	}
}
