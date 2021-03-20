// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package emitter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockListener struct {
	mock.Mock
}

func (m *MockListener) CB(e Event) {
	m.Called(e)
}

func TestEmitter_Subscribe(t *testing.T) {
	e := New()
	ln1 := new(MockListener)
	ln2 := new(MockListener)

	go e.Subscribe(0).Listen(ln1.CB)
	go e.Subscribe(0).Listen(ln2.CB)

	events := []string{"Hello", "World"}

	for _, event := range events {
		ln1.On("CB", event).Once()
		ln2.On("CB", event).Once()

		e.Emit(event)
		time.Sleep(time.Millisecond)

		ln1.AssertExpectations(t)
		ln2.AssertExpectations(t)
	}
}

func TestEmitter_Unsubscribe(t *testing.T) {
	e := New()
	ln1 := new(MockListener)
	ln2 := new(MockListener)

	s1 := e.Subscribe(0)
	s2 := e.Subscribe(0)

	go s1.Listen(ln1.CB)
	go s2.Listen(ln2.CB)

	s1.Unsubscribe()

	events := []string{"Hello", "World"}

	for _, event := range events {
		ln2.On("CB", event).Once()

		e.Emit(event)
		time.Sleep(time.Millisecond)

		ln1.AssertNotCalled(t, "CB")
		ln2.AssertExpectations(t)
	}
}
