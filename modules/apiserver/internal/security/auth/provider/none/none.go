package none

const (
	noneEmail  = "none@example.com"
	noneSecret = "secret"
)

type noneAuthProvider struct{}

func (n *noneAuthProvider) GetAuthenticationSecret(username, password string) (string, error) {
	return noneSecret, nil
}

func (n *noneAuthProvider) Authenticate(secret string) (string, string, error) {
	return noneEmail, noneSecret, nil
}

func (n *noneAuthProvider) Authorize(username, action, object string) (bool, error) {
	return true, nil
}

func (n *noneAuthProvider) CreatePermissionResource(obj string) error {
	return nil
}

func (n *noneAuthProvider) DeletePermissionResource(obj string) error {
	return nil
}

func NewNoneProvider() (*noneAuthProvider, error) {
	return &noneAuthProvider{}, nil
}
