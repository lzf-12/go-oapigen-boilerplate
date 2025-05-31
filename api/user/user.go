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

	result := make(map[int64]User, 1000)

	var email *string = (*string)(&request.Body.Email)
	var role *string = (*string)(&request.Body.Role)

	user := User{
		Username: &request.Body.Username,
		Email:    email,
		Role:     role,
		Age:      request.Body.Age,
	}

	id := u.NextId
	user.Id = &id

	// placeholder implementation, should be changed to db
	result[id] = user

	u.Users = result
	u.NextId = id + 1

	return CreateUser201JSONResponse(user), nil
}

func (u *UserImpl) GetApiV1Users(ctx context.Context, request GetApiV1UsersRequestObject) (GetApiV1UsersResponseObject, error) {

	var result []User

	// implement get users from db or cache here

	// placeholder implementation
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

	return GetApiV1Users200JSONResponse(result), nil
}
