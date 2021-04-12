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

type twitchChannelInfo struct {
	StartTime time.Time
	EndTime   time.Time
	Oracles   []string
	IsLive    bool
}

type Session struct {
	name        string
	client      *helix.Client
	isConnected bool
	twitchData  map[string]*twitchChannelInfo
}

var activeOracles map[string]*Session

func (t *Session) Close() error {
	t.isConnected = false

	return utils.WriteGobToDisk(constants.DataPath+"/"+t.name+".gob", t.twitchData)
}

func GetSession(s *discordgo.Session) *Session {
	return activeOracles[s.State.SessionID]
}

func New(id string, secret string, name string) (*Session, error) {
	session := &Session{}
	session.name = name
	var err error

	session.client, err = helix.NewClient(&helix.Options{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURI:  "http://localhost",
	})
	if err != nil {
		return session, err
	}

	session.twitchData = make(map[string]*twitchChannelInfo)

	err = readGobFromDisk(constants.DataPath+"/"+session.name+".gob", &session.twitchData)
	if errors.Is(err, os.ErrNotExist) {
		utils.Log.Warn("Oracle info does not exist on disk. Will be created on shutdown.")
		err = nil
	}

	return session, err
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

func (t *Session) RegisterOracle(twitchId string, discordOracle string) (registered bool) {

	if t.twitchData[twitchId] == nil {
		// if twitch channel doesn't exist, register as new channel
		t.twitchData[twitchId] = &twitchChannelInfo{
			IsLive:  false,
			Oracles: []string{},
		}
	}

	// check if twitch session contains discord oracle, register otherwise
	if !t.containsOracle(twitchId, discordOracle) {
		t.twitchData[twitchId].Oracles = append(t.twitchData[twitchId].Oracles, discordOracle)
		return true
	}
	return false
}

func StartOracles(t *Session, s *discordgo.Session) {
	if t.isConnected {
		if activeOracles == nil {
			activeOracles = make(map[string]*Session)
			activeOracles[s.State.SessionID] = t
		} else {
			activeOracles[s.State.SessionID] = t
		}

		go monitorOracles(t, s)
	}
}

func (t *Session) UnregisterOracle(twitchId string, discordOracle string) (unregistered bool) {
	if t.containsOracle(twitchId, discordOracle) {
		t.twitchData[twitchId].Oracles = remove(t.twitchData[twitchId].Oracles, t.getOracleIdx(twitchId, discordOracle))
		// check if Oracles are empty and if so, delete channel from twitch Session
		if len(t.twitchData[twitchId].Oracles) == 0 {
			utils.Log.Debugf("No more channels monitoring for %v. Stopping oracle.\n", twitchId)
			delete(t.twitchData, twitchId)
		}
		return true
	}
	return false
}

func (t *Session) containsOracle(c string, d string) bool {
	if t.twitchData[c] == nil {
		return false
	}
	for _, discordChannel := range t.twitchData[c].Oracles {
		if d == discordChannel {
			return true
		}
	}
	return false
}

func (t *Session) getOracleIdx(c string, d string) int {
	if t.twitchData[c] == nil {
		return -1
	}
	for i, discordChannel := range t.twitchData[c].Oracles {
		if d == discordChannel {
			return i
		}
	}
	return -1
}

func monitorOracles(t *Session, s *discordgo.Session) {
	for t.isConnected {
		var queryChannels []string

		for tc := range t.twitchData {
			queryChannels = append(queryChannels, tc)
		}

		utils.Log.Debug("Sending query request to Twitch.")
		resp, err := t.client.GetStreams(&helix.StreamsParams{
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
		for twitchChannel, tcInfo := range t.twitchData {
			for _, streams := range resp.Data.Streams {
				if streams.UserLogin == twitchChannel && streams.Type == "live" {
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

		for twitchChannel, tcInfo := range t.twitchData {
			if !tcInfo.IsLive && !tcInfo.StartTime.IsZero() && time.Since(tcInfo.StartTime) > constants.TwitchStateChangeTime {
				tcInfo.IsLive = true
				for _, discordChannel := range tcInfo.Oracles {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is online! Watch at http://twitch.tv/"+twitchChannel)
				}
			} else if tcInfo.IsLive && !tcInfo.EndTime.IsZero() && time.Since(tcInfo.EndTime) > constants.TwitchStateChangeTime {
				tcInfo.IsLive = false
				for _, discordChannel := range tcInfo.Oracles {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is now offline!")
				}
			}
		}
		time.Sleep(constants.TwitchQueryInterval)
	}

	delete(activeOracles, s.State.SessionID)
}

func readGobFromDisk(path string, o *map[string]*twitchChannelInfo) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return gob.NewDecoder(file).Decode(o)
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
