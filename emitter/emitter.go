// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package emitter

import (
	"sync"
)

// Event type
type Event interface{}

// NewEventFunc type
type NewEventFunc func(e Event)

// Subscription type
type Subscription struct {
	onRemove func(s *Subscription)
	ch       chan Event
}

// On invokes given function for each event
func (s *Subscription) On(fn NewEventFunc) {
	for e := range s.ch {
		fn(e)
	}
}

// Unsubscribe stops getting new events
func (s *Subscription) Unsubscribe() {
	s.onRemove(s)
	close(s.ch)
}

func (s *Subscription) emit(event Event) {
	select {
	case s.ch <- event:
	default:
	}
}

// Emitter handles event subscriptions
type Emitter struct {
	mtx           sync.RWMutex
	subscriptions map[*Subscription]struct{}
}

// New creates a new Emitter
func New() *Emitter {
	return &Emitter{
		subscriptions: make(map[*Subscription]struct{}),
	}
}

// Subscribe create a new subscription
func (e *Emitter) Subscribe(buffer int) *Subscription {
	s := &Subscription{
		e.delete,
		make(chan Event, buffer),
	}
	e.add(s)
	return s
}

func (e *Emitter) add(s *Subscription) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.subscriptions[s] = struct{}{}
}

func (e *Emitter) delete(s *Subscription) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	delete(e.subscriptions, s)
}

// Emit sends new event to all subscriptions
func (e *Emitter) Emit(event Event) {
	e.mtx.RLock()
	defer e.mtx.RUnlock()
	for s := range e.subscriptions {
		s.emit(event)
	}
}
