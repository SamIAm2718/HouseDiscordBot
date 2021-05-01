package constants

import "errors"

var (
	ErrEmptyAccessToken = errors.New("access token retrieved is empty")
	ErrInvalidToken     = errors.New("access token failed to validate or refresh")
)

var (
	ErrTwitchUserDoesNotExist = errors.New("twitch user does not exist")
	ErrTwitchUserRegistered   = errors.New("twitch user is already registered to discord channel")
)
