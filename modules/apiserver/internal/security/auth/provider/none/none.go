package none

const (
	noneUsername = "username"
	noneSecret   = "secret"
)

type noneAuthProvider struct{}

func (n *noneAuthProvider) GetAuthenticationSecret(username, password string) (string, error) {
	return noneSecret, nil
}

func (n *noneAuthProvider) Authenticate(secret string) (string, string, error) {
	return noneUsername, noneSecret, nil
}

func (n *noneAuthProvider) Authorize(username, action, object string) (bool, error) {
	return true, nil
}

func NewNoneProvider() (*noneAuthProvider, error) {
	return &noneAuthProvider{}, nil
}
