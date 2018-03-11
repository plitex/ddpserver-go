package ddpserver

import (
	"sync"
)

type Emiter struct {
	// Mutex
	mu sync.Mutex

	// As we don't have constructor, we must init on first method call
	initialized bool

	// Map storing listener functions by event name
	listeners map[string][]func(string, ...interface{})
}

func (e *Emiter) init() {
	if !e.initialized {
		e.listeners = make(map[string][]func(string, ...interface{}))
		e.initialized = true
	}
}

func (e *Emiter) On(event string, fn func(string, ...interface{})) {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		e.listeners[event] = append(listeners, fn)
	} else {
		e.listeners[event] = []func(string, ...interface{}){fn}
	}
}

func (e *Emiter) Emit(event string, params ...interface{}) {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		for _, l := range listeners {
			l(event, params)
		}
	}
}
