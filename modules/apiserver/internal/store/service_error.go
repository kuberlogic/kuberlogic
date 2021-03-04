package store

// Error type for 1st level customer facing Service methods
type ServiceError struct {
	ClientMsg string
	Client    bool
	Err       error
}

func NewServiceError(clientMsg string, client bool, err error) *ServiceError {
	return &ServiceError{
		ClientMsg: clientMsg,
		Client:    client,
		Err:       err,
	}
}
