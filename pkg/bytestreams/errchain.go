package bytestreams

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorChain stores multiple indepndent errors.
// It supports standart error matching mechanisms a.k.a errors.Is and errors.As.
type ErrorChain []error

func (errs ErrorChain) Error() string {
	var builder = &strings.Builder{}
	for i, err := range errs {
		_, _ = fmt.Fprintf(builder, "%s", err)
		if i < len(errs)-1 {
			_, _ = builder.WriteString("; ")
		}
	}
	return builder.String()
}

// As method searches for first matching error in chain by errors.As function.
func (errs ErrorChain) As(asErr interface{}) bool {
	for _, err := range errs {
		if errors.As(err, asErr) {
			return true
		}
	}
	return false
}

// Is method searches for first matching error in chain by errors.Is function.
func (errs ErrorChain) Is(isErr error) bool {
	for _, err := range errs {
		if errors.Is(err, isErr) {
			return true
		}
	}
	return false
}
