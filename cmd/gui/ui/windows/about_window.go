package windows

import (
	"fmt"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AboutWindow struct {
	parent fyne.Window
}

func NewAboutWindow(parent fyne.Window) *AboutWindow {
	return &AboutWindow{parent: parent}
}

func (a *AboutWindow) Build() *fyne.Container {
	
	title := canvas.NewText("🔐 Crypto App", theme.PrimaryColor())
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Sigurna kriptografska alatka", theme.ForegroundColor())
	subtitle.TextSize = 14

	
	version := widget.NewLabelWithStyle(
		fmt.Sprintf("Verzija: 1.0.0 (%s/%s)", runtime.GOOS, runtime.GOARCH),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	
	description := widget.NewLabel(`Crypto App je sveobuhvatna kriptografska alatka razvijena u Go programskom jeziku. 
	
Podržani algoritmi:
• LEA (Lightweight Encryption Algorithm) - 128/192/256 bit
• LEA-PCBC (Propagating Cipher Block Chaining) mod
• Foursquare Cipher - klasična šifra
• SHA-256 - sigurna hash funkcija

Funkcionalnosti:
• Enkripcija/dekripcija fajlova sa metapodacima
• File System Watcher - automatska enkripcija
• TCP Server/Client za bezbedan prenos fajlova
• Generisanje i upravljanje ključevima
• Detaljno logovanje svih aktivnosti
• Cross-platform GUI (Fyne)

Sve operacije se detaljno loguju za potrebe bezbednosnog audita i monitoringa.`)
	description.Wrapping = fyne.TextWrapWord

	
	infoGrid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Autor:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Aleksa S."),

		widget.NewLabelWithStyle("Licenca:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("MIT"),

		widget.NewLabelWithStyle("Go verzija:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(runtime.Version()),

		widget.NewLabelWithStyle("GUI Framework:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Fyne v2"),
	)

	
	logInfo := widget.NewLabelWithStyle(
		"Logovi se čuvaju u ./logs/ direktorijumu",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	
	links := container.NewHBox(
		widget.NewHyperlink("GitHub", parseURL("https:
		widget.NewLabel("•"),
		widget.NewHyperlink("Dokumentacija", parseURL("https:
	)

	
	btnClose := widget.NewButton("Zatvori", func() {
		
	})

	
	content := container.NewVBox(
		container.NewCenter(
			container.NewVBox(
				title,
				subtitle,
				widget.NewSeparator(),
			),
		),
		version,
		widget.NewSeparator(),
		description,
		widget.NewSeparator(),
		infoGrid,
		widget.NewSeparator(),
		logInfo,
		container.NewCenter(links),
		container.NewCenter(btnClose),
	)

	return container.NewVBox(content)
}

func parseURL(urlStr string) *url.URL {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return url
}
