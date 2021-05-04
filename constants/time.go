package constants

import "time"

const (
	DiscordMessageDeleteDelay   = time.Second * 30
	TwitchQueryInterval         = time.Second * 10
	TwitchStateChangeTime       = time.Second * 90
	TwitchLiveMessageUpdateTime = time.Second * 30
	TwitchThumbnailUpdateTime   = time.Minute * 5
	TwitchGameUpdateTime        = time.Second * 60
)
