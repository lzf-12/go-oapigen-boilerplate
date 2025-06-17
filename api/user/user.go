package user

import (
	"context"
	"fmt"
	"oapi-to-rest/pkg/errlib"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type UserImpl struct {
	Sqlx *sqlx.DB
}

var _ StrictServerInterface = (*UserImpl)(nil)

const (
	defaultPage     int64 = 1
	defaultPageSize int64 = 20
	maxPageSize     int64 = 100
)

type UserRow struct {
	User
	CountMatched int64 `db:"count_matched"`
}

func (u *UserImpl) GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error) {

	var result PaginatedUserResponse
	params := request.Params

	// build the base query
	baseQuery := `SELECT u.id, u.email, u.first_name, u.last_name, u.is_active, COUNT(*) OVER() AS count_matched FROM users u`

	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if params.Email != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.email = $%d", argIndex))
		args = append(args, *params.Email)
		argIndex++
	}

	if params.IsActive != nil {
		isActive, err := strconv.ParseBool(*params.IsActive)
		if err != nil {
			return GetUser500JSONResponse{}, err
		}
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_active = $%d", argIndex))
		args = append(args, isActive)
		argIndex++
	}

	// construct where
	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	limit, offset := parsePaginationParams(params.Page, params.PageSize)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var userRows []UserRow
	err := u.Sqlx.SelectContext(ctx, &userRows, query, args...)
	if err != nil {
		return GetUser500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBQuery)
	}

	var totalMatched int64
	var users []User

	for _, userRow := range userRows {
		users = append(users, userRow.User)
		totalMatched = userRow.CountMatched
	}

	result.Data = &users
	result.Filters = toInterfacePtr(params)
	result.Pagination = buildPagination(params.Page, params.PageSize, totalMatched)

	return GetUser200JSONResponse(result), nil
}

func parsePaginationParams(pagePtr, pageSizePtr *int64) (limit, offset int64) {
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
