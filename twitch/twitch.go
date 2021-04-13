package twitch

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

type discordChannel struct {
	ChannelID            string // ID of discord channel
	LiveNotificationSent bool   // Whether or not a channel was notified of being live
}

type twitchChannelInfo struct {
	DisplayName     string                       // Twitch display name
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

// Map of Discord sessions to twitch sessions
var activeSessions map[string]*Session

func (t *Session) Close() error {
	t.isConnected = false

	return utils.WriteGobToDisk(constants.DataPath, t.name, t.twitchData)
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

func (t *Session) Open() error {
	resp, err := t.client.RequestAppAccessToken([]string{""})
	if err != nil {
		return err
	} else if resp.Data.AccessToken == "" {
		return errors.New("failed to get access token")
	}
	t.client.SetAppAccessToken(resp.Data.AccessToken)
	t.isConnected = true

	return nil
}

func (t *Session) RegisterChannel(twitchID string, discordGuildID string, discordChannelID string) (registered bool) {

	if t.twitchData[twitchID] == nil {
		// if twitch channel doesn't exist, register as new channel
		t.twitchData[twitchID] = &twitchChannelInfo{
			DiscordChannels: make(map[string][]*discordChannel),
		}
	}

	// check if twitch session contains discord oracle, register otherwise
	if t.getChannelIdx(twitchID, discordGuildID, discordChannelID) < 0 {
		dc := &discordChannel{
			ChannelID:            discordChannelID,
			LiveNotificationSent: false,
		}
		t.twitchData[twitchID].DiscordChannels[discordGuildID] = append(t.twitchData[twitchID].DiscordChannels[discordGuildID], dc)
		return true
	}

	return false
}

// Adds session to activeSessions and begins to monitor twitch
func StartMonitoring(t *Session, s *discordgo.Session) {
	if t.isConnected {
		if activeSessions == nil {
			activeSessions = make(map[string]*Session)
			activeSessions[s.State.SessionID] = t
		} else {
			activeSessions[s.State.SessionID] = t
		}

		go monitorChannels(t, s)
	}
}

func (t *Session) UnregisterChannel(twitchID string, discordGuildID string, discordChannelID string) (unregistered bool) {
	if t.getChannelIdx(twitchID, discordGuildID, discordChannelID) >= 0 {
		t.twitchData[twitchID].DiscordChannels[discordGuildID] = remove(t.twitchData[twitchID].DiscordChannels[discordGuildID], t.getChannelIdx(twitchID, discordGuildID, discordChannelID))

		// Check if no more channels in Discord server are monitoring for Twitch channel and if so delete from map
		if len(t.twitchData[twitchID].DiscordChannels[discordGuildID]) == 0 {
			delete(t.twitchData[twitchID].DiscordChannels, discordGuildID)
		}

		// check if Oracles are empty and if so, delete channel from twitch Session
		if len(t.twitchData[twitchID].DiscordChannels) == 0 {
			utils.Log.Debugf("No more channels monitoring for %v. Deleting Twitch info for %v.\n", twitchID, twitchID)
			delete(t.twitchData, twitchID)
		}
		return true
	}
	return false
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
		var queryChannels []string

		for twitchChannel := range ts.twitchData {
			queryChannels = append(queryChannels, twitchChannel)
		}

		utils.Log.Debug("Sending query request to Twitch.")
		resp, err := ts.client.GetStreams(&helix.StreamsParams{
			UserLogins: queryChannels,
		})
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"error": err}).Error("Failed to query twitch.")
		}

		if constants.Debug {
			empJSON, err := json.MarshalIndent(resp.Data.Streams, "", "  ")
			if err != nil {
				utils.Log.WithFields(logrus.Fields{"error": err}).Debug("Error marshaling Twitch JSON response.")
			} else {
				utils.Log.Debugf("Twitch Response: %+v\n", string(empJSON))
			}
		}

		// populate start/end time
	OUTER:
		for twitchChannel, tcInfo := range ts.twitchData {
			for _, streams := range resp.Data.Streams {
				if streams.UserLogin == twitchChannel && streams.Type == "live" {
					tcInfo.DisplayName = streams.UserName
					tcInfo.StartTime = streams.StartedAt
					tcInfo.EndTime = time.Time{}
					continue OUTER
				}
			}

			// stream not found, update times
			tcInfo.StartTime = time.Time{}
			if tcInfo.EndTime.IsZero() {
				tcInfo.EndTime = time.Now()
			}
		}

		for _, tcInfo := range ts.twitchData {
			if !tcInfo.StartTime.IsZero() && time.Since(tcInfo.StartTime) > constants.TwitchStateChangeTime {
				for _, discordChannels := range tcInfo.DiscordChannels {
					for _, discordChannel := range discordChannels {
						if !discordChannel.LiveNotificationSent {
							discordChannel.LiveNotificationSent = true
							go ds.ChannelMessageSend(discordChannel.ChannelID, tcInfo.DisplayName+" is online! Watch at http://twitch.tv/"+tcInfo.DisplayName)
						}
					}
				}
			} else if !tcInfo.EndTime.IsZero() && time.Since(tcInfo.EndTime) > constants.TwitchStateChangeTime {
				for _, discordChannels := range tcInfo.DiscordChannels {
					for _, discordChannel := range discordChannels {
						if discordChannel.LiveNotificationSent {
							discordChannel.LiveNotificationSent = false
							go ds.ChannelMessageSend(discordChannel.ChannelID, tcInfo.DisplayName+" is now offline!")
						}
					}
				}
			}
		}
		time.Sleep(constants.TwitchQueryInterval)
	}

	delete(activeSessions, ds.State.SessionID)
}

func readGobFromDisk(path string, name string, o *map[string]*twitchChannelInfo) error {
	file, err := os.Open(path + "/" + name + ".gob")
	if err != nil {
		return err
	}

	return gob.NewDecoder(file).Decode(o)
}

func remove(s []*discordChannel, i int) []*discordChannel {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
