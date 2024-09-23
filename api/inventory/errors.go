package inventory

import (
	"errors"
	"fmt"
)

var (
	ErrNoHost    = errors.New("host does not exist")
	ErrManyHosts = errors.New("multiple hosts exist")
)

func getHumanErrorOnNon200(value int) string {
	switch value {
	case 401:
		return fmt.Sprintf("Host inventory rejected unauthorized request (status code %d).", value)
	case 403:
		return fmt.Sprintf("Host inventory rejected forbidden request (status code %d).", value)
	default:
		return fmt.Sprintf("Host inventory rejected the request (status code %d).", value)
	}
}
