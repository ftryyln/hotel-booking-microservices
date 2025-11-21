package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestRegisterAndLogin(t *testing.T) {
	repo := &userRepoStub{users: map[uuid.UUID]domain.User{}}
	issuer := &issuerStub{}
	svc := auth.NewService(repo, issuer)

	// register
	resp, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "User@Example.com",
		Password: "secret",
		Role:     "admin",
	})
	require.NoError(t, err)
	require.Equal(t, "user@example.com", repo.lastCreated.Email)
	require.Equal(t, "admin", repo.lastCreated.Role)
	require.NotEmpty(t, resp.AccessToken)

	// duplicate email
	_, err = svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret",
	})
	require.Error(t, err)

	// login success
	login, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: "secret",
	})
	require.NoError(t, err)
	require.Equal(t, resp.ID, login.ID)

	// login wrong password
	_, err = svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong",
	})
	require.Error(t, err)
}

func TestMeListGet(t *testing.T) {
	repo := &userRepoStub{users: map[uuid.UUID]domain.User{}}
	issuer := &issuerStub{}
	svc := auth.NewService(repo, issuer)
	user := domain.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		Password:  "hash",
		Role:      "customer",
		CreatedAt: time.Now(),
	}
	repo.users[user.ID] = user

	found, err := svc.Me(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)

	list, err := svc.List(context.Background(), query.Options{Limit: 10, Offset: 0})
	require.NoError(t, err)
	require.Len(t, list, 1)

	got, err := svc.Get(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, user.Email, got.Email)

	_, err = svc.Me(context.Background(), uuid.New())
	require.Error(t, err)
}

// stubs

type userRepoStub struct {
	users       map[uuid.UUID]domain.User
	lastCreated domain.User
}

func (u *userRepoStub) Create(ctx context.Context, user domain.User) error {
	if _, exists := u.findByEmail(user.Email); exists {
		return errors.New("duplicate")
	}
	u.users[user.ID] = user
	u.lastCreated = user
	return nil
}

func (u *userRepoStub) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	if usr, ok := u.findByEmail(email); ok {
		return usr, nil
	}
	return domain.User{}, errors.New("not found")
}

func (u *userRepoStub) findByEmail(email string) (domain.User, bool) {
	for _, v := range u.users {
		if v.Email == email {
			return v, true
		}
	}
	return domain.User{}, false
}

func (u *userRepoStub) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if usr, ok := u.users[id]; ok {
		return usr, nil
	}
	return domain.User{}, errors.New("not found")
}

func (u *userRepoStub) List(ctx context.Context, _ query.Options) ([]domain.User, error) {
	var out []domain.User
	for _, v := range u.users {
		out = append(out, v)
	}
	return out, nil
}

type issuerStub struct{}

func (i *issuerStub) Generate(ctx context.Context, user domain.User) (access, refresh string, err error) {
	return "access", "refresh", nil
}
