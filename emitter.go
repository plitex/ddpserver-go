package ddpserver

import (
	"sync"
)

type Emitter struct {
	// Mutex
	mu sync.Mutex

	// As we don't have constructor, we must init on first method call
	initialized bool

	// Map storing listener functions by event name
	listeners map[string][]func(string, ...interface{})
}

func (e *Emitter) init() {
	if !e.initialized {
		e.listeners = make(map[string][]func(string, ...interface{}))
		e.initialized = true
	}
}

func (e *Emitter) On(event string, fn func(string, ...interface{})) {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		e.listeners[event] = append(listeners, fn)
	} else {
		e.listeners[event] = []func(string, ...interface{}){fn}
	}
}

func (e *Emitter) Emit(event string, params ...interface{}) {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		for _, l := range listeners {
			l(event, params)
		}
	}
}
