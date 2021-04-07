package twitch

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/constants"
	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

type (
	twitchOracles map[string][]string
	twitchStates  map[string]bool
)

type oracleInfo struct {
	Oracles twitchOracles
	States  twitchStates
}

type TwitchSession struct {
	name        string
	client      *helix.Client
	info        *oracleInfo
	isConnected bool
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

	session.info = &oracleInfo{
		make(twitchOracles),
		make(twitchStates),
	}

	err = utils.ReadJSONFromDisk(constants.DataPath+"/"+session.name+".json", session.info)
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

	return utils.WriteJSONToDisk(constants.DataPath+"/"+t.name+".json", t.info)
}

func (t *TwitchSession) ContainsOracle(c string, d string) bool {
	for _, discordChannel := range t.info.Oracles[c] {
		if d == discordChannel {
			return true
		}
	}

	return false
}

func (t *TwitchSession) RegisterOracle(c string, d string) {
	if t.info.Oracles[c] != nil {
		if !t.ContainsOracle(c, d) {
			t.info.Oracles[c] = append(t.info.Oracles[c], d)
		}
	} else {
		t.info.Oracles[c] = []string{d}
	}
}

func (t *TwitchSession) UnregisterOracle(c string, d string) {
	if t.ContainsOracle(c, d) {
		for i, discordChannel := range t.info.Oracles[c] {
			if d == discordChannel {
				t.info.Oracles[c][i] = t.info.Oracles[c][len(t.info.Oracles[c])-1]
				t.info.Oracles[c][len(t.info.Oracles[c])-1] = ""
				t.info.Oracles[c] = t.info.Oracles[c][:len(t.info.Oracles[c])-1]
				return
			}
		}
	}
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
		queryChannels := []string{}

		for twitchChannel := range t.info.Oracles {
			if len(t.info.Oracles[twitchChannel]) == 0 {
				utils.Log.Infof("No more channels monitoring for %v. Shutting down oracle.\n", twitchChannel)
				delete(t.info.Oracles, twitchChannel)
				delete(t.info.States, twitchChannel)
			} else {
				queryChannels = append(queryChannels, twitchChannel)
			}
		}

		utils.Log.Debug("Sending query request to Twitch.")

		resp, err := t.client.GetStreams(&helix.StreamsParams{
			UserLogins: queryChannels,
		})
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"error": err}).Error("Failed to query twitch.")
		}

		utils.Log.Debugf("Twitch Response: %+v\n", resp)

		for twitchChannel := range t.info.Oracles {
			foundMatch := false
			for _, streams := range resp.Data.Streams {
				foundMatch = strings.EqualFold(twitchChannel, streams.UserLogin)

				if foundMatch {
					if streams.Type == "live" && !t.info.States[twitchChannel] {
						t.info.States[twitchChannel] = true
						for _, discordChannel := range t.info.Oracles[twitchChannel] {
							s.ChannelMessageSend(discordChannel, twitchChannel+" is online! Watch at http://twitch.tv/"+twitchChannel)
						}
					}
					break
				}
			}

			if !foundMatch && t.info.States[twitchChannel] {
				t.info.States[twitchChannel] = false
				for _, discordChannel := range t.info.Oracles[twitchChannel] {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is now offline!")
				}
			}
		}
		time.Sleep(time.Minute)
	}

	delete(activeOracles, s.State.SessionID)
}
