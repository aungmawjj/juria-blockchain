// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package emitter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) OnNewEvent(e Event) {
	m.Called(e)
}

func TestEmitter_Subscribe(t *testing.T) {
	e := New()
	c1 := new(MockClient)
	c2 := new(MockClient)

	go e.Subscribe(0).Listen(c1.OnNewEvent)
	go e.Subscribe(0).Listen(c2.OnNewEvent)

	events := []string{"Hello", "World"}

	for _, event := range events {
		c1.On("OnNewEvent", event).Once()
		c2.On("OnNewEvent", event).Once()

		e.Emit(event)
		time.Sleep(time.Millisecond)

		c1.AssertExpectations(t)
		c2.AssertExpectations(t)
	}
}

func TestEmitter_Unsubscribe(t *testing.T) {
	e := New()
	c1 := new(MockClient)
	c2 := new(MockClient)

	s1 := e.Subscribe(0)
	s2 := e.Subscribe(0)

	go s1.Listen(c1.OnNewEvent)
	go s2.Listen(c2.OnNewEvent)

	s1.Unsubscribe()

	events := []string{"Hello", "World"}

	for _, event := range events {
		c2.On("OnNewEvent", event).Once()

		e.Emit(event)
		time.Sleep(time.Millisecond)

		c1.AssertNotCalled(t, "OnNewEvent")
		c2.AssertExpectations(t)
	}
}
