package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type Usecase interface {
	GetByID(ctx context.Context, userID user.ID) (*user.User, error)
	Create(
		ctx context.Context,
		username user.Username,
		email user.Email,
		password user.Password,
		language mail.Language,
	) error
	VerifyEmail(
		ctx context.Context,
		value user.EmailVerificationTokenValue,
	) error
	ResendVerificationEmail(
		ctx context.Context,
		email user.Email,
		language mail.Language,
	) error
	UpdateUsername(
		ctx context.Context,
		userID user.ID,
		username user.Username,
	) (*user.User, error)
	UpdateEmail(
		ctx context.Context,
		userID user.ID,
		email user.Email,
		password user.Password,
		language mail.Language,
	) error
	VerifyEmailChange(
		ctx context.Context,
		value user.EmailChangeTokenValue,
	) error
	RequestPasswordReset(
		ctx context.Context,
		email user.Email,
		language mail.Language,
	) error
	ResetPassword(
		ctx context.Context,
		value user.PasswordResetTokenValue,
		newPassword user.Password,
	) error
	UpdatePassword(
		ctx context.Context,
		userID user.ID,
		oldPassword user.Password,
		newPassword user.Password,
	) (*user.User, error)
	Delete(ctx context.Context, userID user.ID, language mail.Language) error
	DeleteUnverifiedUsers(ctx context.Context) error
	DeleteExpiredEmailChangeTokens(ctx context.Context) error
	DeleteExpiredPasswordResetTokens(ctx context.Context) error
}

type UserUsecase struct {
	userRepo                   user.Repository
	refreshTokenRepo           auth.RefreshTokenRepository
	mediaService               media.Service
	emailVerificationTokenRepo user.EmailVerificationTokenRepository
	emailChangeTokenRepo       user.EmailChangeTokenRepository
	passwordResetTokenRepo     user.PasswordResetTokenRepository
	mailService                mail.Service
	supportAddress             mail.Address
	adminAddress               mail.Address
	webBaseURL                 *url.URL
}

func NewUserUsecase(
	userRepo user.Repository,
	refreshTokenRepo auth.RefreshTokenRepository,
	mediaService media.Service,
	emailVerificationTokenRepo user.EmailVerificationTokenRepository,
	emailChangeTokenRepo user.EmailChangeTokenRepository,
	passwordResetTokenRepo user.PasswordResetTokenRepository,
	mailService mail.Service,
	supportAddress mail.Address,
	adminAddress mail.Address,
	webBaseURL *url.URL,
) *UserUsecase {
	return &UserUsecase{
		userRepo:                   userRepo,
		refreshTokenRepo:           refreshTokenRepo,
		mediaService:               mediaService,
		emailVerificationTokenRepo: emailVerificationTokenRepo,
		emailChangeTokenRepo:       emailChangeTokenRepo,
		passwordResetTokenRepo:     passwordResetTokenRepo,
		mailService:                mailService,
		supportAddress:             supportAddress,
		adminAddress:               adminAddress,
		webBaseURL:                 webBaseURL,
	}
}

func (uu *UserUsecase) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	u, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

func (uu *UserUsecase) Create(
	ctx context.Context,
	username user.Username,
	email user.Email,
	password user.Password,
	language mail.Language,
) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	existingUser, err := uu.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("get user by email: %w", err)
	}

	if existingUser != nil && existingUser.IsEmailVerified() {
		m := buildMessage(
			uu.supportAddress,
			mail.Address(existingUser.Email),
			language,
			logoURL(uu.webBaseURL),
			alreadyRegisteredMailContent(language, loginURL(uu.webBaseURL, language)),
		)
		if err := uu.mailService.Send(ctx, m); err != nil {
			return fmt.Errorf("send already registered mail: %w", err)
		}
		return nil
	}

	u := user.NewUser(
		username,
		email,
		user.HashedPassword(hashed),
		user.RoleUser,
	)

	if existingUser != nil {
		u.ID = existingUser.ID
		if err := uu.userRepo.ReplaceUnverified(ctx, &u); err != nil {
			return fmt.Errorf("replace unverified user: %w", err)
		}
	} else {
		err := uu.userRepo.Create(ctx, &u)
		if errors.Is(err, exception.ErrAlreadyExists) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
	}

	if err := uu.sendVerificationMail(ctx, &u, language); err != nil {
		return err
	}

	return nil
}

func (uu *UserUsecase) VerifyEmail(
	ctx context.Context,
	value user.EmailVerificationTokenValue,
) error {
	token, err := uu.emailVerificationTokenRepo.FindValidByValue(ctx, value)
	if err != nil {
		return fmt.Errorf("find email verification token: %w", err)
	}

	existingUser, err := uu.userRepo.GetByID(ctx, token.UserID)
	if errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("verify email: %w", exception.ErrInvalid)
	}
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if existingUser.IsEmailVerified() {
		if err := uu.emailVerificationTokenRepo.DeleteAllForUser(ctx, existingUser.ID); err != nil {
			return fmt.Errorf("delete email verification tokens: %w", err)
		}
		return nil
	}

	verifiedAt := time.Now()
	existingUser.EmailVerifiedAt = &verifiedAt
	if err := uu.userRepo.UpdateEmailVerifiedAt(ctx, existingUser); err != nil {
		return fmt.Errorf("update email verified at: %w", err)
	}

	if err := uu.emailVerificationTokenRepo.DeleteAllForUser(ctx, existingUser.ID); err != nil {
		return fmt.Errorf("delete email verification tokens: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		uu.adminAddress,
		mail.LanguageJA,
		logoURL(uu.webBaseURL),
		userVerifiedMailContent(verifiedAt),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		slog.WarnContext(ctx, "user_verified_mail_failed", "userId", fmt.Sprint(existingUser.ID), "message", err.Error())
	}

	return nil
}

func (uu *UserUsecase) ResendVerificationEmail(
	ctx context.Context,
	email user.Email,
	language mail.Language,
) error {
	existingUser, err := uu.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, exception.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}

	if existingUser.IsEmailVerified() {
		return nil
	}

	latestToken, err := uu.emailVerificationTokenRepo.FindLatestByUserID(ctx, existingUser.ID)
	if err != nil && !errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("find latest email verification token: %w", err)
	}
	if latestToken != nil && time.Since(latestToken.CreatedAt) < user.EmailVerificationResendInterval {
		return nil
	}

	if err := uu.sendVerificationMail(ctx, existingUser, language); err != nil {
		return err
	}

	return nil
}

func (uu *UserUsecase) sendVerificationMail(
	ctx context.Context,
	u *user.User,
	language mail.Language,
) error {
	if err := uu.emailVerificationTokenRepo.DeleteAllForUser(ctx, u.ID); err != nil {
		return fmt.Errorf("delete email verification tokens: %w", err)
	}

	token, err := uu.emailVerificationTokenRepo.Create(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("create email verification token: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(u.Email),
		language,
		logoURL(uu.webBaseURL),
		verificationMailContent(language, verificationURL(uu.webBaseURL, language, token.Value)),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		return fmt.Errorf("send verification mail: %w", err)
	}

	return nil
}

func (uu *UserUsecase) UpdateUsername(
	ctx context.Context,
	userID user.ID,
	newUsername user.Username,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	u := user.NewUser(
		newUsername,
		existingUser.Email,
		existingUser.HashedPassword,
		existingUser.Role,
	)
	u.ID = userID

	if err := uu.userRepo.UpdateUsername(ctx, &u); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) UpdateEmail(
	ctx context.Context,
	userID user.ID,
	newEmail user.Email,
	password user.Password,
	language mail.Language,
) error {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword),
		[]byte(password),
	); err != nil {
		return fmt.Errorf("login: %w", exception.ErrInvalid)
	}

	if newEmail == existingUser.Email {
		return nil
	}

	otherUser, err := uu.userRepo.GetByEmail(ctx, newEmail)
	if err != nil && !errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("get user by email: %w", err)
	}

	if otherUser != nil && otherUser.IsEmailVerified() {
		m := buildMessage(
			uu.supportAddress,
			mail.Address(newEmail),
			language,
			logoURL(uu.webBaseURL),
			emailChangeAlreadyRegisteredMailContent(language, loginURL(uu.webBaseURL, language)),
		)
		if err := uu.mailService.Send(ctx, m); err != nil {
			return fmt.Errorf("send already registered mail: %w", err)
		}
		return nil
	}

	if err := uu.emailChangeTokenRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete email change tokens: %w", err)
	}

	token, err := uu.emailChangeTokenRepo.Create(ctx, userID, newEmail, language)
	if err != nil {
		return fmt.Errorf("create email change token: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(newEmail),
		language,
		logoURL(uu.webBaseURL),
		emailChangeVerificationMailContent(
			language,
			emailChangeVerificationURL(uu.webBaseURL, language, token.Value),
		),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		return fmt.Errorf("send email change verification mail: %w", err)
	}

	return nil
}

func (uu *UserUsecase) VerifyEmailChange(
	ctx context.Context,
	value user.EmailChangeTokenValue,
) error {
	token, err := uu.emailChangeTokenRepo.FindValidByValue(ctx, value)
	if err != nil {
		return fmt.Errorf("find email change token: %w", err)
	}

	existingUser, err := uu.userRepo.GetByID(ctx, token.UserID)
	if errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("verify email change: %w", exception.ErrInvalid)
	}
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	otherUser, err := uu.userRepo.GetByEmail(ctx, token.NewEmail)
	if err != nil && !errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("get user by email: %w", err)
	}
	if otherUser != nil && otherUser.ID == existingUser.ID {
		if err := uu.emailChangeTokenRepo.DeleteAllForUser(ctx, existingUser.ID); err != nil {
			return fmt.Errorf("delete email change tokens: %w", err)
		}
		return nil
	}

	if otherUser != nil {
		if otherUser.IsEmailVerified() {
			return fmt.Errorf("verify email change: %w", exception.ErrAlreadyExists)
		}
		if err := uu.userRepo.Delete(ctx, otherUser.ID); err != nil {
			return fmt.Errorf("delete unverified user: %w", err)
		}
	}

	oldEmail := existingUser.Email
	existingUser.Email = token.NewEmail
	if err := uu.userRepo.UpdateEmail(ctx, existingUser); err != nil {
		return fmt.Errorf("update email: %w", err)
	}

	if err := uu.emailChangeTokenRepo.DeleteAllForUser(ctx, existingUser.ID); err != nil {
		return fmt.Errorf("delete email change tokens: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(oldEmail),
		token.Language,
		logoURL(uu.webBaseURL),
		emailChangedMailContent(token.Language, token.NewEmail, contactURL(uu.webBaseURL, token.Language)),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		slog.WarnContext(ctx, "email_changed_mail_failed", "userId", fmt.Sprint(existingUser.ID), "message", err.Error())
	}

	return nil
}

func (uu *UserUsecase) RequestPasswordReset(
	ctx context.Context,
	email user.Email,
	language mail.Language,
) error {
	existingUser, err := uu.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, exception.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}

	if !existingUser.IsEmailVerified() {
		return nil
	}

	latestToken, err := uu.passwordResetTokenRepo.FindLatestByUserID(ctx, existingUser.ID)
	if err != nil && !errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("find latest password reset token: %w", err)
	}
	if latestToken != nil && time.Since(latestToken.CreatedAt) < user.PasswordResetRequestInterval {
		return nil
	}

	if err := uu.passwordResetTokenRepo.DeleteAllForUser(ctx, existingUser.ID); err != nil {
		return fmt.Errorf("delete password reset tokens: %w", err)
	}

	token, err := uu.passwordResetTokenRepo.Create(ctx, existingUser.ID, language)
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(existingUser.Email),
		language,
		logoURL(uu.webBaseURL),
		passwordResetMailContent(language, passwordResetURL(uu.webBaseURL, language, token.Value)),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		return fmt.Errorf("send password reset mail: %w", err)
	}

	return nil
}

func (uu *UserUsecase) ResetPassword(
	ctx context.Context,
	value user.PasswordResetTokenValue,
	newPassword user.Password,
) error {
	token, err := uu.passwordResetTokenRepo.FindValidByValue(ctx, value)
	if err != nil {
		return fmt.Errorf("find password reset token: %w", err)
	}

	existingUser, err := uu.userRepo.GetByID(ctx, token.UserID)
	if errors.Is(err, exception.ErrNotFound) {
		return fmt.Errorf("reset password: %w", exception.ErrInvalid)
	}
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	u := user.NewUser(
		existingUser.Username,
		existingUser.Email,
		user.HashedPassword(hashed),
		existingUser.Role,
	)
	u.ID = existingUser.ID

	if err := uu.userRepo.UpdatePassword(ctx, &u); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := uu.passwordResetTokenRepo.DeleteAllForUser(ctx, u.ID); err != nil {
		return fmt.Errorf("delete password reset tokens: %w", err)
	}

	if err := uu.refreshTokenRepo.RevokeAllForUser(ctx, u.ID); err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(u.Email),
		token.Language,
		logoURL(uu.webBaseURL),
		passwordChangedMailContent(token.Language, contactURL(uu.webBaseURL, token.Language)),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		slog.WarnContext(ctx, "password_changed_mail_failed", "userId", fmt.Sprint(u.ID), "message", err.Error())
	}

	return nil
}

func (uu *UserUsecase) DeleteExpiredPasswordResetTokens(ctx context.Context) error {
	if err := uu.passwordResetTokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("delete expired password reset tokens: %w", err)
	}

	return nil
}

func (uu *UserUsecase) DeleteExpiredEmailChangeTokens(ctx context.Context) error {
	if err := uu.emailChangeTokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("delete expired email change tokens: %w", err)
	}

	return nil
}

func (uu *UserUsecase) UpdatePassword(
	ctx context.Context,
	userID user.ID,
	oldPassword user.Password,
	newPassword user.Password,
) (*user.User, error) {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword),
		[]byte(oldPassword),
	); err != nil {
		return nil, fmt.Errorf("login: %w", exception.ErrInvalid)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.NewUser(
		existingUser.Username,
		existingUser.Email,
		user.HashedPassword(hashed),
		existingUser.Role,
	)
	u.ID = userID

	if err := uu.userRepo.UpdatePassword(ctx, &u); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	if err := uu.refreshTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("revoke refresh tokens: %w", err)
	}

	return &u, nil
}

func (uu *UserUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	language mail.Language,
) error {
	existingUser, err := uu.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if err := uu.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := uu.mediaService.DeleteAllForUser(ctx, userID); err != nil {
		slog.WarnContext(ctx, "media_delete_failed", "userId", fmt.Sprint(userID), "message", err.Error())
	}

	m := buildMessage(
		uu.supportAddress,
		mail.Address(existingUser.Email),
		language,
		logoURL(uu.webBaseURL),
		withdrawnMailContent(language),
	)
	if err := uu.mailService.Send(ctx, m); err != nil {
		slog.WarnContext(ctx, "withdrawn_mail_failed", "userId", fmt.Sprint(userID), "message", err.Error())
	}

	return nil
}

func (uu *UserUsecase) DeleteUnverifiedUsers(ctx context.Context) error {
	before := time.Now().Add(-user.UnverifiedUserRetention)
	if err := uu.userRepo.DeleteUnverifiedCreatedBefore(ctx, before); err != nil {
		return fmt.Errorf("delete unverified users: %w", err)
	}

	return nil
}
