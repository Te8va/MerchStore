package errors

import "errors"

var (
	ErrAlreadyRegistered   = errors.New("user with this username is already registered")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrItemNotFound        = errors.New("item not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrUserNotFound        = errors.New("user not found")
	ErrWrongPassword       = errors.New("wrong password provided")
	ErrNoLoginOrPassword   = errors.New("no login or password provided")
	ErrWrongAdminHeader    = errors.New("wrong admin header")
	ErrWrongMIME           = errors.New("wrong MIME type used")
	ErrWrongJSON           = errors.New("something is wrong in json")
)
