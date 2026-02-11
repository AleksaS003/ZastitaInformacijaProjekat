package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/AleksaS003/zastitaprojekat/cmd/gui/ui"
)

func main() {

	myApp := app.NewWithID("com.cryptoapp.gui")

	myApp.Settings().SetTheme(theme.DarkTheme())

	mainWindow := myApp.NewWindow("Crypto App - GUI")
	mainWindow.Resize(fyne.NewSize(900, 700))
	mainWindow.CenterOnScreen()
	mainWindow.SetMaster()

	mainUI := ui.NewMainUI(mainWindow)

	mainWindow.SetContent(mainUI.Build())
	mainWindow.SetMainMenu(mainUI.CreateMenu())

	mainWindow.ShowAndRun()
}
