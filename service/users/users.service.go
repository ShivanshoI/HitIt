package users

import (
	"context"
	"pog/database/users"
	"pog/internal"
	"time"
)

type UserService struct {
	repo *users.UserRepository
}

func NewUserService(repo *users.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// SignIn authenticates a user with an identifier (email/phone) and password.
func (s *UserService) SignIn(ctx context.Context, identifier, password string) (*SignInResponse, error) {
	user, err := s.repo.GetByIdentifierAndPassword(ctx, identifier, password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, internal.NewUnauthorized("Invalid email or password.")
	}

	userResp := UserResponse{
		ID:        user.ID.Hex(),
		FirstName: user.FirstName,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastName != nil {
		userResp.LastName = *user.LastName
	}
	if user.PhoneNumber != nil {
		userResp.PhoneNumber = *user.PhoneNumber
	}
	if user.EmailAddress != nil {
		userResp.EmailAddress = *user.EmailAddress
	}
	if user.NickName != nil {
		userResp.NickName = *user.NickName
	}
	if user.OrganizationID != nil {
		orgID := user.OrganizationID.Hex()
		userResp.OrganizationID = &orgID
	}

	token, err := internal.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, internal.NewInternalError("Failed to generate token")
	}

	return &SignInResponse{
		User:  userResp,
		Token: token,
	}, nil
}

// SignUp creates a new user account.
func (s *UserService) SignUp(ctx context.Context, payload *users.User) (*SignInResponse, error) {
	identifier := ""
	if payload.EmailAddress != nil && *payload.EmailAddress != "" {
		identifier = *payload.EmailAddress
	} else if payload.PhoneNumber != nil && *payload.PhoneNumber != "" {
		identifier = *payload.PhoneNumber
	}

	if identifier != "" {
		exists, _ := s.repo.ExistsByIdentifier(ctx, identifier)
		if exists {
			return nil, internal.NewBadRequest("User with this email or phone already exists")
		}
	}

	user, err := s.repo.Create(ctx, payload)
	if err != nil {
		return nil, internal.NewInternalError("Failed to create user account")
	}

	userResp := UserResponse{
		ID:        user.ID.Hex(),
		FirstName: user.FirstName,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastName != nil {
		userResp.LastName = *user.LastName
	}
	if user.NickName != nil {
		userResp.NickName = *user.NickName
	}
	if user.PhoneNumber != nil {
		userResp.PhoneNumber = *user.PhoneNumber
	}
	if user.EmailAddress != nil {
		userResp.EmailAddress = *user.EmailAddress
	}
	if user.NickName != nil {
		userResp.NickName = *user.NickName
	}
	if user.OrganizationID != nil {
		orgID := user.OrganizationID.Hex()
		userResp.OrganizationID = &orgID
	}

	// 4. Generate Token
	token, err := internal.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, internal.NewInternalError("Failed to generate token")
	}

	return &SignInResponse{
		User:  userResp,
		Token: token,
	}, nil
}

// GetMe returns the currently authenticated user's profile.
func (s *UserService) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userResp := UserResponse{
		ID:        user.ID.Hex(),
		FirstName: user.FirstName,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastName != nil {
		userResp.LastName = *user.LastName
	}
	if user.PhoneNumber != nil {
		userResp.PhoneNumber = *user.PhoneNumber
	}
	if user.EmailAddress != nil {
		userResp.EmailAddress = *user.EmailAddress
	}
	if user.NickName != nil {
		userResp.NickName = *user.NickName
	}
	if user.OrganizationID != nil {
		orgID := user.OrganizationID.Hex()
		userResp.OrganizationID = &orgID
	}

	return &userResp, nil
}
