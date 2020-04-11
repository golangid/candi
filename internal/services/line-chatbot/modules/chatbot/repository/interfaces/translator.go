package interfaces

import (
	"context"
)

// Translator abstraction
type Translator interface {
	Translate(ctx context.Context, from, to, text string) (result string)
}
