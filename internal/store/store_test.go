// package store_test

// import (
// 	"io"
// 	"testing"

// 	"github.com/theverything/communique/internal/store"
// )

// type mockClient struct {
// 	id string
// }

// func (m *mockClient) Write(payload []byte) (int, error) {
// 	return len(payload), nil
// }

// func newMockClient(id string) *mockClient {
// 	return &mockClient{
// 		id: id,
// 	}
// }

// func isIn(s []io.Writer, w io.Writer) bool {
// 	for _, wr := range s {
// 		if wr == w {
// 			return true
// 		}
// 	}

// 	return false
// }

// func TestStore(t *testing.T) {
// 	store := store.New()
// 	client1 := newMockClient("1")
// 	client2 := newMockClient("2")
// 	client3 := newMockClient("3")

// 	store.RegisterClient(client1)
// 	store.RegisterClient(client2)
// 	store.RegisterClient(client3)

// 	store.JoinTopics([]string{"foo"}, client1)
// 	store.JoinTopics([]string{"foo"}, client2)
// 	store.JoinTopics([]string{"foo"}, client3)
// 	store.JoinTopics([]string{"bar"}, client3)

// 	clients := store.GetTopicClients("foo")

// 	if len(clients) != 3 {
// 		t.Fatal("clients were not added to topic foo")
// 	}

// 	store.LeaveTopics([]string{"foo"}, client1)

// 	clients = store.GetTopicClients("foo")

// 	if len(clients) != 2 || !isIn(clients, client2) || !isIn(clients, client3) {
// 		t.Fatal("client1 was not removed from topic foo")
// 	}

// 	store.UnregisterClient(client1)
// 	store.UnregisterClient(client2)

// 	clients = store.GetTopicClients("foo")

// 	if len(clients) != 1 || !isIn(clients, client3) {
// 		t.Fatal("client2 was not removed from topic foo")
// 	}

// 	store.LeaveTopics([]string{"bar"}, client3)

// 	clients = store.GetTopicClients("bar")

// 	if len(clients) != 0 {
// 		t.Fatal("client3 was not removed from topic bar")
// 	}

// 	store.UnregisterClient(client3)

// 	clients = store.GetTopicClients("foo")

// 	if len(clients) != 0 {
// 		t.Fatal("client3 was not removed from topic foo")
// 	}
// }
