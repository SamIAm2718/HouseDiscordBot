package constants

import "time"

const (
	TwitchQueryInterval         = time.Second * 10
	TwitchStateChangeTime       = time.Second * 90
	TwitchLiveMessageUpdateTime = time.Minute * 5
)
