package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Bios-Marcel/whohasmygames/httputility"
)

type Session interface {
	GetFriends(steamId SteamID) ([]*Friend, error)
	GetOwnFriends() ([]*Friend, error)
	GetFriendProfiles(friends []*Friend) (map[SteamID]*PlayerProfile, error)
	GetOwnedGames(friends []*Friend, forceRefresh bool) (map[SteamID][]*Game, error)
}

// SteamID represents the "database" ID from steam. While this is clearly
// a number, steam decides to be retarded and send strings anways.
type SteamID string

// Game represents a game, but only contains the appID and the anime.
// While more info is available, it's not part of this struct.
type Game struct {
	AppID uint64 `json:"appid"`
	Name  string `json:"name"`
}

// PlayerProfile represents a steam profile. It only contains enough
// information to render a player (image+name). While more info is available
// it's not part of this struct.
type PlayerProfile struct {
	SteamID     SteamID `json:"steamid"`
	Personaname string  `json:"personaname"`
	AvatarURL   string  `json:"avatar"`
}

// Friend only holds a players SteamID. More info can be found in the
// corresponding PlayerProfile object.
type Friend struct {
	SteamID SteamID `json:"steamid"`
}

//FIXME Move into impl.go

type steamIDResponse struct {
	SteamID SteamID `json:"steamid"`
}

func GetSteamID(apiKey string, steamIdOrVanityName string) (SteamID, error) {
	_, parseError := strconv.ParseUint(steamIdOrVanityName, 10, 64)
	if parseError == nil {
		return SteamID(steamIdOrVanityName), nil
	}

	steamIdRequest, err := httputility.MakeGetRequest(fmt.Sprintf("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v1?key=%s&vanityurl=%s", apiKey, steamIdOrVanityName))
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
