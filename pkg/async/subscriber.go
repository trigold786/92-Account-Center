package async

import (
	"context"
	"log"
	"sync"
)

type Handler interface {
	Handle(ctx context.Context, msg *Message) error
}

type HandlerFunc func(ctx context.Context, msg *Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg *Message) error {
	return f(ctx, msg)
}

type Subscriber struct {
	rdb      interface{}
	Stream   string
	Group    string
	Consumer string
	handlers map[string][]Handler
	mu       sync.RWMutex
}

func NewSubscriber(rdb interface{}, stream, group, consumer string) *Subscriber {
	return &Subscriber{
		rdb:      rdb,
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		handlers: make(map[string][]Handler),
	}
}

func (s *Subscriber) RegisterHandler(eventType string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[eventType] = append(s.handlers[eventType], handler)
}

func (s *Subscriber) Start(ctx context.Context) error {
	log.Printf("[async] subscriber started: stream=%s group=%s consumer=%s", s.Stream, s.Group, s.Consumer)
	<-ctx.Done()
	return nil
}
