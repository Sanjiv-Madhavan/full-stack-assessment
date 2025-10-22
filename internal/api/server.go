package api

import "context"

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) GetHealth(ctx context.Context, req GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}
