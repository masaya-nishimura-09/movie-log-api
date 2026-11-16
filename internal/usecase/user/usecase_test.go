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
	userByEmail   *user.User
	deletedUserID user.ID
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
	if r.getByEmailErr != nil {
		return nil, r.getByEmailErr
	}
	if r.userByEmail != nil {
		return r.userByEmail, nil
	}
	return r.user, nil
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

type fakeEmailChangeTokenRepo struct {
	token          *user.EmailChangeToken
	findErr        error
	created        *user.EmailChangeToken
	deletedUserIDs []user.ID
	deletedExpired bool
}

func (r *fakeEmailChangeTokenRepo) Create(
	ctx context.Context,
	userID user.ID,
	newEmail user.Email,
	language mail.Language,
) (*user.EmailChangeToken, error) {
	r.created = &user.EmailChangeToken{
		UserID:   userID,
		NewEmail: newEmail,
		Language: language,
		Value:    user.EmailChangeTokenValue("testchangetokenvalue"),
	}
	return r.created, nil
}

func (r *fakeEmailChangeTokenRepo) FindValidByValue(
	ctx context.Context,
	value user.EmailChangeTokenValue,
) (*user.EmailChangeToken, error) {
	return r.token, r.findErr
}

func (r *fakeEmailChangeTokenRepo) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	r.deletedUserIDs = append(r.deletedUserIDs, userID)
	return nil
}

func (r *fakeEmailChangeTokenRepo) DeleteExpired(
	ctx context.Context,
) error {
	r.deletedExpired = true
	return nil
}

type fakePasswordResetTokenRepo struct {
	token          *user.PasswordResetToken
	findErr        error
	latestToken    *user.PasswordResetToken
	created        *user.PasswordResetToken
	deletedUserIDs []user.ID
	deletedExpired bool
}

func (r *fakePasswordResetTokenRepo) Create(
	ctx context.Context,
	userID user.ID,
	language mail.Language,
) (*user.PasswordResetToken, error) {
	r.created = &user.PasswordResetToken{
		UserID:   userID,
		Language: language,
		Value:    user.PasswordResetTokenValue("testresettokenvalue"),
	}
	return r.created, nil
}

func (r *fakePasswordResetTokenRepo) FindValidByValue(
	ctx context.Context,
	value user.PasswordResetTokenValue,
) (*user.PasswordResetToken, error) {
	return r.token, r.findErr
}

func (r *fakePasswordResetTokenRepo) FindLatestByUserID(
	ctx context.Context,
	userID user.ID,
) (*user.PasswordResetToken, error) {
	if r.latestToken == nil {
		return nil, exception.ErrNotFound
	}
	return r.latestToken, nil
}

func (r *fakePasswordResetTokenRepo) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	r.deletedUserIDs = append(r.deletedUserIDs, userID)
	return nil
}

func (r *fakePasswordResetTokenRepo) DeleteExpired(
	ctx context.Context,
) error {
	r.deletedExpired = true
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
	r.deletedUserID = userID
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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
				tokenRepo, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, mailService,
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

func TestRequestPasswordReset(t *testing.T) {
	email := user.Email("test@example.com")
	language := mail.LanguageJA
	userID := user.ID(1)
	verifiedAt := time.Now()

	t.Run(
		"sends a password reset mail when the user is verified",
		func(t *testing.T) {
			repo := &fakeRepository{user: &user.User{ID: userID, Email: email, EmailVerifiedAt: &verifiedAt}}
			resetTokenRepo := &fakePasswordResetTokenRepo{
				latestToken: &user.PasswordResetToken{
					UserID:    userID,
					CreatedAt: time.Now().Add(-user.PasswordResetRequestInterval - time.Second),
				},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.RequestPasswordReset(ctx, email, language); err != nil {
				t.Fatalf("RequestPasswordReset(ctx, %v, %v) error = %v", email, language, err)
			}

			if resetTokenRepo.created == nil ||
				resetTokenRepo.created.UserID != userID ||
				resetTokenRepo.created.Language != language {
				t.Fatalf(
					"RequestPasswordReset(ctx, %v, %v) created %v, want a token for user %v with %v",
					email, language, resetTokenRepo.created, userID, language,
				)
			}
			if len(resetTokenRepo.deletedUserIDs) != 1 || resetTokenRepo.deletedUserIDs[0] != userID {
				t.Errorf(
					"RequestPasswordReset(ctx, %v, %v) deleted tokens of %v, want %v",
					email, language, resetTokenRepo.deletedUserIDs, userID,
				)
			}

			if len(mailService.sent) != 1 {
				t.Fatalf(
					"RequestPasswordReset(ctx, %v, %v) sent %d mails, want 1",
					email, language, len(mailService.sent),
				)
			}
			m := mailService.sent[0]
			wantURL := "https://web.example.com/ja/reset-password?token=testresettokenvalue"
			if m.To != mail.Address(email) ||
				m.Subject != mail.Subject("【Cinelier】パスワードの再設定") ||
				!strings.Contains(string(m.TextBody), wantURL) {
				t.Errorf(
					"RequestPasswordReset(ctx, %v, %v) sent %v, want a reset mail to %v with %v",
					email, language, m, email, wantURL,
				)
			}
		},
	)

	tests := []struct {
		name        string
		repo        *fakeRepository
		latestToken *user.PasswordResetToken
	}{
		{
			name: "does nothing when the user does not exist",
			repo: &fakeRepository{getByEmailErr: exception.ErrNotFound},
		},
		{
			name: "does nothing when the user is not verified",
			repo: &fakeRepository{user: &user.User{ID: userID, Email: email}},
		},
		{
			name: "does nothing when the last mail was sent within the request interval",
			repo: &fakeRepository{user: &user.User{ID: userID, Email: email, EmailVerifiedAt: &verifiedAt}},
			latestToken: &user.PasswordResetToken{
				UserID:    userID,
				CreatedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTokenRepo := &fakePasswordResetTokenRepo{latestToken: tt.latestToken}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				tt.repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.RequestPasswordReset(ctx, email, language); err != nil {
				t.Fatalf("RequestPasswordReset(ctx, %v, %v) error = %v, want nil", email, language, err)
			}
			if resetTokenRepo.created != nil || len(mailService.sent) != 0 {
				t.Errorf(
					"RequestPasswordReset(ctx, %v, %v) created %v and sent %v, want neither",
					email, language, resetTokenRepo.created, mailService.sent,
				)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	value := user.PasswordResetTokenValue("testresettokenvalue")
	newPassword := user.Password("newpassword")
	userID := user.ID(1)
	email := user.Email("test@example.com")

	t.Run(
		"updates the password, deletes the tokens, revokes the refresh tokens and notifies the user",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Username:       user.Username("Test"),
				Email:          email,
				HashedPassword: hashPassword(t, user.Password("oldpassword")),
				Role:           user.RoleUser,
			}
			repo := &fakeRepository{user: existing}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			resetTokenRepo := &fakePasswordResetTokenRepo{
				token: &user.PasswordResetToken{UserID: userID, Language: mail.LanguageEN},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, refreshTokenRepo, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.ResetPassword(ctx, value, newPassword); err != nil {
				t.Fatalf("ResetPassword(ctx, %v, %v) error = %v", value, newPassword, err)
			}

			got := repo.updated
			if got == nil || got.ID != userID {
				t.Fatalf("ResetPassword(ctx, %v, %v) updated %v, want user %v", value, newPassword, got, userID)
			}
			if err := bcrypt.CompareHashAndPassword(
				[]byte(got.HashedPassword),
				[]byte(newPassword),
			); err != nil {
				t.Errorf(
					"ResetPassword(ctx, %v, %v) HashedPassword = %v, want a bcrypt hash of %v",
					value, newPassword, got.HashedPassword, newPassword,
				)
			}
			if len(resetTokenRepo.deletedUserIDs) != 1 || resetTokenRepo.deletedUserIDs[0] != userID {
				t.Errorf(
					"ResetPassword(ctx, %v, %v) deleted tokens of %v, want %v",
					value, newPassword, resetTokenRepo.deletedUserIDs, userID,
				)
			}
			if refreshTokenRepo.revokedUserID != userID {
				t.Errorf(
					"ResetPassword(ctx, %v, %v) revoked refresh tokens of %v, want %v",
					value, newPassword, refreshTokenRepo.revokedUserID, userID,
				)
			}
			if len(mailService.sent) != 1 ||
				mailService.sent[0].To != mail.Address(email) ||
				mailService.sent[0].Subject != mail.Subject("[Cinelier] Your password has been changed") {
				t.Errorf(
					"ResetPassword(ctx, %v, %v) sent %v, want one notice in English to %v",
					value, newPassword, mailService.sent, email,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is invalid or expired",
		func(t *testing.T) {
			repo := &fakeRepository{}
			resetTokenRepo := &fakePasswordResetTokenRepo{findErr: exception.ErrInvalid}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			err := uu.ResetPassword(ctx, value, newPassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf("ResetPassword(ctx, %v, %v) error = %v, want %v", value, newPassword, err, exception.ErrInvalid)
			}
			if repo.updated != nil {
				t.Errorf("ResetPassword(ctx, %v, %v) updated %v, want no update", value, newPassword, repo.updated)
			}
		},
	)

	t.Run(
		"succeeds even when notifying the user fails",
		func(t *testing.T) {
			repo := &fakeRepository{user: &user.User{ID: userID, Email: email}}
			resetTokenRepo := &fakePasswordResetTokenRepo{
				token: &user.PasswordResetToken{UserID: userID, Language: mail.LanguageJA},
			}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo,
				&fakeMailService{err: errors.New("send mail")},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.ResetPassword(ctx, value, newPassword); err != nil {
				t.Fatalf("ResetPassword(ctx, %v, %v) error = %v, want nil", value, newPassword, err)
			}
			if repo.updated == nil {
				t.Errorf("ResetPassword(ctx, %v, %v) did not update the password", value, newPassword)
			}
		},
	)
}

func TestDeleteExpiredPasswordResetTokens(t *testing.T) {
	t.Run(
		"deletes the expired password reset tokens",
		func(t *testing.T) {
			resetTokenRepo := &fakePasswordResetTokenRepo{}
			uu := NewUserUsecase(
				&fakeRepository{}, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, resetTokenRepo, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.DeleteExpiredPasswordResetTokens(ctx); err != nil {
				t.Fatalf("DeleteExpiredPasswordResetTokens(ctx) error = %v", err)
			}
			if !resetTokenRepo.deletedExpired {
				t.Errorf("DeleteExpiredPasswordResetTokens(ctx) did not delete the expired tokens")
			}
		},
	)
}

func TestDeleteUnverifiedUsers(t *testing.T) {
	t.Run(
		"deletes the unverified users created before the retention period",
		func(t *testing.T) {
			repo := &fakeRepository{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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

	language := mail.LanguageJA

	t.Run(
		"sends a verification mail to the new email without updating the email",
		func(t *testing.T) {
			existing := &user.User{
				ID:             userID,
				Email:          user.Email("test@example.com"),
				HashedPassword: hashPassword(t, password),
			}
			repo := &fakeRepository{user: existing, getByEmailErr: exception.ErrNotFound}
			changeTokenRepo := &fakeEmailChangeTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.UpdateEmail(ctx, userID, newEmail, password, language); err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) error = %v",
					userID, newEmail, password, language, err,
				)
			}

			if repo.updated != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) updated %v, want no update",
					userID, newEmail, password, language, repo.updated,
				)
			}
			got := changeTokenRepo.created
			if got == nil || got.UserID != userID || got.NewEmail != newEmail || got.Language != language {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) created token %v, want a token for user %v with %v and %v",
					userID, newEmail, password, language, got, userID, newEmail, language,
				)
			}
			if len(changeTokenRepo.deletedUserIDs) != 1 || changeTokenRepo.deletedUserIDs[0] != userID {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) deleted tokens of %v, want %v",
					userID, newEmail, password, language, changeTokenRepo.deletedUserIDs, userID,
				)
			}

			if len(mailService.sent) != 1 {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) sent %d mails, want 1",
					userID, newEmail, password, language, len(mailService.sent),
				)
			}
			m := mailService.sent[0]
			wantURL := "https://web.example.com/ja/verify-email-change?token=testchangetokenvalue"
			if m.To != mail.Address(newEmail) ||
				m.Subject != mail.Subject("【Cinelier】新しいメールアドレスの確認") ||
				!strings.Contains(string(m.TextBody), wantURL) {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) sent %v, want a verification mail to %v with %v",
					userID, newEmail, password, language, m, newEmail, wantURL,
				)
			}
		},
	)

	t.Run(
		"sends an already registered mail without creating a token when the new email is verified by another user",
		func(t *testing.T) {
			verifiedAt := time.Now()
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					Email:          user.Email("test@example.com"),
					HashedPassword: hashPassword(t, password),
				},
				userByEmail: &user.User{ID: user.ID(2), Email: newEmail, EmailVerifiedAt: &verifiedAt},
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.UpdateEmail(ctx, userID, newEmail, password, language); err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) error = %v",
					userID, newEmail, password, language, err,
				)
			}
			if changeTokenRepo.created != nil {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) created token %v, want none",
					userID, newEmail, password, language, changeTokenRepo.created,
				)
			}
			if len(mailService.sent) != 1 ||
				mailService.sent[0].To != mail.Address(newEmail) ||
				mailService.sent[0].Subject != mail.Subject("【Cinelier】このメールアドレスは登録済みです") {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) sent %v, want one already registered mail to %v",
					userID, newEmail, password, language, mailService.sent, newEmail,
				)
			}
		},
	)

	t.Run(
		"sends a verification mail when the new email is used by an unverified user",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					Email:          user.Email("test@example.com"),
					HashedPassword: hashPassword(t, password),
				},
				userByEmail: &user.User{ID: user.ID(2), Email: newEmail},
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.UpdateEmail(ctx, userID, newEmail, password, language); err != nil {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) error = %v",
					userID, newEmail, password, language, err,
				)
			}
			if changeTokenRepo.created == nil ||
				len(mailService.sent) != 1 ||
				mailService.sent[0].Subject != mail.Subject("【Cinelier】新しいメールアドレスの確認") {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) created %v and sent %v, want a token and one verification mail",
					userID, newEmail, password, language, changeTokenRepo.created, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid and sends no mail when the password is incorrect",
		func(t *testing.T) {
			repo := &fakeRepository{
				user: &user.User{
					ID:             userID,
					HashedPassword: hashPassword(t, password),
				},
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()
			wrongPassword := user.Password("wrongpassword")

			err := uu.UpdateEmail(ctx, userID, newEmail, wrongPassword, language)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) error = %v, want %v",
					userID, newEmail, wrongPassword, language, err, exception.ErrInvalid,
				)
			}
			if changeTokenRepo.created != nil || len(mailService.sent) != 0 {
				t.Errorf(
					"UpdateEmail(ctx, %v, %v, %v, %v) created %v and sent %v, want neither",
					userID, newEmail, wrongPassword, language, changeTokenRepo.created, mailService.sent,
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			err := uu.UpdateEmail(ctx, userID, newEmail, password, language)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"UpdateEmail(ctx, %v, %v, %v, %v) error = %v, want %v",
					userID, newEmail, password, language, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestVerifyEmailChange(t *testing.T) {
	value := user.EmailChangeTokenValue("testchangetokenvalue")
	userID := user.ID(1)
	oldEmail := user.Email("old@example.com")
	newEmail := user.Email("new@example.com")

	t.Run(
		"changes the email, deletes the tokens and notifies the old email",
		func(t *testing.T) {
			repo := &fakeRepository{
				user:          &user.User{ID: userID, Email: oldEmail},
				getByEmailErr: exception.ErrNotFound,
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{
				token: &user.EmailChangeToken{UserID: userID, NewEmail: newEmail, Language: mail.LanguageEN},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmailChange(ctx, value); err != nil {
				t.Fatalf("VerifyEmailChange(ctx, %v) error = %v", value, err)
			}

			if repo.updated == nil || repo.updated.ID != userID || repo.updated.Email != newEmail {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) updated %v, want user %v with email %v",
					value, repo.updated, userID, newEmail,
				)
			}
			if len(changeTokenRepo.deletedUserIDs) != 1 || changeTokenRepo.deletedUserIDs[0] != userID {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) deleted tokens of %v, want %v",
					value, changeTokenRepo.deletedUserIDs, userID,
				)
			}
			if len(mailService.sent) != 1 ||
				mailService.sent[0].To != mail.Address(oldEmail) ||
				mailService.sent[0].Subject != mail.Subject("[Cinelier] Your email address has been changed") {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) sent %v, want one notice in English to %v",
					value, mailService.sent, oldEmail,
				)
			}
		},
	)

	t.Run(
		"deletes the unverified user using the new email before changing the email",
		func(t *testing.T) {
			otherID := user.ID(2)
			repo := &fakeRepository{
				user:        &user.User{ID: userID, Email: oldEmail},
				userByEmail: &user.User{ID: otherID, Email: newEmail},
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{
				token: &user.EmailChangeToken{UserID: userID, NewEmail: newEmail, Language: mail.LanguageJA},
			}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmailChange(ctx, value); err != nil {
				t.Fatalf("VerifyEmailChange(ctx, %v) error = %v", value, err)
			}
			if repo.deletedUserID != otherID {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) deleted user %v, want %v",
					value, repo.deletedUserID, otherID,
				)
			}
			if repo.updated == nil || repo.updated.Email != newEmail {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) updated %v, want email %v",
					value, repo.updated, newEmail,
				)
			}
		},
	)

	t.Run(
		"returns ErrAlreadyExists without changing the email when the new email is verified by another user",
		func(t *testing.T) {
			verifiedAt := time.Now()
			repo := &fakeRepository{
				user:        &user.User{ID: userID, Email: oldEmail},
				userByEmail: &user.User{ID: user.ID(2), Email: newEmail, EmailVerifiedAt: &verifiedAt},
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{
				token: &user.EmailChangeToken{UserID: userID, NewEmail: newEmail, Language: mail.LanguageJA},
			}
			mailService := &fakeMailService{}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, mailService,
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			err := uu.VerifyEmailChange(ctx, value)
			if !errors.Is(err, exception.ErrAlreadyExists) {
				t.Fatalf("VerifyEmailChange(ctx, %v) error = %v, want %v", value, err, exception.ErrAlreadyExists)
			}
			if repo.updated != nil || repo.deletedUserID != 0 || len(mailService.sent) != 0 {
				t.Errorf(
					"VerifyEmailChange(ctx, %v) updated %v, deleted %v and sent %v, want none",
					value, repo.updated, repo.deletedUserID, mailService.sent,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is invalid or expired",
		func(t *testing.T) {
			repo := &fakeRepository{}
			changeTokenRepo := &fakeEmailChangeTokenRepo{findErr: exception.ErrInvalid}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			err := uu.VerifyEmailChange(ctx, value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf("VerifyEmailChange(ctx, %v) error = %v, want %v", value, err, exception.ErrInvalid)
			}
			if repo.updated != nil {
				t.Errorf("VerifyEmailChange(ctx, %v) updated %v, want no update", value, repo.updated)
			}
		},
	)

	t.Run(
		"succeeds even when notifying the old email fails",
		func(t *testing.T) {
			repo := &fakeRepository{
				user:          &user.User{ID: userID, Email: oldEmail},
				getByEmailErr: exception.ErrNotFound,
			}
			changeTokenRepo := &fakeEmailChangeTokenRepo{
				token: &user.EmailChangeToken{UserID: userID, NewEmail: newEmail, Language: mail.LanguageJA},
			}
			uu := NewUserUsecase(
				repo, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, &fakeMailService{err: errors.New("send mail")},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.VerifyEmailChange(ctx, value); err != nil {
				t.Fatalf("VerifyEmailChange(ctx, %v) error = %v, want nil", value, err)
			}
			if repo.updated == nil || repo.updated.Email != newEmail {
				t.Errorf("VerifyEmailChange(ctx, %v) updated %v, want email %v", value, repo.updated, newEmail)
			}
		},
	)
}

func TestDeleteExpiredEmailChangeTokens(t *testing.T) {
	t.Run(
		"deletes the expired email change tokens",
		func(t *testing.T) {
			changeTokenRepo := &fakeEmailChangeTokenRepo{}
			uu := NewUserUsecase(
				&fakeRepository{}, &fakeRefreshTokenRepo{}, &fakeMediaService{},
				&fakeEmailVerificationTokenRepo{}, changeTokenRepo, &fakePasswordResetTokenRepo{}, &fakeMailService{},
				testSupportAddress, testAdminAddress, testWebBaseURL,
			)

			ctx := context.Background()

			if err := uu.DeleteExpiredEmailChangeTokens(ctx); err != nil {
				t.Fatalf("DeleteExpiredEmailChangeTokens(ctx) error = %v", err)
			}
			if !changeTokenRepo.deletedExpired {
				t.Errorf("DeleteExpiredEmailChangeTokens(ctx) did not delete the expired tokens")
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
				&fakeEmailVerificationTokenRepo{}, &fakeEmailChangeTokenRepo{}, &fakePasswordResetTokenRepo{}, &fakeMailService{},
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
