package Post

import (
	"context"
	"database/sql"
	"main/internal/storage/db"
	"main/pkg/PostAPIService"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetPost(ctx context.Context, in *PostAPIService.SetPostRequest) (*emptypb.Empty, error) {

	// get profile id
	// TODO: proper value
	profileId := ctx.Value("ProfileID").(int64)

	tx, err := s.conn.Begin()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not begin transaction")
	}

	// check post constraints
	// check title
	if in.Post.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post should have a title")
	}
	// check num images
	if in.Post.NumImages < 1 && in.Post.NumImages > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "post should have a least one image but no more than ten")
	}

	txQuery := s.query.WithTx(tx)
	defer func(tx *sql.Tx) {
		_ = tx.Commit()
	}(tx)

	// insert post
	postId, err := txQuery.InsertPost(ctx, db.InsertPostParams{
		Title:       in.Post.GetTitle(),
		Description: in.Post.GetDescription(),
		NumImages:   int16(in.Post.GetNumImages()),
	})
	if err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "could not create post")
	}

	// post and profile relation
	err = txQuery.InsertProfilerHasPost(ctx, db.InsertProfilerHasPostParams{
		ProfileID: profileId,
		PostID:    postId,
	})
	if err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "could not create post")
	}

	return &emptypb.Empty{}, nil

}
