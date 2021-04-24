//+build !demo

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Bios-Marcel/whohasmygames/httputility"
	"github.com/Bios-Marcel/whohasmygames/maths"
)

type steamSession struct {
	apiToken        string
	targetAccountId SteamID

	ownedGamesCache map[SteamID][]*Game
}

func NewSession(apiToken string, targetAccountId SteamID) (Session, error) {
	return &steamSession{
		apiToken:        apiToken,
		targetAccountId: targetAccountId,
		ownedGamesCache: make(map[SteamID][]*Game),
	}, nil
}

type friendslistContainer struct {
	Friendslist friendsContainer `json:"friendslist"`
}

func (fc friendslistContainer) friends() []*Friend {
	return fc.Friendslist.Friends
}

type friendsContainer struct {
	Friends []*Friend `json:"friends"`
}

func (session *steamSession) GetOwnFriends() ([]*Friend, error) {
	return session.GetFriends(session.targetAccountId)
}

func (session *steamSession) GetFriends(steamId SteamID) ([]*Friend, error) {
	request, err := httputility.MakeGetRequest(fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetFriendList/v1?key=%s&steamid=%s&relationship=friend", session.apiToken, steamId))
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

type playerProfilesResponse struct {
	Response playerProfileContainer `json:"response"`
}

type playerProfileContainer struct {
	Players []*PlayerProfile `json:"players"`
}

func (session *steamSession) GetFriendProfiles(friends []*Friend) (map[SteamID]*PlayerProfile, error) {
	if len(friends) == 0 {
		return nil, nil
	}

	mapping := make(map[SteamID]*PlayerProfile, len(friends))

	//SteamAPI seems to only gives us up to 100 profiles at once.
	for i := 0; i < len(friends); i += 100 {
		friendBatch := friends[i:maths.Min(i+100, len(friends))]
		var urlBuilder strings.Builder
		for index, fren := range friendBatch {
			if index != 0 {
				urlBuilder.WriteRune(',')
			}
			urlBuilder.WriteString(string(fren.SteamID))
		}
		url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2?key=%s&steamids=%s", session.apiToken, urlBuilder.String())

		request, err := httputility.MakeGetRequest(url)
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

type ownedGamesResponse struct {
	Response *ownedGamesContainer `json:"response"`
}

type ownedGamesContainer struct {
	Games []*Game `json:"games"`
}

func (session *steamSession) GetOwnedGames(friends []*Friend, forceRefresh bool) (map[SteamID][]*Game, error) {
	waitgroup := &sync.WaitGroup{}
	waitgroup.Add(len(friends))

	ownedGames := make(map[SteamID][]*Game, len(friends))
	for _, tempFren := range friends {
		fren := tempFren

		if !forceRefresh {
			cachedOwnedGames, avail := session.ownedGamesCache[fren.SteamID]
			if avail && cachedOwnedGames != nil {
				ownedGames[fren.SteamID] = cachedOwnedGames
				waitgroup.Done()
				continue
			}
		}

		go func() {
			url := fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetOwnedGames/v1?key=%s&include_played_free_games=true&include_appinfo=true&steamid=%s", session.apiToken, fren.SteamID)
			request, err := httputility.MakeGetRequest(url)
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
			session.ownedGamesCache[fren.SteamID] = ownedGamesResponse.Response.Games
			waitgroup.Done()
		}()
	}

	waitgroup.Wait()

	return ownedGames, nil
}
