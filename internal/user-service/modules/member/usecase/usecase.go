package usecase

import "context"

// MemberUsecase abstraction
type MemberUsecase interface {
	Hello(ctx context.Context, request string) (result string)
}
