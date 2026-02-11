package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/AleksaS003/zastitaprojekat/cmd/gui/ui/windows"
)

type MainUI struct {
	window fyne.Window
	tabs   *container.AppTabs

	// Prozori
	encryptWin    *windows.EncryptWindow
	decryptWin    *windows.DecryptWindow
	foursquareWin *windows.FoursquareWindow
	leaWin        *windows.LEAWindow
	fswWin        *windows.FSWWindow
	networkWin    *windows.NetworkWindow
	hashWin       *windows.HashWindow
	logsWin       *windows.LogsWindow
	aboutWin      *windows.AboutWindow
}

func NewMainUI(window fyne.Window) *MainUI {
	return &MainUI{
		window:        window,
		encryptWin:    windows.NewEncryptWindow(window),
		decryptWin:    windows.NewDecryptWindow(window),
		foursquareWin: windows.NewFoursquareWindow(window),
		leaWin:        windows.NewLEAWindow(window),
		fswWin:        windows.NewFSWWindow(window),
		networkWin:    windows.NewNetworkWindow(window),
		hashWin:       windows.NewHashWindow(window),
		logsWin:       windows.NewLogsWindow(window),
		aboutWin:      windows.NewAboutWindow(window),
	}
}

func (m *MainUI) Build() *fyne.Container {
	m.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Enkripcija", theme.ConfirmIcon(), m.encryptWin.Build()),
		container.NewTabItemWithIcon("Dekripcija", theme.CancelIcon(), m.decryptWin.Build()),
		container.NewTabItemWithIcon("Foursquare", theme.ViewRefreshIcon(), m.foursquareWin.Build()),
		container.NewTabItemWithIcon("LEA Ključevi", theme.InfoIcon(), m.leaWin.Build()),
		container.NewTabItemWithIcon("FSW Watcher", theme.VisibilityIcon(), m.fswWin.Build()),
		container.NewTabItemWithIcon("TCP Server/Client", theme.ComputerIcon(), m.networkWin.Build()),
		container.NewTabItemWithIcon("SHA-256 Hash", theme.DocumentIcon(), m.hashWin.Build()),
		container.NewTabItemWithIcon("Logovi", theme.HistoryIcon(), m.logsWin.Build()),
		container.NewTabItemWithIcon("O programu", theme.HelpIcon(), m.aboutWin.Build()),
	)

	m.tabs.SetTabLocation(container.TabLocationLeading)

	statusBar := m.createStatusBar()

	return container.NewBorder(nil, statusBar, nil, nil, m.tabs)
}

func (m *MainUI) createStatusBar() *fyne.Container {
	status := widget.NewLabelWithStyle(
		"Status: Spremno",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	return container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		status,
	)
}

func (m *MainUI) CreateMenu() *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Izlaz", func() { m.window.Close() }),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Dokumentacija", func() {
			// TODO: Otvori dokumentaciju
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("O programu", func() {
			m.tabs.SelectIndex(8)
		}),
	)

	return fyne.NewMainMenu(fileMenu, helpMenu)
}
