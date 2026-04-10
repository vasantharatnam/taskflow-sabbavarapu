package auth

import "context"

type contextKey string

const UserContextKey contextKey = "authenticated_user"

type AuthenticatedUser struct {
	UserID string
	Email  string
}

func WithUser(ctx context.Context, user AuthenticatedUser) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func UserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(UserContextKey).(AuthenticatedUser)
	return user, ok
}