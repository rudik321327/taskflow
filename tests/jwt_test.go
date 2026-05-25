package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/taskflow/taskflow/internal/auth"
)

func TestJWT_RoundTrip(t *testing.T) {
	issuer := auth.NewIssuer("a-secret-key-of-sufficient-length", time.Hour, "taskflow-test")

	token, exp, err := issuer.Issue(42, "alice@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.True(t, exp.After(time.Now()))

	claims, err := issuer.Parse(token)
	require.NoError(t, err)
	require.EqualValues(t, 42, claims.UserID)
	require.Equal(t, "alice@example.com", claims.Email)
}

func TestJWT_InvalidToken(t *testing.T) {
	issuer := auth.NewIssuer("secret-key", time.Hour, "taskflow-test")
	_, err := issuer.Parse("garbage.token.value")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWT_WrongSecretRejected(t *testing.T) {
	issuerA := auth.NewIssuer("secret-A", time.Hour, "taskflow-test")
	issuerB := auth.NewIssuer("secret-B", time.Hour, "taskflow-test")

	tok, _, err := issuerA.Issue(1, "a@b.c")
	require.NoError(t, err)

	_, err = issuerB.Parse(tok)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}
