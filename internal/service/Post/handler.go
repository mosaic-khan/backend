package Post

import (
	"context"
	"database/sql"
	"errors"
	"main/internal/storage/db"
	"main/pkg/PostAPIService"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetPost(ctx context.Context, in *PostAPIService.SetPostRequest) (*emptypb.Empty, error) {

	// get profile id
	profileId := ctx.Value("ProfileID").(int64)

	tx, err := s.conn.Begin()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not begin transaction")
	}

	// check post constraints
	// check title
	if in.Post.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post should have a title")
	}
	// check num images
	if in.Post.GetNumImages() < 1 && in.Post.GetNumImages() > 10 {
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
		ProfileID:   profileId,
	})
	if err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "could not create post")
	}

	// insert post ingredients
	for ingredient, amount := range in.Post.Ingredients {
		ingredientId, err := txQuery.GetIngredientId(ctx, ingredient)
		if errors.Is(err, sql.ErrNoRows) {
			// insert ingredient if not exists
			ingredientId, err = txQuery.InsertIngredient(ctx, ingredient)
			if err != nil {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "could not add ingredients")
			}
		} else if err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "could not add ingredients")
		}

		err = txQuery.InsertPostHasIngredient(ctx, db.InsertPostHasIngredientParams{
			PostID:       postId,
			IngredientID: ingredientId,
			Amount: sql.NullString{
				String: amount,
				Valid:  true,
			},
		})
		if err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "coult not insert post ingredients")
		}
	}

	return &emptypb.Empty{}, nil

}

func (s *Server) GetPost(ctx context.Context, in *PostAPIService.GetPostRequest) (*PostAPIService.GetPostResponse, error) {

	// get post
	post, err := s.query.GetPost(ctx, in.GetPostID())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Errorf(codes.InvalidArgument, "post id %d doesn't exist\n", in.GetPostID())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get post with id %d\n", in.GetPostID())
	}

	// get post ingredients
	ingredients, err := s.query.GetPostIngredient(ctx, in.GetPostID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get ingredients of post with id %d\n", in.GetPostID())
	}

	ingredientsMap := make(map[string]string)
	for _, i := range ingredients {
		ingredientsMap[i.Name] = i.Amount.String
	}

	return &PostAPIService.GetPostResponse{
		Post: &PostAPIService.Post{
			Title:       post.Title,
			Description: post.Description,
			NumImages:   int32(post.NumImages),
			Ingredients: ingredientsMap,
		},
	}, nil

}

func (s *Server) SuggestIngredient(ctx context.Context, in *PostAPIService.SuggestIngredientRequest) (*PostAPIService.SuggestIngredientResponse, error) {

	suggestions, err := s.query.GetSimilarIngredient(ctx, sql.NullString{String: in.GetName(), Valid: true})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get similar ingredients")
	}

	return &PostAPIService.SuggestIngredientResponse{
		Ingerdients: suggestions,
	}, nil

}
