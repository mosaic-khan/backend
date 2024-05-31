package Post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"main/internal/service/utils"
	"main/internal/storage/db"
	"main/pkg/PostAPIService"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetPost(ctx context.Context, in *PostAPIService.SetPostRequest) (*PostAPIService.SetPostResponse, error) {
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
		CategoryID:  int16(in.GetCategoryID()),
		NumImages:   int16(in.GetNumImages()),
		ProfileID:   profileId,
	})
	if err != nil {
		var driverErr *pq.Error
		errors.As(err, &driverErr)
		_ = tx.Rollback()
		if driverErr.Code == ("23503") {
			return nil, status.Errorf(codes.InvalidArgument, "category with id %d does not exists", in.GetCategoryID())
		} else {
			return nil, status.Errorf(codes.Internal, "could not create post")
		}
	}

	// insert post ingredients
	for ingredient, amount := range in.Ingredients {
		ingredientId, err := txQuery.GetIngredientId(ctx, ingredient)
		if errors.Is(err, sql.ErrNoRows) {
			// insert ingredient if not exists
			ingredientId, err = txQuery.InsertIngredient(ctx, ingredient)
			if err != nil {
				_ = tx.Rollback()
				return nil, status.Errorf(codes.Internal, "could not add ingredients")
			}
		} else if err != nil {
			_ = tx.Rollback()
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
			_ = tx.Rollback()
			return nil, status.Errorf(codes.Internal, "coult not insert post ingredients")
		}
	}

	return &PostAPIService.SetPostResponse{Id: postId}, nil

}

func (s *Server) GetPost(ctx context.Context, in *PostAPIService.GetPostRequest) (*PostAPIService.GetPostResponse, error) {

	profileID := ctx.Value("ProfileID").(int64)

	// get post
	post, err := s.query.GetPost(ctx, db.GetPostParams{PostID: in.PostID, ProfileID: profileID})
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

	response := &PostAPIService.GetPostResponse{
		Post: &PostAPIService.Post{
			Id:            post.ID,
			Title:         post.Title,
			Description:   post.Description,
			Category:      post.Category,
			NumImages:     int32(post.NumImages),
			NumLikes:      post.NumLikes,
			NumComments:   post.NumComments,
			Like:          post.Liked,
			Ingredients:   ingredientsMap,
			ImageUrls:     imageUrls,
			Username:      post.Username,
			ProfilePicUrl: post.ProfilePicAddress,
		},
	}

	pinned := new(bool)

	if profileID == post.ProfileID && !post.Pinned {
		*pinned = false
	} else if profileID == post.ProfileID && post.Pinned {
		*pinned = true
	} else {
		pinned = nil
	}
	response.Post.Pinned = pinned

	return response, nil
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

	if in.PageNumber == nil {
		in.PageNumber = new(int32)
		*in.PageNumber = 1
	}

	postsDB, err := s.query.GetPostsPreview(ctx, db.GetPostsPreviewParams{
		ProfileID:   in.ProfileID,
		ProfileID_2: profileID,
		Offset:      (in.GetPageNumber() - 1) * 20,
	})
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
			NumLikes:         post.NumLikes,
			NumComments:      post.NumComments,
			IsLiked:          post.Isliked,
		}
	}

	return &PostAPIService.GetProfilePostsResponse{PostPreview: posts, PageNumber: in.GetPageNumber()}, nil
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

	if len(tokenAud) < 1 || tokenAud[0] != "Media PostImage" {
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
	var driverErr *pq.Error
	if err != nil {
		errors.As(err, &driverErr)
	}
	if driverErr != nil && driverErr.Code == ("23503") { // error code 23503 = foreign_key_violation
		return nil, status.Errorf(codes.InvalidArgument, "post with id %d does not exists", in.GetPostId())
	} else if driverErr != nil && driverErr.Code == ("23505") { // error code 23505 = unique_violation
		return nil, status.Errorf(codes.InvalidArgument, "already liked post with id %d", in.GetPostId())
	} else if err != nil {
		return nil, status.Error(codes.Internal, "could not like post")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Dislike(ctx context.Context, in *PostAPIService.DislikeRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.DislikePost(ctx, db.DislikePostParams{ProfileID: profileID, PostID: in.GetPostId()})
	if err != nil {
		return nil, status.Error(codes.Internal, "could not dislike post")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) AddComment(ctx context.Context, in *PostAPIService.AddCommentRequest) (*PostAPIService.Comment, error) {
	profileID := ctx.Value("ProfileID").(int64)

	safe, err := utils.CommentClient.IsSafe(in.GetComment())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not check comment regarding swears")
	} else if !safe {
		return nil, status.Errorf(codes.PermissionDenied, "comment contains swear words")
	}

	id, err := s.query.AddComment(ctx, db.AddCommentParams{
		PostID:    in.PostID,
		ProfileID: profileID,
		Comment:   in.Comment,
	})
	var driverErr *pq.Error
	if err != nil {
		errors.As(err, &driverErr)
	}
	if driverErr != nil && driverErr.Code == ("23503") { // error code 23503 = foreign_key_violation
		return nil, status.Errorf(codes.InvalidArgument, "post with id %d does not exists", in.GetPostID())
	} else if err != nil {
		return nil, status.Error(codes.Internal, "error while adding comment")
	}

	comment, err := s.query.GetComment(ctx, db.GetCommentParams{ID: id, ProfileID: profileID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get comment with id %d", id)
	}

	return &PostAPIService.Comment{
		ID:         id,
		Name:       comment.FirstName,
		Username:   comment.Username,
		ProfileUrl: comment.ProfilePicAddress,
		Comment:    comment.Comment,
		Time:       comment.Time.GoString(),
		HasReplies: comment.HasReplies,
		IsLiked:    comment.Isliked,
		NumLikes:   comment.NumLikes,
		Owned:      comment.Owned,
	}, nil
}

func (s *Server) AddReply(ctx context.Context, in *PostAPIService.AddReplyRequest) (*PostAPIService.Comment, error) {
	profileID := ctx.Value("ProfileID").(int64)

	safe, err := utils.CommentClient.IsSafe(in.GetComment())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not check comment regarding swears")
	} else if !safe {
		return nil, status.Errorf(codes.PermissionDenied, "comment contains swear words")
	}

	id, err := s.query.AddReply(ctx, db.AddReplyParams{
		ParentID: sql.NullInt64{
			Int64: in.CommentID,
			Valid: true,
		},
		ProfileID: profileID,
		Comment:   in.Comment,
	})
	var driverErr *pq.Error
	if err != nil {
		errors.As(err, &driverErr)
	}
	if driverErr != nil && driverErr.Code == ("23503") { // error code 23503 = foreign_key_violation
		return nil, status.Errorf(codes.InvalidArgument, "comment or post does not exist")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "error while adding reply")
	}

	comment, err := s.query.GetComment(ctx, db.GetCommentParams{ID: id, ProfileID: profileID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get comment with id %d", id)
	}

	return &PostAPIService.Comment{
		ID:         id,
		Name:       comment.FirstName,
		Username:   comment.Username,
		ProfileUrl: comment.ProfilePicAddress,
		Comment:    comment.Comment,
		Time:       comment.Time.GoString(),
		HasReplies: comment.HasReplies,
		IsLiked:    comment.Isliked,
		NumLikes:   comment.NumLikes,
		Owned:      comment.Owned,
	}, nil
}

func (s *Server) GetComments(ctx context.Context, in *PostAPIService.GetCommentsRequest) (*PostAPIService.GetCommentsResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	commentsDB, err := s.query.GetPostsComments(ctx, db.GetPostsCommentsParams{
		ProfileID: profileID,
		PostID:    in.PostID,
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, status.Errorf(codes.Internal, "error while getting posts comments")
	}

	comments := make([]*PostAPIService.Comment, len(commentsDB))

	for i, comment := range commentsDB {
		comments[i] = &PostAPIService.Comment{
			ID:         comment.ID,
			Name:       comment.FirstName,
			Username:   comment.Username,
			ProfileUrl: comment.ProfilePicAddress,
			Comment:    comment.Comment,
			Time:       comment.Time.GoString(),
			HasReplies: comment.HasReplies,
			IsLiked:    comment.Isliked,
			NumLikes:   comment.NumLikes,
			Owned:      comment.Owned,
		}
	}

	return &PostAPIService.GetCommentsResponse{Comments: comments}, nil
}

func (s *Server) GetReplies(ctx context.Context, in *PostAPIService.GetRepliesRequest) (*PostAPIService.GetRepliesResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	commentsDB, err := s.query.GetReplies(ctx, db.GetRepliesParams{
		ProfileID: profileID,
		ParentID: sql.NullInt64{
			Int64: in.CommentID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting replies")
	}

	comments := make([]*PostAPIService.Comment, len(commentsDB))

	for i, comment := range commentsDB {
		comments[i] = &PostAPIService.Comment{
			ID:         comment.ID,
			Name:       comment.FirstName,
			Username:   comment.Username,
			ProfileUrl: comment.ProfilePicAddress,
			Comment:    comment.Comment,
			Time:       comment.Time.GoString(),
			HasReplies: comment.HasReplies,
			IsLiked:    comment.Isliked.(bool),
			NumLikes:   comment.NumLikes,
			Owned:      comment.Owned,
		}
	}

	return &PostAPIService.GetRepliesResponse{Comments: comments}, nil
}

func (s *Server) LikeComment(ctx context.Context, in *PostAPIService.LikeCommentRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.LikeCommentOrReply(ctx, db.LikeCommentOrReplyParams{
		ProfileID: profileID,
		CommentID: in.CommentID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while liking comment")
	}

	return nil, nil
}

func (s *Server) DislikeComment(ctx context.Context, in *PostAPIService.DislikeCommentRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.DislikeCommentOrReply(ctx, db.DislikeCommentOrReplyParams{
		ProfileID: profileID,
		CommentID: in.CommentID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while disliking comment")
	}

	return nil, nil
}

func (s *Server) GetCategories(ctx context.Context, _ *emptypb.Empty) (*PostAPIService.GetCategoriesResponse, error) {
	categoriesDB, err := s.query.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get categories")
	}

	categories := make([]*PostAPIService.Category, len(categoriesDB))

	for i, c := range categoriesDB {
		temp := &PostAPIService.Category{Id: int32(c.ID), Name: c.Name, Level: int32(c.Level)}
		if c.Parent.Valid {
			p := int32(c.Parent.Int16)
			temp.Parent = &p
		} else {
			temp.Parent = nil
		}
		categories[i] = temp
	}

	return &PostAPIService.GetCategoriesResponse{Categories: categories}, nil
}

func (s *Server) ReportComment(ctx context.Context, in *PostAPIService.RepostCommentRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.ReportComment(ctx, db.ReportCommentParams{
		ProfileID: profileID,
		CommentID: in.Id,
	})
	var driverErr *pq.Error
	if err != nil {
		errors.As(err, &driverErr)
	}

	if driverErr != nil && driverErr.Code == ("23505") {
		return nil, status.Errorf(codes.AlreadyExists, "unique_violation")
	}

	return nil, nil
}

func (s *Server) PinPost(ctx context.Context, in *PostAPIService.PinPostRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	p, err := s.query.GetPostProfileId(ctx, in.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not pin post, cant find post owner")
	}
	if p != profileID {
		return nil, status.Errorf(codes.InvalidArgument, "Don't have permission to pin post with id %d", in.GetId())
	}

	c, err := s.query.GetPinsCount(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not pin post")
	}
	if c > 3 {
		return nil, status.Errorf(codes.InvalidArgument, "Can not pin posts any more, maximum number of pins 3")
	}

	err = s.query.AddPin(ctx, db.AddPinParams{ProfileID: profileID, PostID: in.GetId()})
	if err != nil {

		var driverErr *pq.Error
		if errors.As(err, &driverErr) && driverErr.Code == "23505" { // error code 23505 = unique_violation
			return nil, status.Errorf(codes.InvalidArgument, "post with id %d is already pinned", in.GetId())
		}

		return nil, status.Errorf(codes.Internal, "Could not pin post")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UnpinPost(ctx context.Context, in *PostAPIService.UnpinPostRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.RemovePin(ctx, db.RemovePinParams{ProfileID: profileID, PostID: in.GetId()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not unpin post with id %d", in.GetId())
	}

	return &emptypb.Empty{}, nil

}

func (s *Server) GetPins(ctx context.Context, in *emptypb.Empty) (*PostAPIService.GetPinsResponse, error) {

	profileID := ctx.Value("ProfileID").(int64)

	pins, err := s.query.GetPins(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retreive pined post")
	}

	getPinsReq := &PostAPIService.GetPinsResponse{}

	for _, p := range pins {
		getPinsReq.PinedPost = append(getPinsReq.PinedPost, &PostAPIService.PinedPost{Title: p.Title, ImageUrl: p.ImageUrl.String})
	}

	return getPinsReq, nil

}

func (s *Server) DeleteComment(ctx context.Context, in *PostAPIService.DeleteCommentRequest) (*emptypb.Empty, error) {

	profileID := ctx.Value("ProfileID").(int64)

	err := s.query.DeleteComment(ctx, db.DeleteCommentParams{ProfileID: profileID, ID: in.Id})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not delete comment with id %d", in.Id)
	}

	return &emptypb.Empty{}, nil
}
