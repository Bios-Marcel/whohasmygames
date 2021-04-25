package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Bios-Marcel/whohasmygames/api"
	"github.com/Bios-Marcel/whohasmygames/stringutility"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const steamApiTokenKey = "steam-api-token"
const steamAccountIdKey = "steam-account-id"

func main() {
	a := app.NewWithID("me.marcelschr.whohasmygames")

	apiToken := getAPIToken(a)
	accountId := getTargetAccountId(a)

	window := a.NewWindow("Who Has My Games?")
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(800, 600))

	if apiToken == "" || accountId == "" {
		loginContainer := createLoginContainer(a, func() {
			transitionToMainScreen(a, window)
		})

		window.SetContent(loginContainer)
	} else {
		transitionToMainScreen(a, window)
	}

	window.ShowAndRun()
}

func transitionToMainScreen(a fyne.App, window fyne.Window) {
	mainScreen := initAndSetMainScreen(window, a)
	mainTab := container.NewTabItem("Main", mainScreen.container)
	settingsTab := container.NewTabItem("Settings", createLoginContainer(a, func() {
		readyMainScreen(a, mainScreen)
		mainScreen.container.Refresh()
	}))
	tabPane := container.NewAppTabs(mainTab, settingsTab)
	window.SetContent(tabPane)
}

type mainScreen struct {
	container *fyne.Container
	session   api.Session

	profiles              map[api.SteamID]*api.PlayerProfile
	sourceFriends         []*api.Friend
	filteredSourceFriends []*api.Friend
	targetFriends         []*api.Friend
}

func readyMainScreen(a fyne.App, mainScreen *mainScreen) {
	session, err := api.NewSession(getAPIToken(a), api.SteamID(getTargetAccountId(a)))
	if err != nil {
		panic(err)
	}

	sourceFriends, friendsErr := session.GetOwnFriends()
	if friendsErr != nil {
		panic(friendsErr)
	}

	profiles, profileError := session.GetFriendProfiles(sourceFriends)
	if profileError != nil {
		panic(profileError)
	}

	mainScreen.profiles = profiles
	mainScreen.sourceFriends = sourceFriends
	mainScreen.filteredSourceFriends = make([]*api.Friend, len(sourceFriends))
	copy(mainScreen.filteredSourceFriends, mainScreen.sourceFriends)
	mainScreen.targetFriends = nil
	mainScreen.session = session
}

func initAndSetMainScreen(window fyne.Window, a fyne.App) *mainScreen {
	mainScreen := &mainScreen{}
	readyMainScreen(a, mainScreen)

	var targetFriendsList *widget.List

	sourceFriendsList := widget.NewList(
		func() int {
			return len(mainScreen.filteredSourceFriends)
		},

		func() fyne.CanvasObject {
			return widget.NewButton("", func() {})
		},

		func(id widget.ListItemID, obj fyne.CanvasObject) {
			button := obj.(*widget.Button)
			button.OnTapped = func() {
				//Avoiding adding the same friend twice.
				for _, friend := range mainScreen.targetFriends {
					if friend == mainScreen.filteredSourceFriends[id] {
						return
					}
				}

				mainScreen.targetFriends = append(mainScreen.targetFriends, mainScreen.filteredSourceFriends[id])
				targetFriendsList.Refresh()
			}
			button.SetText(mainScreen.profiles[mainScreen.filteredSourceFriends[id].SteamID].Personaname)
		})

	targetFriendsList = widget.NewList(
		func() int {
			return len(mainScreen.targetFriends)
		},

		func() fyne.CanvasObject {
			return widget.NewButton("", func() {})
		},

		func(id widget.ListItemID, obj fyne.CanvasObject) {
			button := obj.(*widget.Button)
			button.OnTapped = func() {
				targetIndex := -1
				for index, friend := range mainScreen.targetFriends {
					if friend == mainScreen.targetFriends[id] {
						targetIndex = index
						break
					}

				}

				if targetIndex != -1 {
					mainScreen.targetFriends = append(mainScreen.targetFriends[:targetIndex], mainScreen.targetFriends[targetIndex+1:]...)
					targetFriendsList.Refresh()
				}
			}
			button.SetText(mainScreen.profiles[mainScreen.targetFriends[id].SteamID].Personaname)
		})

	gamesYouAllOwnText := widget.NewMultiLineEntry()

	confirmButton := widget.NewButton("Tell me", func() {
		stringJoiner := stringutility.NewStringJoiner('\n')
		//Makes sure to update the text after we done. Whether the text is
		//empty (failure or no common games) or we actually found games.
		defer func() {
			gamesYouAllOwnText.SetText(stringJoiner.String())
			gamesYouAllOwnText.Refresh()
		}()

		selfAsFriend := &api.Friend{SteamID: api.SteamID(getTargetAccountId(a))}
		ownedGames, err := mainScreen.session.GetOwnedGames(append(mainScreen.targetFriends, selfAsFriend), false)
		if err != nil {
			panic(err)
		}

		ownOwnedGames := ownedGames[selfAsFriend.SteamID]

	GAME_LOOP:
		for _, ownOwnedGame := range ownOwnedGames {

		FREN_LOOP:
			for friendSteamId, friendsOwnedGames := range ownedGames {
				//Avoid checking self, the result is obvious ;)
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
			stringJoiner.WriteString(ownOwnedGame.Name)
		}
	})

	sourceFriendsSearchField := widget.NewEntry()
	sourceFriendsSearchField.PlaceHolder = "Type to search"
	lowercaser := cases.Lower(language.English)
	sourceFriendsSearchField.OnChanged = func(filterValue string) {
		lowercased := lowercaser.String(filterValue)
		mainScreen.filteredSourceFriends = mainScreen.filteredSourceFriends[:cap(mainScreen.filteredSourceFriends)]
		if lowercased == "" {
			for index, fren := range mainScreen.sourceFriends {
				mainScreen.filteredSourceFriends[index] = fren
			}
		} else {
			var index int
			for _, fren := range mainScreen.sourceFriends {
				profile, avail := mainScreen.profiles[fren.SteamID]
				if avail && strings.Contains(lowercaser.String(profile.Personaname), lowercased) {
					mainScreen.filteredSourceFriends[index] = fren
					index++
				}
			}
			mainScreen.filteredSourceFriends = mainScreen.filteredSourceFriends[:index]
		}

		sourceFriendsList.Refresh()
	}

	clearTargetFriendsButton := widget.NewButton("Clear", func() {
		mainScreen.targetFriends = nil
		targetFriendsList.Refresh()
	})

	sourceFriendsWithSearch := container.NewBorder(sourceFriendsSearchField, nil, nil, nil, sourceFriendsList)

	targetFriendsListContainerLayout := NewPrioVBoxLayout()
	targetFriendsListContainerLayout.SetGrow(targetFriendsList, true)

	targetFriendsListContainer := container.New(targetFriendsListContainerLayout,
		container.NewHBox(clearTargetFriendsButton),
		widget.NewSeparator(),
		targetFriendsList)

	friendsSplitter := container.NewHSplit(
		sourceFriendsWithSearch,
		targetFriendsListContainer,
	)
	resultScrollText := container.NewVScroll(
		gamesYouAllOwnText,
	)

	copyResultTextButton := widget.NewButton("Copy Result", func() {
		window.Clipboard().SetContent(gamesYouAllOwnText.Text)
	})

	layout := NewPrioVBoxLayout()
	layout.SetGrow(friendsSplitter, true)
	layout.SetGrow(resultScrollText, true)

	mainScreen.container = container.New(layout, friendsSplitter, confirmButton, resultScrollText, copyResultTextButton)

	return mainScreen
}

func createLoginContainer(app fyne.App, afterSave func()) *fyne.Container {
	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.Text = app.Preferences().String(steamApiTokenKey)
	accountIdEntry := widget.NewEntry()
	accountIdEntry.Text = app.Preferences().String(steamAccountIdKey)
	return container.NewBorder(
		nil,
		widget.NewButton("Save", func() {
			targetAccountId, err := api.GetSteamID(apiKeyEntry.Text, accountIdEntry.Text)
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
