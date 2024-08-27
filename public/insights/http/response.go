package http

import (
	"fmt"
)

type Response struct {
	Code int
	Data []byte
}

func (r Response) String() string {
	return fmt.Sprintf("%d: %s", r.Code, string(r.Data))
}
