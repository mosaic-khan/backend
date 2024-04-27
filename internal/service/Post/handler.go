package Post

import (
	"context"
	"main/pkg/PostAPIService"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetPost(ctx context.Context, in *PostAPIService.SetPostRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
