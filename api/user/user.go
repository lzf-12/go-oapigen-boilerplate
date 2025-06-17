package user

import (
	"context"
	"oapi-to-rest/pkg/db"
	"oapi-to-rest/pkg/errlib"
	"strconv"
)

type UserImpl struct {
	Db *db.SQLite
	Qb *db.QueryBuilder
}

var _ StrictServerInterface = (*UserImpl)(nil)

const (
	defaultPage     int64 = 1
	defaultPageSize int64 = 20
	maxPageSize     int64 = 100
)

func (u *UserImpl) GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error) {

	var result PaginatedUserResponse
	params := request.Params

	query := u.Qb.Select(
		"u.id, u.email, u.first_name, u.last_name, u.is_active, COUNT(*) OVER() AS total_items",
	).Table("users u")

	if params.Email != nil {
		query = query.Where("u.email", "=", params.Email)
	}

	if params.IsActive != nil {
		isactive, err := strconv.ParseBool(*params.IsActive)
		if err != nil {
			return GetUser500JSONResponse{}, err
		}

		query = query.Where("u.is_active", "=", isactive)
	}

	limit, offset := GetPagination(params.Page, params.PageSize)
	query.Limit(limit).Offset(offset)

	rows, err := query.Get()
	if err != nil {
		return GetUser500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBQuery)
	}

	var totalMatched int64
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.Id,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.IsActive,
			&totalMatched,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	result.Data = &users
	result.Filters = toInterfacePtr(params)

	result.Pagination = buildPagination(params.Page, params.PageSize, totalMatched)

	return GetUser200JSONResponse(result), nil
}

func GetPagination(pagePtr, pageSizePtr *int64) (limit, offset int64) {
	page := defaultPage
	pageSize := defaultPageSize

	if pagePtr != nil && *pagePtr > 0 {
		page = *pagePtr
	}
	if pageSizePtr != nil {
		if *pageSizePtr > 0 && *pageSizePtr <= maxPageSize {
			pageSize = *pageSizePtr
		} else if *pageSizePtr > maxPageSize {
			pageSize = maxPageSize
		}
	}

	limit = int64(pageSize)
	offset = int64((page - 1) * pageSize)
	return
}

func toInterfacePtr[T any](v T) *interface{} {
	var i interface{} = v
	return &i
}

func buildPagination(currentPage, pageSize *int64, totalMatched int64) *struct {
	CurrentPage *int64 `json:"currentPage,omitempty"`
	PageSize    *int64 `json:"pageSize,omitempty"`
	TotalItems  *int64 `json:"totalItems,omitempty"`
	TotalPages  *int64 `json:"totalPages,omitempty"`
} {
	var (
		page     int64 = 1
		size     int64 = 10
		hasPages bool  = false
	)

	if currentPage != nil && *currentPage > 0 {
		page = *currentPage
	}

	if pageSize != nil && *pageSize > 0 {
		size = *pageSize
		hasPages = true
	}

	var totalPages int64 = 1
	if hasPages {
		totalPages = totalMatched / size
		if totalMatched%size != 0 {
			totalPages++
		}

		if totalPages < 1 {
			totalPages = 1
		}
	}
	currentPageVal := page
	pageSizeVal := size
	totalItemsVal := totalMatched

	return &struct {
		CurrentPage *int64 "json:\"currentPage,omitempty\""
		PageSize    *int64 "json:\"pageSize,omitempty\""
		TotalItems  *int64 "json:\"totalItems,omitempty\""
		TotalPages  *int64 "json:\"totalPages,omitempty\""
	}{
		PageSize:    &pageSizeVal,
		CurrentPage: &currentPageVal,
		TotalItems:  &totalItemsVal,
		TotalPages:  &totalPages,
	}
}
