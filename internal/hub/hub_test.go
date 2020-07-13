// package hub_test

// import (
// 	"context"
// 	"io"
// 	"sync"
// 	"testing"
// 	"time"

// 	dispatch "github.com/theverything/communique/internal/hub"
// )

// type mockStore struct {
// 	mu *sync.Mutex

// 	Client           io.Writer
// 	JoinCalled       int
// 	LeaveCalled      int
// 	RegisterCalled   int
// 	UnregisterCalled int
// }

// func (m *mockStore) RegisterClient(client io.Writer) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.RegisterCalled = m.RegisterCalled + 1
// }

// func (m *mockStore) UnregisterClient(client io.Writer) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.UnregisterCalled = m.UnregisterCalled + 1
// }

// func (m *mockStore) JoinTopics(topics []string, client io.Writer) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.JoinCalled = m.JoinCalled + 1
// }

// func (m *mockStore) LeaveTopics(topics []string, client io.Writer) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.LeaveCalled = m.LeaveCalled + 1
// }

// func (m *mockStore) GetTopicClients(topic string) []io.Writer {
// 	return []io.Writer{m.Client}
// }

// func newMockStore(client io.Writer) *mockStore {
// 	mu := new(sync.Mutex)

// 	return &mockStore{
// 		Client: client,
// 		mu:     mu,
// 	}
// }

// type mockClient struct {
// 	id     string
// 	called int
// 	mu     *sync.Mutex
// }

// func (m *mockClient) Write(payload []byte) (int, error) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	m.called = m.called + 1

// 	return len(payload), nil
// }

// func (m *mockClient) calledTimes() int {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	return m.called
// }

// func newMockClient(id string) *mockClient {
// 	mu := new(sync.Mutex)

// 	return &mockClient{
// 		id: id,
// 		mu: mu,
// 	}
// }

// func TestMemoryDispatch(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	client := newMockClient("1")
// 	store := newMockStore(client)
// 	topics := []string{"foo"}

// 	d := dispatch.New(ctx, dispatch.DispatcherConfig{Concurrency: 5}, store).Start()

// 	d.Join(topics, client)
// 	d.Join(topics, client)
// 	d.Join(topics, client)
// 	d.Leave(topics, client)
// 	d.Leave(topics, client)
// 	d.Leave(topics, client)
// 	d.Leave(topics, client)
// 	d.Register(client)
// 	d.Register(client)
// 	d.Register(client)
// 	d.Register(client)
// 	d.Register(client)
// 	d.Unregister(client)
// 	d.Unregister(client)
// 	d.Unregister(client)
// 	d.Unregister(client)
// 	d.Unregister(client)
// 	d.Unregister(client)
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))
// 	d.Dispatch("foo", []byte("hello"))

// 	// wait for calls to be processed
// 	time.Sleep(time.Second * 2)

// 	cancel()

// 	d.Wait()

// 	if store.JoinCalled != 3 {
// 		t.Error("Join not called")
// 	}

// 	if store.LeaveCalled != 4 {
// 		t.Error("Leave not called")
// 	}

// 	if store.RegisterCalled != 5 {
// 		t.Error("Register not called")
// 	}

// 	if store.UnregisterCalled != 6 {
// 		t.Error("Unregister not called")
// 	}

// 	if client.calledTimes() != 10 {
// 		t.Error("Dispatch not called")
// 	}
// }
