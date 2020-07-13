package store_test

import (
	"io"
	"testing"

	"github.com/theverything/communique/internal/store"
)

type mockClient struct {
	id string
}

func (m *mockClient) Write(payload []byte) (int, error) {
	return len(payload), nil
}

func newMockClient(id string) *mockClient {
	return &mockClient{
		id: id,
	}
}

func isIn(m map[io.Writer]struct{}, w io.Writer) bool {
	_, ok := m[w]

	return ok
}

func TestStore(t *testing.T) {
	store := store.New()
	client1 := newMockClient("1")
	client2 := newMockClient("2")
	client3 := newMockClient("3")

	store.Set("foo", client1)
	store.Set("foo", client2)
	store.Set("foo", client3)
	store.Set("bar", client3)

	clients := store.Get("foo")

	if len(clients) != 3 {
		t.Fatal("clients were not added to topic foo")
	}

	store.Remove("foo", client1)

	clients = store.Get("foo")

	if len(clients) != 2 || !isIn(clients, client2) || !isIn(clients, client3) {
		t.Fatal("client1 was not removed from topic foo")
	}

	store.Remove("foo", client2)

	clients = store.Get("foo")

	if len(clients) != 1 || !isIn(clients, client3) {
		t.Fatal("client2 was not removed from topic foo")
	}

	store.Remove("bar", client3)

	clients = store.Get("bar")

	if clients != nil {
		t.Fatal("client3 was not removed from topic bar")
	}

	store.Remove("foo", client3)

	clients = store.Get("foo")

	if clients != nil {
		t.Fatal("client3 was not removed from topic foo")
	}
}
