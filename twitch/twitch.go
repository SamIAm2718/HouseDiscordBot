package twitch

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SamIAm2718/HouseDiscordBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix"
)

type (
	TwitchOracles map[string][]string
	TwitchStates  map[string]bool
)

type TwitchSession struct {
	client      *helix.Client
	dataPath    string
	oracles     TwitchOracles
	states      TwitchStates
	isConnected bool
}

var activeOracles map[string]*TwitchSession

func GetSession(s *discordgo.Session) *TwitchSession {
	return activeOracles[s.State.SessionID]
}

func New(id string, secret string, path string) (*TwitchSession, error) {
	session := &TwitchSession{}
	var err error

	session.client, err = helix.NewClient(&helix.Options{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURI:  "http://localhost",
	})
	if err != nil {
		return session, err
	}
	session.dataPath = path
	session.oracles = make(TwitchOracles)
	session.states = make(TwitchStates)

	err = utils.ReadJSONFromDisk(session.dataPath+"/twitchoracles.json", &session.oracles)
	if err != nil {
		return session, err
	}

	err = utils.ReadJSONFromDisk(session.dataPath+"/twitchstates.json", &session.states)

	return session, err
}

func (t *TwitchSession) Open() error {
	resp, err := t.client.RequestAppAccessToken([]string{""})
	if err != nil {
		return err
	} else if resp.Data.AccessToken == "" {
		return errors.New("failure getting Access token")
	}
	t.client.SetAppAccessToken(resp.Data.AccessToken)
	t.isConnected = true

	return nil
}

func (t *TwitchSession) Close() error {
	t.isConnected = false

	err := utils.WriteJSONToDisk(t.dataPath+"/twitchoracles.json", t.oracles)
	if err != nil {
		return err
	}

	return utils.WriteJSONToDisk(t.dataPath+"/twitchstates.json", t.states)
}

func (t *TwitchSession) ContainsOracle(c string, d string) bool {
	for _, discordChannel := range t.oracles[c] {
		if d == discordChannel {
			return true
		}
	}

	return false
}

func (t *TwitchSession) RegisterOracle(c string, d string) {
	if t.oracles[c] != nil {
		if !t.ContainsOracle(c, d) {
			t.oracles[c] = append(t.oracles[c], d)
		}
	} else {
		t.oracles[c] = []string{d}
	}
}

func (t *TwitchSession) UnregisterOracle(c string, d string) {
	if t.ContainsOracle(c, d) {
		for i, discordChannel := range t.oracles[c] {
			if d == discordChannel {
				t.oracles[c][i] = t.oracles[c][len(t.oracles[c])-1]
				t.oracles[c][len(t.oracles[c])-1] = ""
				t.oracles[c] = t.oracles[c][:len(t.oracles[c])-1]
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

		for twitchChannel := range t.oracles {
			if len(t.oracles[twitchChannel]) == 0 {
				fmt.Printf("No more channels monitoring for %v. Shutting down oracle.\n", twitchChannel)
				delete(t.oracles, twitchChannel)
				delete(t.oracles, twitchChannel)
			} else {
				queryChannels = append(queryChannels, twitchChannel)
			}
		}

		resp, err := t.client.GetStreams(&helix.StreamsParams{
			UserLogins: queryChannels,
		})
		if err != nil {
			fmt.Println("Failed to query twitch", err)
		}

		for twitchChannel := range t.oracles {
			foundMatch := false
			for _, streams := range resp.Data.Streams {
				foundMatch = strings.EqualFold(twitchChannel, streams.UserLogin)

				if foundMatch {
					if streams.Type == "live" && !t.states[twitchChannel] {
						t.states[twitchChannel] = true
						for _, discordChannel := range t.oracles[twitchChannel] {
							s.ChannelMessageSend(discordChannel, twitchChannel+" is online! Watch at http://twitch.tv/"+twitchChannel)
						}
					}
					break
				}
			}

			if !foundMatch && t.states[twitchChannel] {
				t.states[twitchChannel] = false
				for _, discordChannel := range t.oracles[twitchChannel] {
					s.ChannelMessageSend(discordChannel, twitchChannel+" is now offline!")
				}
			}
		}
		time.Sleep(2 * time.Minute)
	}

	delete(activeOracles, s.State.SessionID)
}
