package user

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	testSupportAddress = mail.Address("support@example.com")
	testAdminAddress   = mail.Address("admin@example.com")
	testWebBaseURL     = &url.URL{Scheme: "https", Host: "web.example.com"}
)

type fakeRepository struct {
	user          *user.User
	getErr        error
	getByEmailErr error
	createErr     error
	updated       *user.User
	created       *user.User
	replaced      *user.User
	verified      *user.User
	deletedBefore time.Time
}

func (r *fakeRepository) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return r.user, r.getErr
}

func (r *fakeRepository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	return r.user, r.getByEmailErr
}

func (r *fakeRepository) Create(ctx context.Context, u *user.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	u.ID = user.ID(1)
	r.created = u
	return nil
}

func (r *fakeRepository) ReplaceUnverified(ctx context.Context, u *user.User) error {
	r.replaced = u
	return nil
}

func (r *fakeRepository) UpdateEmailVerifiedAt(ctx context.Context, u *user.User) error {
	r.verified = u
	return nil
}

func (r *fakeRepository) DeleteUnverifiedCreatedBefore(ctx context.Context, before time.Time) error {
	r.deletedBefore = before
	return nil
}

type fakeEmailVerificationTokenRepo struct {
	token          *user.EmailVerificationToken
	findErr        error
	latestToken    *user.EmailVerificationToken
	created        *user.EmailVerificationToken
	deletedUserIDs []user.ID
}

func (r *fakeEmailVerificationTokenRepo) Create(
	ctx context.Context,
	userID user.ID,
) (*user.EmailVerificationToken, error) {
	r.created = &user.EmailVerificationToken{
		UserID: userID,
		Value:  user.EmailVerificationTokenValue("testtokenvalue"),
	}
	return r.created, nil
}

func (r *fakeEmailVerificationTokenRepo) FindValidByValue(
	ctx context.Context,
	value user.EmailVerificationTokenValue,
) (*user.EmailVerificationToken, error) {
	return r.token, r.findErr
}

func (r *fakeEmailVerificationTokenRepo) FindLatestByUserID(
	ctx context.Context,
	userID user.ID,
) (*user.EmailVerificationToken, error) {
	if r.latestToken == nil {
		return nil, exception.ErrNotFound
	}
	return r.latestToken, nil
}

func (r *fakeEmailVerificationTokenRepo) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	r.deletedUserIDs = append(r.deletedUserIDs, userID)
	return nil
}

type fakeMailService struct {
	sent []mail.Message
	err  error
}

func (s *fakeMailService) Send(ctx context.Context, m mail.Message) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, m)
	return nil
}

func (r *fakeRepository) UpdateUsername(ctx context.Context, u *user.User) error {
	r.updated = u
	return nil
}

func (r *fakeRepository) UpdateEmail(ctx context.Context, u *user.User) error {
	r.updated = u
	return nil
}

func (r *fakeRepository) UpdatePassword(ctx context.Context, u *user.User) error {
	r.updated = u
	return nil
}

func (r *fakeRepository) Delete(ctx context.Context, userID user.ID) error {
	return nil
}

type fakeRefreshTokenRepo struct {
	revokedUserID user.ID
}

func (r *fakeRefreshTokenRepo) Create(
	ctx context.Context,
	principal *auth.Principal,
) (*auth.RefreshToken, error) {
	return nil, nil
}

func (r *fakeRefreshTokenRepo) FindValidByValue(
	ctx context.Context,
	value auth.RefreshTokenValue,
) (*auth.RefreshToken, error) {
	return nil, nil
}

func (r *fakeRefreshTokenRepo) Revoke(
	ctx context.Context,
	id auth.RefreshTokenID,
) error {
	return nil
}

func (r *fakeRefreshTokenRepo) RevokeAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	r.revokedUserID = userID
	return nil
}

func (r *fakeRefreshTokenRepo) DeleteExpired(
	ctx context.Context,
) error {
	return nil
}

type fakeMediaService struct {
	deletedAllUserID user.ID
	err              error
}

func (s *fakeMediaService) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	return "", s.err
}

func (s *fakeMediaService) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	return s.err
}

func (s *fakeMediaService) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	s.deletedAllUserID = userID
	return s.err
}

func hashPassword(t *testing.T, password user.Password) user.HashedPassword {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf(
			"GenerateFromPassword(%v, %v) = %v",
			password, bcrypt.DefaultCost, err,
		)
	}
	return user.HashedPassword(hashed)
}

func TestCreate(t *testing.T) {
	username := user.Username("Test")
	email := user.Email("test@example.com")
	password := user.Password("testpassword")
	language := mail.LanguageJA

	t.Run(
		"creates the user with the hashed password and sends a verification mail when the email is new",
		func(t *testing.T) {
			repo := &fakeRepository{getByEmailErr: exception.ErrNotFound}
			tokenRepo := &fakeEmailVerificationTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Create(ctx, username, email, password, language); err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) error = %v",
					username, email, password, language, err,
				)
			}

			got := repo.created
			if got == nil {
				t.Fatalf("Create(ctx, %v, %v, %v, %v) did not create a user", username, email, password, language)
			}
			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					username, email, password, language, got.HashedPassword, password,
				)
			}
			if got.Username != username || got.Email != email || got.Role != user.RoleUser || got.IsEmailVerified() {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) created %v, want an unverified user with role %v",
					username, email, password, language, got, user.RoleUser,
				)
			}

			if tokenRepo.created == nil || tokenRepo.created.UserID != got.ID {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) created token %v, want a token for user %v",
					username, email, password, language, tokenRepo.created, got.ID,
				)
			}

			if len(mailService.sent) != 1 {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) sent %d mails, want 1",
					username, email, password, language, len(mailService.sent),
				)
			}
			m := mailService.sent[0]
			wantURL := "https://web.example.com/ja/verify-email?token=testtokenvalue"
			wantLogoURL := "https://web.example.com/mail/logo.png"
			if m.From != testSupportAddress ||
				m.To != mail.Address(email) ||
				m.Subject != mail.Subject("【Cinelier】メールアドレスの確認") ||
				!strings.Contains(string(m.HTMLBody), wantURL) ||
				!strings.Contains(string(m.HTMLBody), wantLogoURL) ||
				!strings.Contains(string(m.TextBody), wantURL) {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) sent %v, want a verification mail from %v to %v with %v and the logo %v",
					username, email, password, language, m, testSupportAddress, email, wantURL, wantLogoURL,
				)
			}
		},
	)

	t.Run(
		"sends an already registered mail without creating a user when the email is verified",
		func(t *testing.T) {
			verifiedAt := time.Now()
			existing := &user.User{
				ID:              user.ID(1),
				Email:           email,
				EmailVerifiedAt: &verifiedAt,
			}
			repo := &fakeRepository{user: existing}
			tokenRepo := &fakeEmailVerificationTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()
			language := mail.LanguageEN

			if err := uu.Create(ctx, username, email, password, language); err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) error = %v",
					username, email, password, language, err,
				)
			}

			if repo.created != nil || repo.replaced != nil || tokenRepo.created != nil {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) created %v, replaced %v and created token %v, want none",
					username, email, password, language, repo.created, repo.replaced, tokenRepo.created,
				)
			}

			if len(mailService.sent) != 1 {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) sent %d mails, want 1",
					username, email, password, language, len(mailService.sent),
				)
			}
			m := mailService.sent[0]
			wantURL := "https://web.example.com/en/login"
			if m.To != mail.Address(email) ||
				m.Subject != mail.Subject("[Cinelier] This email address is already registered") ||
				!strings.Contains(string(m.TextBody), wantURL) {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) sent %v, want an already registered mail to %v with %v",
					username, email, password, language, m, email, wantURL,
				)
			}
		},
	)

	t.Run(
		"replaces the unverified user and sends a verification mail when the email is not verified",
		func(t *testing.T) {
			existing := &user.User{
				ID:             user.ID(5),
				Username:       user.Username("Old"),
				Email:          email,
				HashedPassword: user.HashedPassword("oldhash"),
				Role:           user.RoleUser,
			}
			repo := &fakeRepository{user: existing}
			tokenRepo := &fakeEmailVerificationTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Create(ctx, username, email, password, language); err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) error = %v",
					username, email, password, language, err,
				)
			}

			got := repo.replaced
			if got == nil || repo.created != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) replaced %v and created %v, want only a replacement",
					username, email, password, language, got, repo.created,
				)
			}
			if got.ID != existing.ID || got.Username != username {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) replaced with %v, want ID %v and Username %v",
					username, email, password, language, got, existing.ID, username,
				)
			}
			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(password),
			); err != nil {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					username, email, password, language, got.HashedPassword, password,
				)
			}

			if len(tokenRepo.deletedUserIDs) != 1 || tokenRepo.deletedUserIDs[0] != existing.ID {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) deleted tokens of %v, want %v",
					username, email, password, language, tokenRepo.deletedUserIDs, existing.ID,
				)
			}
			if len(mailService.sent) != 1 ||
				mailService.sent[0].Subject != mail.Subject("【Cinelier】メールアドレスの確認") {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) sent %v, want one verification mail",
					username, email, password, language, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"returns nil without sending a mail when another request created the user first",
		func(t *testing.T) {
			repo := &fakeRepository{
				getByEmailErr: exception.ErrNotFound,
				createErr:     exception.ErrAlreadyExists,
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Create(ctx, username, email, password, language); err != nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) error = %v, want nil",
					username, email, password, language, err,
				)
			}
			if len(mailService.sent) != 0 {
				t.Errorf(
					"Create(ctx, %v, %v, %v, %v) sent %v, want no mail",
					username, email, password, language, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"returns an error when sending the verification mail fails",
		func(t *testing.T) {
			repo := &fakeRepository{getByEmailErr: exception.ErrNotFound}
			mailService := &fakeMailService{err: errors.New("send mail")}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Create(ctx, username, email, password, language); err == nil {
				t.Fatalf(
					"Create(ctx, %v, %v, %v, %v) error = nil, want an error",
					username, email, password, language,
				)
			}
		},
	)
}

func TestVerifyEmail(t *testing.T) {
	value := user.EmailVerificationTokenValue("testtokenvalue")
	userID := user.ID(1)

	t.Run(
		"marks the user as verified, deletes the tokens and notifies the admin",
		func(t *testing.T) {
			repo := &fakeRepository{user: &user.User{ID: userID, Email: user.Email("test@example.com")}}
			tokenRepo := &fakeEmailVerificationTokenRepo{
				token: &user.EmailVerificationToken{UserID: userID},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmail(ctx, value); err != nil {
				t.Fatalf("VerifyEmail(ctx, %v) error = %v", value, err)
			}

			if repo.verified == nil || repo.verified.ID != userID || !repo.verified.IsEmailVerified() {
				t.Errorf(
					"VerifyEmail(ctx, %v) updated %v, want user %v to be verified",
					value, repo.verified, userID,
				)
			}
			if len(tokenRepo.deletedUserIDs) != 1 || tokenRepo.deletedUserIDs[0] != userID {
				t.Errorf(
					"VerifyEmail(ctx, %v) deleted tokens of %v, want %v",
					value, tokenRepo.deletedUserIDs, userID,
				)
			}
			if len(mailService.sent) != 1 ||
				mailService.sent[0].To != testAdminAddress ||
				strings.Contains(string(mailService.sent[0].TextBody), "test@example.com") {
				t.Errorf(
					"VerifyEmail(ctx, %v) sent %v, want one mail to %v without the user's email",
					value, mailService.sent, testAdminAddress,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is invalid or expired",
		func(t *testing.T) {
			repo := &fakeRepository{}
			tokenRepo := &fakeEmailVerificationTokenRepo{findErr: exception.ErrInvalid}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			err := uu.VerifyEmail(ctx, value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf("VerifyEmail(ctx, %v) error = %v, want %v", value, err, exception.ErrInvalid)
			}
			if repo.verified != nil || len(mailService.sent) != 0 {
				t.Errorf(
					"VerifyEmail(ctx, %v) updated %v and sent %v, want neither",
					value, repo.verified, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"does not notify the admin again when the user is already verified",
		func(t *testing.T) {
			verifiedAt := time.Now()
			repo := &fakeRepository{user: &user.User{ID: userID, EmailVerifiedAt: &verifiedAt}}
			tokenRepo := &fakeEmailVerificationTokenRepo{
				token: &user.EmailVerificationToken{UserID: userID},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmail(ctx, value); err != nil {
				t.Fatalf("VerifyEmail(ctx, %v) error = %v", value, err)
			}
			if repo.verified != nil || len(mailService.sent) != 0 {
				t.Errorf(
					"VerifyEmail(ctx, %v) updated %v and sent %v, want neither",
					value, repo.verified, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"succeeds even when notifying the admin fails",
		func(t *testing.T) {
			repo := &fakeRepository{user: &user.User{ID: userID}}
			tokenRepo := &fakeEmailVerificationTokenRepo{
				token: &user.EmailVerificationToken{UserID: userID},
			}
			mailService := &fakeMailService{err: errors.New("send mail")}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmail(ctx, value); err != nil {
				t.Fatalf("VerifyEmail(ctx, %v) error = %v, want nil", value, err)
			}
			if repo.verified == nil {
				t.Errorf("VerifyEmail(ctx, %v) did not verify the user", value)
			}
		},
	)
}

func TestResendVerificationEmail(t *testing.T) {
	email := user.Email("test@example.com")
	language := mail.LanguageJA
	userID := user.ID(1)

	t.Run(
		"sends a new verification mail when the user is not verified",
		func(t *testing.T) {
			repo := &fakeRepository{user: &user.User{ID: userID, Email: email}}
			tokenRepo := &fakeEmailVerificationTokenRepo{
				latestToken: &user.EmailVerificationToken{
					UserID:    userID,
					CreatedAt: time.Now().Add(-user.EmailVerificationResendInterval - time.Second),
				},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.ResendVerificationEmail(ctx, email, language); err != nil {
				t.Fatalf("ResendVerificationEmail(ctx, %v, %v) error = %v", email, language, err)
			}
			if len(tokenRepo.deletedUserIDs) != 1 || tokenRepo.created == nil {
				t.Errorf(
					"ResendVerificationEmail(ctx, %v, %v) deleted tokens of %v and created %v, want the old tokens replaced",
					email, language, tokenRepo.deletedUserIDs, tokenRepo.created,
				)
			}
			if len(mailService.sent) != 1 || mailService.sent[0].To != mail.Address(email) {
				t.Errorf(
					"ResendVerificationEmail(ctx, %v, %v) sent %v, want one mail to %v",
					email, language, mailService.sent, email,
				)
			}
		},
	)

	verifiedAt := time.Now()
	tests := []struct {
		name        string
		repo        *fakeRepository
		latestToken *user.EmailVerificationToken
	}{
		{
			name: "does nothing when the user does not exist",
			repo: &fakeRepository{getByEmailErr: exception.ErrNotFound},
		},
		{
			name: "does nothing when the user is already verified",
			repo: &fakeRepository{user: &user.User{ID: userID, Email: email, EmailVerifiedAt: &verifiedAt}},
		},
		{
			name: "does nothing when the last mail was sent within the resend interval",
			repo: &fakeRepository{user: &user.User{ID: userID, Email: email}},
			latestToken: &user.EmailVerificationToken{
				UserID:    userID,
				CreatedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := &fakeEmailVerificationTokenRepo{latestToken: tt.latestToken}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				tt.repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				tokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.ResendVerificationEmail(ctx, email, language); err != nil {
				t.Fatalf("ResendVerificationEmail(ctx, %v, %v) error = %v, want nil", email, language, err)
			}
			if tokenRepo.created != nil || len(mailService.sent) != 0 {
				t.Errorf(
					"ResendVerificationEmail(ctx, %v, %v) created %v and sent %v, want neither",
					email, language, tokenRepo.created, mailService.sent,
				)
			}
		})
	}
}

func TestDeleteUnverifiedUsers(t *testing.T) {
	t.Run(
		"deletes the unverified users created before the retention period",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			before := time.Now().Add(-user.UnverifiedUserRetention)
			if err := uu.DeleteUnverifiedUsers(ctx); err != nil {
				t.Fatalf("DeleteUnverifiedUsers(ctx) error = %v", err)
			}
			after := time.Now().Add(-user.UnverifiedUserRetention)

			if repo.deletedBefore.Before(before) || repo.deletedBefore.After(after) {
				t.Errorf(
					"DeleteUnverifiedUsers(ctx) deleted users created before %v, want between %v and %v",
					repo.deletedBefore, before, after,
				)
			}
		},
	)
}

func TestUpdateUsername(t *testing.T) {
	userID := user.ID(1)
	newUsername := user.Username("NewName")

	t.Run(
		"returns the user with the new username and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, user.Password("testpassword")),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdateUsername(ctx, userID, newUsername)
			if err != nil {
				t.Fatalf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, %v",
					userID, newUsername, got, err,
				)
			}

			if got.ID != userID ||
				got.Username != newUsername ||
				got.Email != existing.Email ||
				got.HashedPassword != existing.HashedPassword ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, want Username %v and the other fields of %v",
					userID, newUsername, got, newUsername, existing,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdateUsername(ctx, userID, newUsername)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateUsername(ctx, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newUsername, got, err, exception.ErrNotFound,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateUsername(ctx, %v, %v) returned %v and updated %v, want nil for both",
					userID, newUsername, got, repo.updated,
				)
			}
		},
	)
}

func TestUpdateEmail(t *testing.T) {
	userID := user.ID(1)
	password := user.Password("testpassword")
	newEmail := user.Email("new@example.com")

	t.Run(
		"returns the user with the new email and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, password),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdateEmail(ctx, userID, newEmail, password)
			if err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, newEmail, password, got, err,
				)
			}

			if got.ID != userID ||
				got.Username != existing.Username ||
				got.Email != newEmail ||
				got.HashedPassword != existing.HashedPassword ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, want Email %v and the other fields of %v",
					userID, newEmail, password, got, newEmail, existing,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid and does not update when the password is incorrect",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, password),
				},
			}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()
			wrongPassword := user.Password("wrongpassword")

			got, err := uu.UpdateEmail(ctx, userID, newEmail, wrongPassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newEmail, wrongPassword, got, err, exception.ErrInvalid,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) returned %v and updated %v, want nil for both",
					userID, newEmail, wrongPassword, got, repo.updated,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdateEmail(ctx, userID, newEmail, password)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, newEmail, password, got, err, exception.ErrNotFound,
				)
			}
			if got != nil || repo.updated != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v) returned %v and updated %v, want nil for both",
					userID, newEmail, password, got, repo.updated,
				)
			}
		},
	)
}

func TestUpdatePassword(t *testing.T) {
	userID := user.ID(1)
	oldPassword := user.Password("oldpassword")
	newPassword := user.Password("newpassword")

	t.Run(
		"hashes the new password and keeps the other fields",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, oldPassword),
				Role:           user.RoleAdmin,
			}
			repo := &fakeRepository{user: existing}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if err != nil {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, oldPassword, newPassword, got, err,
				)
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(newPassword),
			); err != nil {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					userID, oldPassword, newPassword, got.HashedPassword, newPassword,
				)
			}
			if got.ID != userID ||
				got.Username != existing.Username ||
				got.Email != existing.Email ||
				got.Role != existing.Role {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, want the other fields of %v",
					userID, oldPassword, newPassword, got, existing,
				)
			}
		},
	)

	t.Run(
		"revokes all refresh tokens of the user",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, oldPassword),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(
				repo, refreshTokenRepo, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if err != nil {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v",
					userID, oldPassword, newPassword, got, err,
				)
			}

			if refreshTokenRepo.revokedUserID != userID {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) revoked user id = %v, want %v",
					userID, oldPassword, newPassword,
					refreshTokenRepo.revokedUserID, userID,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid and neither updates nor revokes when the old password is incorrect",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, oldPassword),
				},
			}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(
				repo, refreshTokenRepo, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()
			wrongPassword := user.Password("wrongpassword")

			got, err := uu.UpdatePassword(ctx, userID, wrongPassword, newPassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, wrongPassword, newPassword, got, err, exception.ErrInvalid,
				)
			}
			if got != nil ||
				repo.updated != nil ||
				refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) returned %v, updated %v and revoked user id %v, want no result, update, or revocation",
					userID, wrongPassword, newPassword,
					got, repo.updated, refreshTokenRepo.revokedUserID,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			repo := &fakeRepository{getErr: exception.ErrNotFound}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			uu := NewUserUsecase(
				repo, refreshTokenRepo, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			got, err := uu.UpdatePassword(ctx, userID, oldPassword, newPassword)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdatePassword(ctx, %v, %v, %v) (*user.User, error) = %v, %v, want %v",
					userID, oldPassword, newPassword, got, err, exception.ErrNotFound,
				)
			}
			if got != nil ||
				repo.updated != nil ||
				refreshTokenRepo.revokedUserID != 0 {
				t.Errorf(
					"UpdatePassword(ctx, %v, %v, %v) returned %v, updated %v and revoked user id %v, want no result, update, or revocation",
					userID, oldPassword, newPassword,
					got, repo.updated, refreshTokenRepo.revokedUserID,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	userID := user.ID(1)

	t.Run(
		"deletes all media of the user",
		func(t *testing.T) {
			mediaService := &fakeMediaService{}
			uu := NewUserUsecase(
				&fakeRepository{}, &fakeRefreshTokenRepo{}, mediaService,
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Delete(ctx, userID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v) error = %v",
					userID, err,
				)
			}
			if mediaService.deletedAllUserID != userID {
				t.Errorf(
					"Delete(ctx, %v) deletes media of user %v, want %v",
					userID, mediaService.deletedAllUserID, userID,
				)
			}
		},
	)

	t.Run(
		"succeeds even when deleting the media fails",
		func(t *testing.T) {
			mediaService := &fakeMediaService{
				err: errors.New("delete media"),
			}
			uu := NewUserUsecase(
				&fakeRepository{}, &fakeRefreshTokenRepo{}, mediaService,
				&fakeEmailVerificationTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.Delete(ctx, userID); err != nil {
				t.Fatalf(
					"Delete(ctx, %v) error = %v, want nil",
					userID, err,
				)
			}
		},
	)
}
