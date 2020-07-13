package hub_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/theverything/communique/internal/hub"
)

type mockStore struct {
	mu *sync.Mutex

	Client       io.Writer
	SetCalled    int
	GetCalled    int
	RemoveCalled int
}

func (m *mockStore) Get(topic string) map[io.Writer]struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetCalled = m.GetCalled + 1

	return map[io.Writer]struct{}{
		m.Client: {},
	}
}

func (m *mockStore) Remove(topic string, client io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RemoveCalled = m.RemoveCalled + 1
}

func (m *mockStore) Set(topic string, client io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SetCalled = m.SetCalled + 1
}

func newMockStore(client io.Writer) *mockStore {
	mu := new(sync.Mutex)

	return &mockStore{
		Client: client,
		mu:     mu,
	}
}

type mockClient struct {
	id     string
	called int
	mu     *sync.Mutex
}

func (m *mockClient) Write(payload []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.called = m.called + 1

	return len(payload), nil
}

func (m *mockClient) calledTimes() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.called
}

func newMockClient(id string) *mockClient {
	mu := new(sync.Mutex)

	return &mockClient{
		id: id,
		mu: mu,
	}
}

func TestMemoryDispatch(t *testing.T) {
	doneChan := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	client := newMockClient("1")
	store := newMockStore(client)
	topic := "foo"

	d := hub.New(ctx, hub.Config{Concurrency: 5}, store)

	go func() {
		d.Start()
		doneChan <- struct{}{}
	}()

	d.Join(topic, client)
	d.Join(topic, client)
	d.Join(topic, client)
	d.Leave(topic, client)
	d.Leave(topic, client)
	d.Leave(topic, client)
	d.Leave(topic, client)
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))
	d.Dispatch("foo", []byte("hello"))

	// wait for calls to be processed
	time.Sleep(time.Second * 2)

	cancel()

	<-doneChan

	if store.SetCalled != 3 {
		t.Errorf("Join not called received %d", store.SetCalled)
	}

	if store.RemoveCalled != 4 {
		t.Errorf("Leave not called received %d", store.RemoveCalled)
	}

	if store.GetCalled != 10 {
		t.Errorf("Dispatch not called received %d", store.GetCalled)
	}

	if client.calledTimes() != 10 {
		t.Errorf("client not called received %d", client.calledTimes())
	}
}
