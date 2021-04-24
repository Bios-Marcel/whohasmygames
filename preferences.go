//+build !demo

package main

import "fyne.io/fyne/v2"

func getTargetAccountId(a fyne.App) string {
	return a.Preferences().String(steamAccountIdKey)
}

func getAPIToken(a fyne.App) string {
	return a.Preferences().String(steamApiTokenKey)
}
