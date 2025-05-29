//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../../docs/user.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../../docs/user.yaml

package user

import (
	"context"
	"sync"
)

type UserImpl struct {
	Users  map[int64]User
	NextId int64
	Lock   sync.Mutex
}

var _ StrictServerInterface = (*UserImpl)(nil)

func NewUserHandler() *UserImpl {
	return &UserImpl{
		Users:  make(map[int64]User),
		NextId: 1000,
	}
}

func (u *UserImpl) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {

	u.Lock.Lock()
	defer u.Lock.Unlock()

	var user User
	user.Username = &request.Body.Username
	user.Email = (*string)(&request.Body.Email)
	user.Role = (*string)(&request.Body.Role)
	user.Age = request.Body.Age

	// insert into in memory map -> can be changed to call insert db
	u.Users[*user.Id] = user

	return CreateUser201JSONResponse(user), nil
}

func (u *UserImpl) GetUsers(ctx context.Context, request GetUsersRequestObject) (GetUsersResponseObject, error) {

	var result []User

	// implement get users from db or cache here

	// placeholder boilerplate
	username := "username"
	email := "email"
	role := Member
	age := 18

	result = append(result, User{
		Username: &username,
		Email:    &email,
		Role:     (*string)(&role),
		Age:      &age,
	})

	return GetUsers200JSONResponse(result), nil
}
