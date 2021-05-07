package rooms

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func TestAuth_CanJoin(t *testing.T) {
	tests := []struct {
		Kicked     bool
		Invited    bool
		Blocked    bool
		Visibility pb.Visibility
		Expected   bool
	}{
		{Kicked: false, Invited: false, Blocked: false, Visibility: pb.Visibility_VISIBILITY_PUBLIC, Expected: true},
		{Kicked: true, Invited: false, Blocked: false, Visibility: pb.Visibility_VISIBILITY_PUBLIC, Expected: false},
		{Kicked: false, Invited: false, Blocked: true, Visibility: pb.Visibility_VISIBILITY_PUBLIC, Expected: false},
		{Kicked: false, Invited: false, Blocked: false, Visibility: pb.Visibility_VISIBILITY_PRIVATE, Expected: false},
		{Kicked: false, Invited: true, Blocked: false, Visibility: pb.Visibility_VISIBILITY_PRIVATE, Expected: true},
		{Kicked: true, Invited: true, Blocked: false, Visibility: pb.Visibility_VISIBILITY_PRIVATE, Expected: false},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			id := "1234"
			user := 12
			blocker := 2

			room := &Room{
				id:         id,
				visibility: tt.Visibility,
				members:    make(map[int]*Member),
				kicked:     make(map[int]bool),
				invited:    make(map[int]bool),
			}

			if tt.Kicked {
				room.kicked[user] = true
			}

			if tt.Invited {
				room.invited[user] = true
			}

			room.members[blocker] = &Member{id: blocker}

			repository := NewRepository()
			repository.Set(room)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			auth := NewAuth(repository, blocks.NewBackend(db))

			rows := mock.NewRows([]string{"user_id"})
			if tt.Blocked {
				rows.AddRow(blocker)
			}

			mock.ExpectPrepare("^SELECT (.+)").
				ExpectQuery().
				WithArgs(user).
				WillReturnRows(rows)

			res := auth.CanJoin(id, user)
			if res != tt.Expected {
				t.Fatalf("CanJoin actual: %v expected: %v", res, tt.Expected)
			}
		})
	}
}

func TestAuth_FilterWhoCanJoin(t *testing.T) {
	user := int64(12)

	tests := []struct {
		Kicked   bool
		Blocked  bool
		Expected []int64
	}{
		{Kicked: false, Blocked: false, Expected: []int64{user}},
		{Kicked: true, Blocked: false, Expected: []int64{}},
		{Kicked: false, Blocked: true, Expected: []int64{}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			id := "1234"
			blocker := 12

			room := &Room{
				id:         id,
				visibility: pb.Visibility_VISIBILITY_PUBLIC,
				members:    make(map[int]*Member),
				kicked:     make(map[int]bool),
				invited:    make(map[int]bool),
			}

			if tt.Kicked {
				room.kicked[int(user)] = true
			}

			room.members[blocker] = &Member{id: blocker}

			repository := NewRepository()
			repository.Set(room)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			auth := NewAuth(repository, blocks.NewBackend(db))

			rows := mock.NewRows([]string{"user_id"})
			if tt.Blocked {
				rows.AddRow(blocker)
			}

			mock.ExpectPrepare("^SELECT (.+)").
				ExpectQuery().
				WithArgs(user).
				WillReturnRows(rows)

			res := auth.FilterWhoCanJoin(id, []int64{user})
			if !reflect.DeepEqual(res, tt.Expected) {
				t.Fatalf("CanJoin actual: %v expected: %v", res, tt.Expected)
			}
		})
	}
}
