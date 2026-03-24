package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrAccountDisabled   = errors.New("account disabled")
	ErrAccountLocked     = errors.New("account locked")
	ErrUsernameExists    = errors.New("username already exists")
	ErrProjectNotFound   = errors.New("project not found")
	ErrSeasonNotFound    = errors.New("season not found")
	ErrEpisodeNotFound   = errors.New("episode not found")
	ErrModelNotFound     = errors.New("model config not found")
	ErrOldPasswordWrong  = errors.New("old password is incorrect")
)
