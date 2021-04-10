package twitch

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

type twitchChannelInfo struct {
	startTime time.Time
	endTime   time.Time
	Oracles   []string
	isLive    bool
}

type TwitchSession struct {
	name        string
	client      *helix.Client
	isConnected bool
	twitchData 	map[string]*twitchChannelInfo
}

var activeOracles map[string]*TwitchSession

func GetSession(s *discordgo.Session) *TwitchSession {
	return activeOracles[s.State.SessionID]
}

func New(id string, secret string, name string) (*TwitchSession, error) {
	session := &TwitchSession{}
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

	err = ReadGobFromDisk(constants.DataPath+"/"+session.name+".gob", session.twitchData)
	if errors.Is(err, os.ErrNotExist) {
		utils.Log.Warn("Oracle info does not exist on disk. Will be created on shutdown.")
		err = nil
	}
	return session, err
}

func (t *TwitchSession) Open() error {
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

func (t *TwitchSession) Close() error {
	t.isConnected = false

	return utils.WriteGobToDisk(constants.DataPath+"/"+t.name+".gob", t.twitchData)
}

func (t *TwitchSession) containsOracle(c string, d string) bool {
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

func (t *TwitchSession) getOracleIdx(c string, d string) int {
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

func (t *TwitchSession) RegisterOracle(twitchId string, discordOracle string) (registered bool) {

	if t.twitchData[twitchId] == nil {
		// if twitch channel doesn't exist, register as new channel
		t.twitchData[twitchId] = &twitchChannelInfo{
			isLive:  false,
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

func (t *TwitchSession) UnregisterOracle(twitchId string, discordOracle string) (unregistered bool) {
	if t.containsOracle(twitchId, discordOracle) {
		fmt.Println(t.twitchData[twitchId].Oracles)
		t.twitchData[twitchId].Oracles = remove(t.twitchData[twitchId].Oracles, t.getOracleIdx(twitchId, discordOracle))
		// check if Oracles are empty and if so, delete channel from twitch Session
		if len(t.twitchData[twitchId].Oracles) == 0 {
			delete(t.twitchData, twitchId)
		}
		return true
	}
	return false
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func StartOracles(t *TwitchSession, s *discordgo.Session) {
	if t.isConnected {
		if activeOracles == nil {
			activeOracles = make(map[string]*TwitchSession)
			activeOracles[s.State.SessionID] = t
		} else {
			activeOracles[s.State.SessionID] = t
		}

		go monitorOracles(t, s)
	}
}

func monitorOracles(t *TwitchSession, s *discordgo.Session) {
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
				log.Fatalf(err.Error())
			}

			utils.Log.Debugf("Twitch Response: %+v\n", string(empJSON))
		}

		// populate start/end time
		OUTER:
		for twitchChannel, tcInfo := range t.twitchData {
			for _, streams := range resp.Data.Streams {
				if streams.UserLogin == twitchChannel && streams.Type == "live" {
					tcInfo.startTime = streams.StartedAt
					tcInfo.endTime = time.Time{}
					continue OUTER
				}
			}

			// stream not found, update times
			tcInfo.startTime = time.Time{}
			if tcInfo.endTime.IsZero() {
				tcInfo.endTime = time.Now()
			}
		}

		for twitchChannel, tcInfo := range t.twitchData {
			if !tcInfo.isLive && !tcInfo.startTime.IsZero() && time.Since(tcInfo.startTime) > constants.TwitchStateChangeTime {
				tcInfo.isLive = true
				for _, discordChannel := range tcInfo.Oracles {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is online! Watch at http://twitch.tv/"+twitchChannel)
				}
			} else if tcInfo.isLive && !tcInfo.endTime.IsZero() && time.Since(tcInfo.endTime) > constants.TwitchStateChangeTime {
				tcInfo.isLive = false
				for _, discordChannel := range tcInfo.Oracles {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is now offline!")
				}
			}
		}
		time.Sleep(constants.TwitchQueryInterval)
	}

	delete(activeOracles, s.State.SessionID)
}

func ReadGobFromDisk(path string, o map[string]*twitchChannelInfo) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	encoder := gob.NewDecoder(file)
	err = encoder.Decode(&o)
	return err
}