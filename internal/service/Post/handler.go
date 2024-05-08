package Post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main/internal/storage/db"
	"main/pkg/PostAPIService"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

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
	if in.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post should have a title")
	}
	// check num images
	if in.GetNumImages() < 1 && in.GetNumImages() > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "post should have a least one image but no more than ten")
	}

	txQuery := s.query.WithTx(tx)
	defer func(tx *sql.Tx) {
		_ = tx.Commit()
	}(tx)

	// insert post
	postId, err := txQuery.InsertPost(ctx, db.InsertPostParams{
		Title:       in.GetTitle(),
		Description: in.GetDescription(),
		NumImages:   int16(in.GetNumImages()),
		ProfileID:   profileId,
	})
	if err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "could not create post")
	}

	// insert post ingredients
	for ingredient, amount := range in.Ingredients {
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

	profileID := ctx.Value("ProfileID").(int64)

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

	imageUrls, err := s.query.GetPostImages(ctx, in.PostID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get post images")
	}

	// check if profile has liked the post or not
	l, err := s.query.ProfileLikePost(ctx, db.ProfileLikePostParams{ProfileID: profileID, PostID: in.GetPostID()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not check if user has liked the post or not")
	}

	var like bool
	if l > 0 {
		like = true
	} else {
		like = false
	}

	return &PostAPIService.GetPostResponse{
		Post: &PostAPIService.Post{
			Id:            post.ID,
			Title:         post.Title,
			Description:   post.Description,
			NumImages:     int32(post.NumImages),
			NumLikes:      post.NumLikes,
			Like:          like,
			Ingredients:   ingredientsMap,
			ImageUrls:     imageUrls,
			Username:      post.Username,
			ProfilePicUrl: post.ProfilePicAddress,
		},
	}, nil

}

func (s *Server) SuggestIngredient(ctx context.Context, in *PostAPIService.SuggestIngredientRequest) (*PostAPIService.SuggestIngredientResponse, error) {

	suggestions, err := s.query.GetSimilarIngredient(ctx, sql.NullString{String: in.GetName(), Valid: true})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get similar ingredients")
	}

	return &PostAPIService.SuggestIngredientResponse{
		Ingredients: suggestions,
	}, nil

}

func (s *Server) GetProfilePosts(ctx context.Context, in *PostAPIService.GetProfilePostsRequests) (*PostAPIService.GetProfilePostsResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	postsDB, err := s.query.GetPostsPreview(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error fetching posts")
	}

	posts := make([]*PostAPIService.PostPreview, len(postsDB))

	for i, post := range postsDB {
		posts[i] = &PostAPIService.PostPreview{
			Id:               post.ID,
			Title:            post.Title,
			ShortDescription: post.Description,
			Image:            post.ImageUrl.String,
		}
	}

	return &PostAPIService.GetProfilePostsResponse{PostPreview: posts}, nil
}

func (s *Server) AddImageForPost(ctx context.Context, in *PostAPIService.AddImageForPostRequest) (*emptypb.Empty, error) {

	profileID := ctx.Value("ProfileID").(int64)

	// Validate token
	token, err := jwt.Parse(in.PostImageToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.hmacSecret, nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	tokenProfileIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return nil, status.Error(codes.Internal, "error while extracting profileID")
	}

	tokenProfileID, err := strconv.ParseInt(tokenProfileIDStr, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Internal, "error while converting userID")
	}

	if profileID != tokenProfileID {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated user")
	}

	tokenAud, err := token.Claims.GetAudience()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting token Aud")
	}

	if tokenAud[0] != "Media PostImage" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	postIDStr := tokenAud[1]
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error line 205")
	}

	owner, err := s.query.GetPostOwnerProfile(ctx, postID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}
	if owner != profileID {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	filepath := tokenAud[2]
	err = s.query.AddImage(ctx, db.AddImageParams{
		PostID:   postID,
		ImageUrl: fmt.Sprintf("/KhanAPI.MediaAPI/images/%s", filepath),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return nil, nil
}

func (s *Server) Like(ctx context.Context, in *PostAPIService.LikeRequest) (*emptypb.Empty, error) {

	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.LikePost(ctx, db.LikePostParams{ProfileID: profileID, PostID: in.GetPostId()})
	if err != nil {
		return nil, status.Error(codes.Internal, "could not like post")
	}

	return &emptypb.Empty{}, nil
}
