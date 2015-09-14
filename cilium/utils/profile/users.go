package profile

import (
	"sort"
)

type User struct {
	ID   int
	Name string
}

// GetUserID returns the user ID of the given username from the slice of users.
// Returns an unused user ID if the user is not found.
func GetUserID(user string, users []User) (id int, newID bool) {
	lastUserID := 0
	for _, v := range users {
		if v.Name == user {
			return v.ID, false
		}
		if v.ID > lastUserID {
			lastUserID = v.ID
		}
	}
	return lastUserID + 1, true
}

type OrderUsersBy func(u1, u2 *User) bool

// OrderUsersByAscendingID orders the slice of users by ascending ID.
func OrderUsersByAscendingID(users []User) {
	ascID := func(u1, u2 *User) bool {
		return u1.ID < u2.ID
	}
	OrderUsersBy(ascID).sort(users)
}

// OrderUsersByDescendingID orders the slice of users by descending ID.
func OrderUsersByDescendingID(users []User) {
	descID := func(u1, u2 *User) bool {
		return u1.ID > u2.ID
	}
	OrderUsersBy(descID).sort(users)
}

func (by OrderUsersBy) sort(users []User) {
	us := &userSorter{
		users: users,
		by:    by,
	}
	sort.Sort(us)
}

type userSorter struct {
	users []User
	by    func(u1, u2 *User) bool
}

func (s *userSorter) Len() int {
	return len(s.users)
}

func (s *userSorter) Swap(i, j int) {
	s.users[i], s.users[j] = s.users[j], s.users[i]
}

func (s *userSorter) Less(i, j int) bool {
	return s.by(&s.users[i], &s.users[j])
}
