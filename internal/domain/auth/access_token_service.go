package auth

type AccessTokenService interface {
	Generate(principal *Principal) (*AccessToken, error)
	Validate(accessToken *AccessToken) (*Principal, error)
}
