package Search

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"main/internal/storage/db"
	"main/pkg/SearchAPIService"
)

func (s *Server) GetAllCategories(ctx context.Context, _ *emptypb.Empty) (*SearchAPIService.GetAllCategoriesResponse, error) {
	categoriesDB, err := s.query.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get categories")
	}

	categories := make([]*SearchAPIService.Categories, len(categoriesDB))

	for i, c := range categoriesDB {
		temp := &SearchAPIService.Categories{Id: int32(c.ID), Name: c.Name, Level: int32(c.Level)}
		if c.Parent.Valid {
			p := int32(c.Parent.Int16)
			temp.Parent = &p
		} else {
			temp.Parent = nil
		}
		categories[i] = temp
	}

	return &SearchAPIService.GetAllCategoriesResponse{Categories: categories}, nil
}

func (s *Server) SearchCategories(ctx context.Context, in *SearchAPIService.SearchCategoriesRequest) (*SearchAPIService.SearchCategoriesResponse, error) {
	if in.PageNumber == nil {
		in.PageNumber = new(int32)
		*in.PageNumber = 1
	}

	postsDB, err := s.query.GetPostsWithCategory(ctx, db.GetPostsWithCategoryParams{
		Column1: in.GetCategoryID(),
		Offset:  (in.GetPageNumber() - 1) * 20,
	})
	if err != nil {
		log.Println(err.Error())
		return nil, status.Error(codes.Internal, "could not get posts")
	}

	var posts []*SearchAPIService.PostPreviewExplore

	for _, p := range postsDB {
		posts = append(posts, &SearchAPIService.PostPreviewExplore{
			Id:               p.ID,
			Title:            p.Title,
			ShortDescription: p.Description,
			PostImage:        p.PostImage.String,
			Username:         p.Username,
			ProfilePicUrl:    p.ProfilePicAddress,
		})
	}

	return &SearchAPIService.SearchCategoriesResponse{Posts: posts, PageNumber: in.GetPageNumber()}, nil

}

func (s *Server) SearchFoodByName(ctx context.Context, in *SearchAPIService.SearchFoodByNameRequest) (*SearchAPIService.SearchFoodByNameResponse, error) {
	if in.PageNumber == nil {
		in.PageNumber = new(int32)
		*in.PageNumber = 1
	}

	searchResult, err := s.query.SearchName(ctx, db.SearchNameParams{
		Name: in.Name,
		Page: (in.GetPageNumber() - 1) * 20,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while searching for posts")
	}

	posts := make([]*SearchAPIService.PostPreviewExplore, len(searchResult))

	for i, row := range searchResult {
		posts[i] = &SearchAPIService.PostPreviewExplore{
			Id:               row.ID,
			Title:            row.Title,
			ShortDescription: row.Description,
			PostImage:        row.ImageUrl.String,
			Username:         row.Username,
			ProfilePicUrl:    row.ProfilePicAddress,
		}
	}

	return &SearchAPIService.SearchFoodByNameResponse{PostPreview: posts, PageNumber: in.GetPageNumber()}, nil
}

func (s *Server) SearchFoodByIngredient(ctx context.Context, in *SearchAPIService.SearchFoodByIngredientRequest) (*SearchAPIService.SearchFoodByIngredientResponse, error) {
	if in.PageNumber == nil {
		in.PageNumber = new(int32)
		*in.PageNumber = 1
	}

	searchResult, err := s.query.SearchIngredient(ctx, db.SearchIngredientParams{
		Page:       (in.GetPageNumber() - 1) * 20,
		Include:    in.Include,
		Includecnt: int32(len(in.Include)),
		Exclude:    in.Exclude,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while searching for posts")
	}

	posts := make([]*SearchAPIService.PostPreviewExplore, len(searchResult))

	for i, row := range searchResult {
		posts[i] = &SearchAPIService.PostPreviewExplore{
			Id:               row.ID,
			Title:            row.Title,
			ShortDescription: row.Description,
			PostImage:        row.ImageUrl.String,
			Username:         row.Username,
			ProfilePicUrl:    row.ProfilePicAddress,
		}
	}

	return &SearchAPIService.SearchFoodByIngredientResponse{PostPreview: posts, PageNumber: in.GetPageNumber()}, nil
}

func (s *Server) SearchUsername(ctx context.Context, in *SearchAPIService.SearchUsernameRequest) (*SearchAPIService.SearchUsernameResponse, error) {

	profiles, err := s.query.SearchUsername(ctx, in.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error retrieving usernames smiliar to %s\n", in.GetUsername())
	}

	var r []*SearchAPIService.ProfilePreviewExplore

	for _, p := range profiles {
		r = append(r, &SearchAPIService.ProfilePreviewExplore{ProfileID: p.ID, Username: p.Username, Name: p.FirstName, ProfilePicUrl: p.ProfilePicAddress})
	}

	return &SearchAPIService.SearchUsernameResponse{ProfilePreview: r}, nil
}

func (s *Server) MixedSearch(ctx context.Context, in *SearchAPIService.MixedSearchRequest) (*SearchAPIService.MixedSearchResponse, error) {
	if in.PageNumber == nil {
		in.PageNumber = new(int32)
		*in.PageNumber = 1
	}

	searchResult, err := s.query.MixedSearch(ctx, db.MixedSearchParams{
		Page:       (in.GetPageNumber() - 1) * 20,
		Includecnt: len(in.IncludeIng),
		Include:    in.IncludeIng,
		Exclude:    in.ExcludeIng,
		Categories: in.CategoryID,
		Name:       in.Name,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while mixed searching in db")
	}

	posts := make([]*SearchAPIService.PostPreviewExplore, len(searchResult))

	for i, row := range searchResult {
		posts[i] = &SearchAPIService.PostPreviewExplore{
			Id:               row.ID,
			Title:            row.Title,
			ShortDescription: row.Description,
			PostImage:        row.ImageUrl.String,
			Username:         row.Username,
			ProfilePicUrl:    row.ProfilePicAddress,
		}
	}

	return &SearchAPIService.MixedSearchResponse{
		Posts:      posts,
		PageNumber: in.GetPageNumber(),
	}, nil
}
