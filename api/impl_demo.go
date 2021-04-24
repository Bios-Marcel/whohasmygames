//+build demo

package api

type demoSession struct {
}

func NewSession(apiToken string, targetAccountId SteamID) (Session, error) {
	return &demoSession{}, nil
}

func (session *demoSession) GetFriends(steamId SteamID) ([]*Friend, error) {
	return []*Friend{
		{SteamID: "123"},
		{SteamID: "234"},
		{SteamID: "345"},
	}, nil
}

func (session *demoSession) GetOwnFriends() ([]*Friend, error) {
	return session.GetFriends(SteamID("0"))
}

func (session *demoSession) GetFriendProfiles(friends []*Friend) (map[SteamID]*PlayerProfile, error) {
	return map[SteamID]*PlayerProfile{
		"123": {
			SteamID:     "123",
			Personaname: "Kevin (123)",
			AvatarURL:   "NOPE",
		},
		"234": {
			SteamID:     "234",
			Personaname: "John (234)",
			AvatarURL:   "NOPE",
		},
		"345": {
			SteamID:     "345",
			Personaname: "Amir (345)",
			AvatarURL:   "NOPE",
		},
	}, nil
}

func (session *demoSession) GetOwnedGames(friends []*Friend, forceRefresh bool) (map[SteamID][]*Game, error) {
	cslo := &Game{
		AppID: 1,
		Name:  "Counter Strike local offensive",
	}
	brawl := &Game{
		AppID: 2,
		Name:  "Brawlhalla",
	}

	games := make(map[SteamID][]*Game, len(friends))
	for _, fren := range friends {
		if fren.SteamID == "0" {
			games[fren.SteamID] = []*Game{cslo, brawl}
		}
		if fren.SteamID == "123" {
			games[fren.SteamID] = []*Game{cslo, brawl}
		}
		if fren.SteamID == "234" {
			games[fren.SteamID] = []*Game{brawl}
		}
		if fren.SteamID == "345" {
			games[fren.SteamID] = []*Game{cslo}
		}
	}
	return games, nil
}
