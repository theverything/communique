package hub

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/theverything/communique/internal/store"
)

// the following channel buffer sizes are arbitrary
// and can and should be tuned in the future
const (
	joinChannelBufferSize     = 10
	leaveChannelBufferSize    = 10
	dispatchChannelBufferSize = 20
)

// Starter -
type Starter interface {
	Start()
}

// Joiner -
type Joiner interface {
	Join(topic string, client io.Writer)
}

// Leaver -
type Leaver interface {
	Leave(topic string, client io.Writer)
}

// Dispatcher -
type Dispatcher interface {
	Dispatch(topic string, payload []byte)
}

// Hub -
type Hub interface {
	Joiner
	Leaver
	Dispatcher
	Starter
}

type membership struct {
	Topic  string
	Client io.Writer
}

type message struct {
	Topic   string
	Payload []byte
}

type hub struct {
	store       store.Store
	dispatch    chan message
	join        chan membership
	leave       chan membership
	ctx         context.Context
	concurrency uint
	wg          *sync.WaitGroup
}

// Config -
type Config struct {
	Concurrency uint
}

func createMembershipMessage(i int) []byte {
	return []byte(fmt.Sprintf(`{"type":"membership","payload":%d}`, i))
}

func (d *hub) work(id uint) {
	defer d.wg.Done()
	defer func() {
		log.Println("stopping worker", id)
	}()

	for {
		select {
		case membership := <-d.join:
			count := d.store.Set(membership.Topic, membership.Client)

			d.Dispatch(
				membership.Topic,
				createMembershipMessage(count),
			)

		case membership := <-d.leave:
			count := d.store.Remove(membership.Topic, membership.Client)

			d.Dispatch(
				membership.Topic,
				createMembershipMessage(count),
			)

		case dispatch := <-d.dispatch:
			for c := range d.store.Get(dispatch.Topic) {

				go c.Write(dispatch.Payload)
			}

		case <-d.ctx.Done():
			return
		}
	}
}

// Start -
func (d *hub) Start() {
	for i := uint(0); i < d.concurrency; i++ {
		i := i
		d.wg.Add(1)
		go d.work(i)
	}

	d.wg.Wait()
}

// Join -
func (d *hub) Join(topic string, client io.Writer) {
	d.join <- membership{
		Topic:  topic,
		Client: client,
	}
}

// Leave -
func (d *hub) Leave(topic string, client io.Writer) {
	d.leave <- membership{
		Topic:  topic,
		Client: client,
	}
}

// Dispatch -
func (d *hub) Dispatch(topic string, payload []byte) {
	d.dispatch <- message{
		Topic:   topic,
		Payload: payload,
	}
}

// New -
func New(ctx context.Context, config Config, store store.Store) Hub {
	wg := new(sync.WaitGroup)

	return &hub{
		ctx:         ctx,
		store:       store,
		concurrency: config.Concurrency,
		join:        make(chan membership, joinChannelBufferSize),
		leave:       make(chan membership, leaveChannelBufferSize),
		dispatch:    make(chan message, dispatchChannelBufferSize),
		wg:          wg,
	}
}
