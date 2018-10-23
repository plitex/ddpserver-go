package ddpserver

import (
	"log"
	"net/http"
)

type Server struct {
	Emitter

	id string

	running bool

	// Registered sessions.
	sessions map[*Session]bool

	// Register requests from the connections.
	register chan *Session

	// Unregister requests from connections.
	unregister chan *Session

	// Method handlers
	methods map[string]MethodHandler

	// Publication handlers
	publications map[string]PublicationHandler
}

func NewServer() *Server {
	return &Server{
		id:           "0",
		register:     make(chan *Session),
		unregister:   make(chan *Session),
		sessions:     make(map[*Session]bool),
		methods:      make(map[string]MethodHandler),
		publications: make(map[string]PublicationHandler),
	}
}

func (s *Server) Run() {
	if s.running == false {
		go func() {
			s.running = true
			for {
				select {
				case conn := <-s.register:
					s.sessions[conn] = true
					s.Emit("connected")
				case conn := <-s.unregister:
					if _, ok := s.sessions[conn]; ok {
						delete(s.sessions, conn)
						close(conn.send)
						s.Emit("disconnected")
					}
				}
			}
		}()
	}
}

// Method registers new method
func (s *Server) Method(name string, fn MethodHandler) {
	s.methods[name] = fn
}

// Publish registers new publication
func (s *Server) Publish(name string, fn PublicationHandler) {
	s.publications[name] = fn
}

// ServeWs handles websocket requests from the peer.
func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	session := newSession(s, socket)
	s.register <- session
}
