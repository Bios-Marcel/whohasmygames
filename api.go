package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Session struct {
	apiToken        string
	targetAccountId steamID
}

func NewSession(apiToken string, targetAccountId steamID) (*Session, error) {
	return &Session{
		apiToken:        apiToken,
		targetAccountId: targetAccountId,
	}, nil
}

// steamID represents the "database" ID from steam. While this is clearly
// a number, steam decides to be retarded and send strings anways.
type steamID string

type ownedGamesResponse struct {
	Response *ownedGamesContainer `json:"response"`
}

type ownedGamesContainer struct {
	Games []*ownedGame `json:"games"`
}

type ownedGame struct {
	AppID uint64 `json:"appid"`
	Name  string `json:"name"`
}

func (session *Session) getOwnedGames(friends []*friend) (map[steamID][]*ownedGame, error) {
	waitgroup := &sync.WaitGroup{}
	waitgroup.Add(len(friends))

	ownedGames := make(map[steamID][]*ownedGame, len(friends))
	for _, tempFren := range friends {
		fren := tempFren
		go func() {
			url := fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetOwnedGames/v1?key=%s&include_played_free_games=true&include_appinfo=true&steamid=%s", session.apiToken, fren.SteamID)
			request, err := makeGetRequest(url)
			if err != nil {
				return
			}

			result, err := http.DefaultClient.Do(request)
			if err != nil {
				return
			}

			ownedGamesResponse := &ownedGamesResponse{}
			decodeError := json.NewDecoder(result.Body).Decode(ownedGamesResponse)
			if decodeError != nil {
				return
			}

			ownedGames[fren.SteamID] = ownedGamesResponse.Response.Games
			waitgroup.Done()
		}()
	}

	waitgroup.Wait()

	return ownedGames, nil
}

type playerProfilesResponse struct {
	Response playerProfileContainer `json:"response"`
}

type playerProfileContainer struct {
	Players []*playerProfile `json:"players"`
}

type playerProfile struct {
	SteamID     steamID `json:"steamid"`
	Personaname string  `json:"personaname"`
	AvatarURL   string  `json:"avatar"`
}

func (session *Session) getFriendProfiles(friends []*friend) (map[steamID]*playerProfile, error) {
	if len(friends) == 0 {
		return nil, nil
	}

	mapping := make(map[steamID]*playerProfile, len(friends))

	//SteamAPI seems to only gives us up to 100 profiles at once.
	for i := 0; i < len(friends); i += 100 {
		friendBatch := friends[i:min(i+100, len(friends))]
		var urlBuilder strings.Builder
		for index, fren := range friendBatch {
			if index != 0 {
				urlBuilder.WriteRune(',')
			}
			urlBuilder.WriteString(string(fren.SteamID))
		}
		url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2?key=%s&steamids=%s", session.apiToken, urlBuilder.String())

		request, err := makeGetRequest(url)
		if err != nil {
			return nil, err
		}

		result, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}

		response := &playerProfilesResponse{}
		decodeError := json.NewDecoder(result.Body).Decode(response)
		if decodeError != nil {
			return nil, decodeError
		}

		for _, profile := range response.Response.Players {
			mapping[profile.SteamID] = profile
		}
	}

	return mapping, nil
}

type friendslistContainer struct {
	Friendslist friendsContainer `json:"friendslist"`
}

func (fc friendslistContainer) friends() []*friend {
	return fc.Friendslist.Friends
}

type friendsContainer struct {
	Friends []*friend `json:"friends"`
}

type friend struct {
	SteamID steamID `json:"steamid"`
}

func (session *Session) getOwnFriends() ([]*friend, error) {
	return session.getFriends(session.targetAccountId)
}

func (session *Session) getFriends(steamId steamID) ([]*friend, error) {
	request, err := makeGetRequest(fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetFriendList/v1?key=%s&steamid=%s&relationship=friend", session.apiToken, steamId))
	if err != nil {
		return nil, err
	}

	result, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	fc := &friendslistContainer{}
	json.NewDecoder(result.Body).Decode(fc)
	return fc.friends(), nil
}

type steamIDResponse struct {
	SteamID steamID `json:"steamid"`
}

func getSteamID(apiKey string, steamIdOrVanityName string) (steamID, error) {
	_, parseError := strconv.ParseUint(steamIdOrVanityName, 10, 64)
	if parseError == nil {
		return steamID(steamIdOrVanityName), nil
	}

	steamIdRequest, err := makeGetRequest(fmt.Sprintf("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v1?key=%s&vanityurl=%s", apiKey, steamIdOrVanityName))
	if err != nil {
		return "", err
	}

	result, err := http.DefaultClient.Do(steamIdRequest)
	if err != nil {
		return "", err
	}

	responseData := &steamIDResponse{}
	decodeError := json.NewDecoder(result.Body).Decode(responseData)
	if decodeError != nil {
		return "", decodeError
	}

	return responseData.SteamID, nil
}
