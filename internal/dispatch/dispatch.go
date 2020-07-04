package dispatch

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/theverything/communique/internal/store"
)

// the following channel buffer sizes are arbitrary
// and can and should be tuned in the future
const (
	joinChannelBufferSize       = 10
	leaveChannelBufferSize      = 10
	dispatchChannelBufferSize   = 20
	registerChannelBufferSize   = 10
	unregisterChannelBufferSize = 10
)

type Dispatcher interface {
	Join(topics []string, client io.Writer)
	Leave(topics []string, client io.Writer)
	Dispatch(topic string, payload []byte)
	Register(client io.Writer)
	Unregister(client io.Writer)
}

type membership struct {
	Topics []string
	Client io.Writer
}

type message struct {
	Topic   string
	Payload []byte
}

type MemoryDispatch struct {
	store       store.Store
	join        chan membership
	leave       chan membership
	dispatch    chan message
	register    chan io.Writer
	unregister  chan io.Writer
	ctx         context.Context
	concurrency uint
	wg          *sync.WaitGroup
}

// DispatcherConfig -
type DispatcherConfig struct {
	Concurrency uint
}

func (d *MemoryDispatch) work(id uint) {
	defer d.wg.Done()
	defer func() {
		log.Println("stopping worker", id)
	}()

	for {
		select {
		case client := <-d.register:
			d.store.RegisterClient(client)

		case client := <-d.unregister:
			d.store.UnregisterClient(client)

		case membership := <-d.join:
			d.store.JoinTopics(membership.Topics, membership.Client)

		case membership := <-d.leave:
			d.store.LeaveTopics(membership.Topics, membership.Client)

		case dispatch := <-d.dispatch:
			for _, c := range d.store.GetTopicClients(dispatch.Topic) {
				go c.Write(dispatch.Payload)
			}

		case <-d.ctx.Done():
			return
		}
	}
}

// Start -
func (d *MemoryDispatch) Start() *MemoryDispatch {
	for i := uint(0); i < d.concurrency; i++ {
		i := i
		d.wg.Add(1)
		go d.work(i)
	}

	return d
}

// Join -
func (d *MemoryDispatch) Join(topics []string, client io.Writer) {
	d.join <- membership{
		Topics: topics,
		Client: client,
	}
}

// Leave -
func (d *MemoryDispatch) Leave(topics []string, client io.Writer) {
	d.leave <- membership{
		Topics: topics,
		Client: client,
	}
}

// Dispatch -
func (d *MemoryDispatch) Dispatch(topic string, payload []byte) {
	d.dispatch <- message{
		Topic:   topic,
		Payload: payload,
	}
}

// Register -
func (d *MemoryDispatch) Register(client io.Writer) {
	d.register <- client
}

// Unregister -
func (d *MemoryDispatch) Unregister(client io.Writer) {
	d.unregister <- client
}

// Wait -
func (d *MemoryDispatch) Wait() {
	d.wg.Wait()
}

// New -
func New(ctx context.Context, config DispatcherConfig, store store.Store) *MemoryDispatch {
	wg := new(sync.WaitGroup)

	return &MemoryDispatch{
		ctx:         ctx,
		store:       store,
		concurrency: config.Concurrency,
		join:        make(chan membership, joinChannelBufferSize),
		leave:       make(chan membership, leaveChannelBufferSize),
		dispatch:    make(chan message, dispatchChannelBufferSize),
		register:    make(chan io.Writer, registerChannelBufferSize),
		unregister:  make(chan io.Writer, unregisterChannelBufferSize),
		wg:          wg,
	}
}
