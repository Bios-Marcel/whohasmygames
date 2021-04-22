package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const steamApiTokenKey = "steam-api-token"
const steamAccountIdKey = "steam-account-id"

func main() {
	a := app.NewWithID("me.marcelschr.whohasmygames")

	apiToken := a.Preferences().String(steamApiTokenKey)
	accountId := a.Preferences().String(steamAccountIdKey)

	window := a.NewWindow("Who Has My Games?")
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(800, 600))

	if apiToken == "" || accountId == "" {
		loginContainer := createLoginContainer(a, func() {
			window.SetContent(initAndSetMainScreen(window, a))
		})
		window.SetContent(loginContainer)
	} else {
		window.SetContent(initAndSetMainScreen(window, a))
	}

	window.ShowAndRun()
}

func initAndSetMainScreen(window fyne.Window, a fyne.App) fyne.Widget {
	session, err := NewSession(a.Preferences().String(steamApiTokenKey), steamID(a.Preferences().String(steamAccountIdKey)))
	if err != nil {
		panic(err)
	}

	sourceFriends, friendsErr := session.getOwnFriends()
	if friendsErr != nil {
		panic(friendsErr)
	}
	log.Printf("Friends: %d", len(sourceFriends))

	profiles, profileError := session.getFriendProfiles(sourceFriends)
	if profileError != nil {
		panic(profileError)
	}
	log.Printf("Profiles: %d", len(profiles))

	// sourceFriends := []*friend{{SteamID: 123}, {SteamID: 345}, {SteamID: 678}, {SteamID: 910}}
	// profiles := map[steamID]*playerProfile{
	// 	123: {
	// 		SteamID:     123,
	// 		Personaname: "Kevin (123)",
	// 		AvatarURL:   "NOPE",
	// 	},
	// 	345: {
	// 		SteamID:     345,
	// 		Personaname: "Kevin (345)",
	// 		AvatarURL:   "NOPE",
	// 	},
	// 	678: {
	// 		SteamID:     678,
	// 		Personaname: "Kevin (678)",
	// 		AvatarURL:   "NOPE",
	// 	},
	// 	910: {
	// 		SteamID:     910,
	// 		Personaname: "Kevin (910)",
	// 		AvatarURL:   "NOPE",
	// 	},
	// }

	var targetFriends []*friend
	var targetFriendsList *widget.List

	sourceFriendsList := widget.NewList(
		func() int {
			return len(sourceFriends)
		},

		func() fyne.CanvasObject {
			return widget.NewButton("", func() {})
		},

		func(id widget.ListItemID, obj fyne.CanvasObject) {
			button := obj.(*widget.Button)
			button.OnTapped = func() {
				//Avoiding adding the same friend twice.
				for _, friend := range targetFriends {
					if friend == sourceFriends[id] {
						return
					}
				}

				targetFriends = append(targetFriends, sourceFriends[id])
				targetFriendsList.Refresh()
			}
			button.SetText(profiles[sourceFriends[id].SteamID].Personaname)
		})

	targetFriendsList = widget.NewList(
		func() int {
			return len(targetFriends)
		},

		func() fyne.CanvasObject {
			return widget.NewButton("", func() {})
		},

		func(id widget.ListItemID, obj fyne.CanvasObject) {
			button := obj.(*widget.Button)
			button.OnTapped = func() {
				targetIndex := -1
				for index, friend := range targetFriends {
					if friend == targetFriends[id] {
						targetIndex = index
						break
					}

				}

				if targetIndex != -1 {
					targetFriends = append(targetFriends[:targetIndex], targetFriends[targetIndex+1:]...)
					targetFriendsList.Refresh()
				}
			}
			button.SetText(profiles[targetFriends[id].SteamID].Personaname)
		})

	gamesYouAllOwnText := widget.NewLabel("")
	confirmButton := widget.NewButton("Tell me", func() {
		gamesYouAllOwnText.Text = ""

		selfAsFriend := &friend{SteamID: steamID(a.Preferences().String(steamAccountIdKey))}
		ownedGames, err := session.getOwnedGames(append(targetFriends, selfAsFriend))
		if err != nil {
			panic(err)
		}

		ownOwnedGames := ownedGames[selfAsFriend.SteamID]

	GAME_LOOP:
		for _, ownOwnedGame := range ownOwnedGames {

		FREN_LOOP:
			for friendSteamId, friendsOwnedGames := range ownedGames {
				//Avoid checking self
				if selfAsFriend.SteamID == friendSteamId {
					continue
				}

				for _, friendsOwnedGame := range friendsOwnedGames {
					//Since this friend has the game we are checking right now, we are done with this friend.
					if ownOwnedGame.AppID == friendsOwnedGame.AppID {
						continue FREN_LOOP
					}
				}

				//If we reach this point, one of our friends doesn't have the game, meaning we are done with that game.
				continue GAME_LOOP
			}

			//We all got the game!
			gamesYouAllOwnText.Text = gamesYouAllOwnText.Text + "\n" + ownOwnedGame.Name
		}

		gamesYouAllOwnText.Refresh()
	})

	return container.NewVSplit(
		container.NewHSplit(
			sourceFriendsList,
			targetFriendsList,
		),
		container.NewVSplit(
			confirmButton,
			container.NewVScroll(
				gamesYouAllOwnText,
			),
		),
	)
}

func createLoginContainer(app fyne.App, afterSave func()) *fyne.Container {
	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.Text = app.Preferences().String(steamApiTokenKey)
	accountIdEntry := widget.NewEntry()
	accountIdEntry.Text = app.Preferences().String(steamAccountIdKey)
	return container.NewBorder(
		nil,
		widget.NewButton("Save", func() {
			targetAccountId, err := getSteamID(apiKeyEntry.Text, accountIdEntry.Text)
			if err != nil {
				//FIXME Better error handling
				panic(err)
			}

			app.Preferences().SetString(steamApiTokenKey, apiKeyEntry.Text)
			app.Preferences().SetString(steamAccountIdKey, string(targetAccountId))

			if afterSave != nil {
				afterSave()
			}
		}),
		nil,
		nil,
		container.NewGridWithColumns(2,
			widget.NewLabel("API Key"), apiKeyEntry,
			widget.NewLabel("Account ID/Name"), accountIdEntry,
		),
	)
}
