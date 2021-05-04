package twitch

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
)

type discordChannel struct {
	ChannelID            string    // ID of discord channel
	LiveMessageID        string    // ID of LiveMessage
	UpdateTime           time.Time // Time the message was last updated
	LiveNotificationSent bool      // Whether or not a channel was notified of being live
}

type gameInfo struct {
	GameName  string    // Name of game
	StartTime time.Time // Time user switched to game
	EndTime   time.Time // Time user switched off game
}

type twitchChannelInfo struct {
	DisplayName     string                       // Twitch display name
	LogoURL         string                       // URL of Twitch logo
	StreamData      *helix.Stream                // Stream response sent by
	GameList        []*gameInfo                  // List of games played by streamer
	StartTime       time.Time                    // Start time of stream
	EndTime         time.Time                    // End time of stream
	DiscordChannels map[string][]*discordChannel // Map of Discord guild IDs to discordChannel
}

type Session struct {
	name        string                        // Name of the Twitch session
	client      *helix.Client                 // Helix client for sending HTTP requests to twitch
	isConnected bool                          // Status of Helix client connection to twitch
	twitchData  map[string]*twitchChannelInfo // Map of twitch channel to its info
}

var (
	activeSessions map[string]*Session // Map of Discord sessions to twitch sessions
	guildStatus    map[string]bool     // Map of Guild ID to status of guild connection
)

func init() {
	activeSessions = make(map[string]*Session)
	guildStatus = make(map[string]bool)
}

func (t *Session) Close() error {
	t.isConnected = false

	for _, tcInfo := range t.twitchData {
		for gID, status := range guildStatus {
			if !status {
				delete(tcInfo.DiscordChannels, gID)
			}
		}
	}

	return utils.WriteGobToDisk(constants.DataPath, t.name, t.twitchData)
}

// Returns twitch channels being monitored by discord channel
func (s *Session) GetMonitoredChannels(channelID string) []string {
	channels := []string{}

	for tc, tcInfo := range s.twitchData {
		for _, discordChannels := range tcInfo.DiscordChannels {
			for _, discordChannel := range discordChannels {
				if discordChannel.ChannelID == channelID {
					channels = append(channels, s.twitchData[tc].DisplayName)
				}
			}
		}
	}

	return channels
}

func GetSession(s *discordgo.Session) *Session {
	return activeSessions[s.State.SessionID]
}

func New(id string, secret string, name string) (t *Session, err error) {
	t = &Session{}
	t.name = name

	t.client, err = helix.NewClient(&helix.Options{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURI:  "http://localhost",
	})
	if err != nil {
		return t, err
	}

	t.twitchData = make(map[string]*twitchChannelInfo)

	err = readGobFromDisk(constants.DataPath, t.name, &t.twitchData)
	if errors.Is(err, os.ErrNotExist) {
		utils.Log.Warn("Twitch session info does not exist on disk. Will be created on shutdown.")
		err = nil
	}

	return t, err
}

// Attempts to use client ID and secret to get Auth token from twitch.
// If successful then set the session state to connected.
func (t *Session) GetAuthToken() error {
	resp, err := t.client.RequestAppAccessToken([]string{""})
	if err != nil {
		return err
	} else if resp.Data.AccessToken == "" {
		return constants.ErrEmptyAccessToken
	}
	t.client.SetAppAccessToken(resp.Data.AccessToken)
	t.isConnected = true

	return nil
}

// Registers a Discord Channel to monitor the live state of a twitch channel
func (t *Session) RegisterChannel(twitchID string, discordGuildID string, discordChannelID string) (registered error) {
	// if twitch channel doesn't exist, register as new channel
	if t.twitchData[twitchID] == nil {

		// we need to obtain the profile picture url and display name for the twitch channel
		if validateAndRefreshAuthToken(t) {
			resp, err := t.client.GetUsers(&helix.UsersParams{Logins: []string{twitchID}})
			if err != nil {
				utils.Log.WithError(err).Error("Failed to query twitch.")
			}

			if len(resp.Data.Users) == 0 {
				return constants.ErrTwitchUserDoesNotExist
			}

			// register the twitch information channel
			t.twitchData[twitchID] = &twitchChannelInfo{
				DisplayName:     resp.Data.Users[0].DisplayName,
				LogoURL:         resp.Data.Users[0].ProfileImageURL,
				DiscordChannels: make(map[string][]*discordChannel),
			}
		} else {
			return constants.ErrInvalidToken
		}
	}

	// check if twitch session contains discord oracle, register otherwise
	if t.getChannelIdx(twitchID, discordGuildID, discordChannelID) < 0 {
		dc := &discordChannel{
			ChannelID:            discordChannelID,
			LiveNotificationSent: false,
		}
		t.twitchData[twitchID].DiscordChannels[discordGuildID] = append(t.twitchData[twitchID].DiscordChannels[discordGuildID], dc)

		// Writes the data to the disk in case of crash
		if err := utils.WriteGobToDisk(constants.DataPath, t.name, t.twitchData); err != nil {
			utils.Log.WithError(err).Error("Error writing data to disk.")
		}

		return nil
	}

	return constants.ErrTwitchUserRegistered
}

// Sets the current guild as active
func SetGuildActive(guildID string) {
	guildStatus[guildID] = true
}

// Sets the current guild as inactive
func SetGuildInactive(guildID string) {
	guildStatus[guildID] = false
}

// Sets current guild as unavailable
func SetGuildUnavailable(guildID string) {
	delete(guildStatus, guildID)
}

// Adds session to activeSessions if it is connected to Twitch and begins to monitor Twitch
func StartMonitoring(t *Session, s *discordgo.Session) {
	if t.isConnected {
		activeSessions[s.State.SessionID] = t

		go monitorChannels(t, s)
	}
}

// Unregisters a Discord Channel from monitor the live state of a Twitch channel
func (t *Session) UnregisterChannel(twitchID string, discordGuildID string, discordChannelID string) (unregistered bool) {
	if channelIdx := t.getChannelIdx(twitchID, discordGuildID, discordChannelID); channelIdx >= 0 {
		t.twitchData[twitchID].DiscordChannels[discordGuildID] = remove(t.twitchData[twitchID].DiscordChannels[discordGuildID], channelIdx)

		// Check if no more channels in Discord server are monitoring for Twitch channel and if so delete from map
		if len(t.twitchData[twitchID].DiscordChannels[discordGuildID]) == 0 {
			delete(t.twitchData[twitchID].DiscordChannels, discordGuildID)
		}

		// check if Oracles are empty and if so, delete channel from twitch Session
		if len(t.twitchData[twitchID].DiscordChannels) == 0 {
			utils.Log.Debugf("No more channels monitoring for %v. Deleting Twitch info for %v.\n", twitchID, twitchID)
			delete(t.twitchData, twitchID)
		}

		// Writes the data to the disk in case of crash
		if err := utils.WriteGobToDisk(constants.DataPath, t.name, t.twitchData); err != nil {
			utils.Log.WithError(err).Error("Error writing data to disk.")
		}

		return true
	}
	return false
}

func createDiscordLiveEmbedMessage(t *twitchChannelInfo) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField
	if t.StreamData.GameName != "" {
		fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Playing",
				Value:  t.StreamData.GameName + " ",
				Inline: true,
			},
			{
				Name:   "Viewers",
				Value:  fmt.Sprint(t.StreamData.ViewerCount) + " ",
				Inline: true,
			},
		}
	} else {
		fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Viewers",
				Value:  fmt.Sprint(t.StreamData.ViewerCount) + " ",
				Inline: true,
			},
		}
	}

	embed := &discordgo.MessageEmbed{
		URL:         "https://www.twitch.tv/" + t.DisplayName,
		Type:        "",
		Title:       t.StreamData.Title,
		Description: "",
		Timestamp:   "",
		Color:       0x00ff00,
		Footer: &discordgo.MessageEmbedFooter{
			Text:         "Streaming for " + time.Since(t.StartTime).Round(time.Second).String(),
			IconURL:      "",
			ProxyIconURL: ""},
		Image: &discordgo.MessageEmbedImage{
			URL:      strings.Replace(strings.Replace(t.StreamData.ThumbnailURL, "{width}", "1920", -1), "{height}", "1080", -1),
			ProxyURL: "",
			Width:    1920,
			Height:   1080},
		Thumbnail: &discordgo.MessageEmbedThumbnail{},
		Video:     &discordgo.MessageEmbedVideo{},
		Provider:  &discordgo.MessageEmbedProvider{},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    t.DisplayName,
			IconURL: t.LogoURL,
		},
		Fields: fields,
	}

	return embed
}

func createDiscordOfflineEmbedMessage(t *twitchChannelInfo) *discordgo.MessageEmbed {
	games := ""

	for i, game := range t.GameList {
		if game.GameName != "" {
			games += fmt.Sprint(i+1) + ". " + game.GameName + " for " + game.EndTime.Sub(game.StartTime).Round(time.Second).String() + "\n"
		} else {
			games += fmt.Sprint(i+1) + ". Nothing for " + game.EndTime.Sub(game.StartTime).Round(time.Second).String() + "\n"
		}
	}

	embed := &discordgo.MessageEmbed{
		URL:   "",
		Type:  "",
		Title: "",
		Description: "**Started at:** " + t.StartTime.Format("01/02/2006 15:04 MST") + "\n" +
			"__**Ended at:** " + t.EndTime.Format("01/02/2006 15:04 MST") + "__\n" +
			"**Total time streamed:** " + t.EndTime.Sub(t.StartTime).Round(time.Second).String() + "\n\n" +
			"**Games Played**\n" + games,
		Timestamp: "",
		Color:     0xff0000,
		Footer:    &discordgo.MessageEmbedFooter{},
		Image:     &discordgo.MessageEmbedImage{},
		Thumbnail: &discordgo.MessageEmbedThumbnail{},
		Video:     &discordgo.MessageEmbedVideo{},
		Provider:  &discordgo.MessageEmbedProvider{},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    t.DisplayName + " was online.",
			IconURL: t.LogoURL,
		},
		Fields: []*discordgo.MessageEmbedField{},
	}

	return embed
}

// Returns -1 if oracle isn't present or the index of the oracle if it is
func (t *Session) getChannelIdx(twitchID string, discordGuildID string, discordChannelID string) int {
	if t.twitchData[twitchID] == nil {
		return -1
	}
	for i, d := range t.twitchData[twitchID].DiscordChannels[discordGuildID] {
		if d.ChannelID == discordChannelID {
			return i
		}
	}
	return -1
}

func monitorChannels(ts *Session, ds *discordgo.Session) {
	for ts.isConnected {
		if validateAndRefreshAuthToken(ts) {
			var queryChannels []string

			for twitchChannel := range ts.twitchData {
				queryChannels = append(queryChannels, twitchChannel)
			}

			resp, err := ts.client.GetStreams(&helix.StreamsParams{
				UserLogins: queryChannels,
			})
			if err != nil {
				utils.Log.WithError(err).Error("Failed to query twitch.")
			}

			if constants.DebugTwitchResponse {
				empJSON, err := json.MarshalIndent(resp, "", "  ")
				if err != nil {
					utils.Log.WithError(err).Debug("Error marshaling Twitch JSON response.")
				} else {
					utils.Log.Debugf("Twitch getStreams request Response: %+v\n", string(empJSON))
				}
			}

			// populate twitch info start/end time
		OUTER:
			for twitchChannel, tcInfo := range ts.twitchData {
				for _, streams := range resp.Data.Streams {
					if streams.UserLogin == twitchChannel && streams.Type == "live" {
						tcInfo.StreamData = &streams
						tcInfo.StartTime = streams.StartedAt
						tcInfo.EndTime = time.Time{}

						if len(tcInfo.GameList) == 0 {
							tcInfo.GameList = []*gameInfo{
								{
									GameName:  streams.GameName,
									StartTime: streams.StartedAt,
									EndTime:   time.Time{},
								},
							}
						} else if tcInfo.GameList[len(tcInfo.GameList)-1].GameName != streams.GameName {
							tcInfo.GameList[len(tcInfo.GameList)-1].EndTime = time.Now()

							tcInfo.GameList = append(tcInfo.GameList, &gameInfo{
								GameName:  streams.GameName,
								StartTime: time.Now(),
								EndTime:   time.Time{},
							})
						}

						continue OUTER
					}
				}

				// stream not found, update times
				tcInfo.StreamData = nil
				if tcInfo.EndTime.IsZero() {
					tcInfo.EndTime = time.Now()
				}
			}

			sendNotifications(ts, ds)
		}

		time.Sleep(constants.TwitchQueryInterval)
	}

	delete(activeSessions, ds.State.SessionID)
}

func readGobFromDisk(path string, name string, o *map[string]*twitchChannelInfo) error {
	if file, err := os.Open(path + "/" + name + ".gob"); err != nil {
		return err
	} else {
		return gob.NewDecoder(file).Decode(o)
	}
}

func remove(s []*discordChannel, i int) []*discordChannel {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func sendNotifications(ts *Session, ds *discordgo.Session) {
	for _, tcInfo := range ts.twitchData {
		if tcInfo.StreamData != nil && time.Since(tcInfo.StartTime) > constants.TwitchStateChangeTime {
			for guild, discordChannels := range tcInfo.DiscordChannels {
				if connected, available := guildStatus[guild]; available && connected {
					for _, discordChannel := range discordChannels {
						if !discordChannel.LiveNotificationSent {
							discordChannel.LiveNotificationSent = true
							go sendLiveNotification(ds, discordChannel, tcInfo)
						} else if discordChannel.LiveMessageID != "" && time.Since(discordChannel.UpdateTime) > constants.TwitchLiveMessageUpdateTime {
							go updateLiveNotification(ds, discordChannel, tcInfo)
						}
					}
				}
			}
		} else if tcInfo.StreamData == nil && time.Since(tcInfo.EndTime) > constants.TwitchStateChangeTime {
			for guild, discordChannels := range tcInfo.DiscordChannels {
				if connected, available := guildStatus[guild]; available && connected {
					for _, discordChannel := range discordChannels {
						if discordChannel.LiveNotificationSent && discordChannel.LiveMessageID != "" {
							discordChannel.LiveNotificationSent = false
							go sendOfflineNotification(ds, discordChannel, tcInfo)
						}
					}
				}
			}
		}
	}
}

func sendLiveNotification(ds *discordgo.Session, dc *discordChannel, tci *twitchChannelInfo) {
	if m, err := ds.ChannelMessageSendEmbed(dc.ChannelID, createDiscordLiveEmbedMessage(tci)); err != nil {
		utils.Log.WithError(err).Error("Error sending Discord message.")
	} else {
		dc.LiveMessageID = m.ID
		dc.UpdateTime = time.Now()
	}
}

func sendOfflineNotification(ds *discordgo.Session, dc *discordChannel, tci *twitchChannelInfo) {
	tci.GameList[len(tci.GameList)-1].EndTime = time.Now()

	if _, err := ds.ChannelMessageEditEmbed(dc.ChannelID, dc.LiveMessageID, createDiscordOfflineEmbedMessage(tci)); err != nil {
		utils.Log.WithError(err).Error("Error updating Discord message.")
	}

	dc.LiveMessageID = ""
	dc.UpdateTime = time.Time{}
	tci.GameList = nil
}

func updateLiveNotification(ds *discordgo.Session, dc *discordChannel, tci *twitchChannelInfo) {
	if m, err := ds.ChannelMessageEditEmbed(dc.ChannelID, dc.LiveMessageID, createDiscordLiveEmbedMessage(tci)); err != nil {
		dc.LiveNotificationSent = false
		utils.Log.WithError(err).Error("Error updating Discord message.")
	} else {
		dc.LiveMessageID = m.ID
		dc.UpdateTime = time.Now()
	}
}

func validateAndRefreshAuthToken(ts *Session) bool {
	// Validate and refresh Twitch authorization token, if token valid
	if isValid, resp, err := ts.client.ValidateToken(ts.client.GetAppAccessToken()); err != nil {
		utils.Log.WithError(err).Error("Failed to validate Twitch authorization token.")
	} else if !isValid {
		ts.isConnected = false
		for !ts.isConnected {
			utils.Log.Debug("Attempting to get new Twitch authentication token.")
			if ts.GetAuthToken() != nil {
				utils.Log.WithError(err).Error("Failed to get new Twitch authorization token.")
				break
			}
		}

		if ts.isConnected {
			utils.Log.Debug("Successfully got new Twitch authentication token.")
			return true
		}
	} else if resp.StatusCode != 200 {
		utils.Log.WithField("StatusCode", resp.StatusCode).Error("HTTP Error returned from twitch.")
	} else {
		return true
	}

	return false
}
