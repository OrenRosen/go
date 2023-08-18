package async_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/OrenRosen/go/async"
)

func TestStartPool(t *testing.T) {
	gotChannel := make(chan User)
	s := &service{
		gotChannel: gotChannel,
	}

	pool := async.NewPool(
		async.WithPoolContextPropagator(propagator{}))

	user := User{
		Name: "Some User",
	}

	ctx := context.WithValue(context.Background(), "someKey", "someValue")
	ctx = context.WithValue(ctx, "otherKey", "otherValue")

	pool.RunAsync(ctx, func(ctx context.Context) error {
		val, ok := ctx.Value("someKey").(string)
		require.True(t, ok, "didn't find someKey. Expected to be injected")
		require.Equal(t, "someValue", val)

		otherVal := ctx.Value("otherKey")
		require.Nil(t, otherVal, "expected otherKey not to be injected")

		return s.Do(ctx, user)
	})

	select {
	case <-gotChannel:
	case <-time.After(time.Millisecond * 10):
		t.Fatal("the service wasn't called, timeout")
	}

}

func TestStartPool_SendMany(t *testing.T) {
	gotChannel := make(chan User)
	s := &service{
		gotChannel: gotChannel,
	}

	pool := async.NewPool(
		async.WithPoolErrorHandler(&errorHandler{}),
	)

	users := createRandomUsers(10000)
	for _, user := range users {
		u := user
		pool.RunAsync(context.Background(), func(ctx context.Context) error {
			return s.Do(ctx, u)
		})
	}

	receivedUsers := make([]User, len(users))
	for i := range users {
		select {
		case user := <-gotChannel:
			receivedUsers[i] = user
		case <-time.After(time.Millisecond * 100):
			t.Fatal("the service wasn't called, timeout")
		}
	}

	expectUsers(t, users, receivedUsers)
}

type User struct {
	Name string
}

type service struct {
	gotChannel chan User
}

func (s *service) Do(ctx context.Context, user User) error {
	go func() {
		s.gotChannel <- user
	}()
	return nil
}

func createRandomUsers(num int) []User {
	users := make([]User, num)
	for i := 0; i < num; i++ {
		users[i] = User{
			Name: "Some User " + strconv.Itoa(rand.Int()),
		}
	}

	return users
}

func expectUsers(t *testing.T, expected, actual []User) {
	require.Equal(t, len(expected), len(actual), "length of users is different")
	for _, expectedUser := range expected {
		found := false
		for _, receivedUser := range actual {
			if expectedUser.Name == receivedUser.Name {
				found = true
				break
			}
		}

		require.True(t, found, "expectUsers didn't find user")
	}
}
