package server

type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start() {

}
