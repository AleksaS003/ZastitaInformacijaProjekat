package windows

import (
	"fmt"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type LEAWindow struct {
	parent fyne.Window

	keySizeSelect *widget.Select
	keyFileEntry  *widget.Entry

	hexKeyLabel *widget.Label
	rawKeyLabel *widget.Label
	statusLabel *widget.Label
}

func NewLEAWindow(parent fyne.Window) *LEAWindow {
	return &LEAWindow{parent: parent}
}

func (l *LEAWindow) Build() *fyne.Container {
	l.createWidgets()
	return l.createLayout()
}

func (l *LEAWindow) createWidgets() {
	l.keySizeSelect = widget.NewSelect([]string{"128", "192", "256"}, func(s string) {})
	l.keySizeSelect.SetSelected("256")

	l.keyFileEntry = widget.NewEntry()
	l.keyFileEntry.SetPlaceHolder("lea.key")

	l.hexKeyLabel = widget.NewLabel("")
	l.hexKeyLabel.Wrapping = fyne.TextWrapBreak
	l.hexKeyLabel.TextStyle = fyne.TextStyle{Monospace: true}

	l.rawKeyLabel = widget.NewLabel("")
	l.rawKeyLabel.Wrapping = fyne.TextWrapBreak
	l.rawKeyLabel.TextStyle = fyne.TextStyle{Monospace: true}

	l.statusLabel = widget.NewLabel("")
}

func (l *LEAWindow) createLayout() *fyne.Container {
	btnSelectOutput := widget.NewButton("💾 Pregledaj", func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				l.keyFileEntry.SetText(writer.URI().Path())
			}
		}, l.parent).Show()
	})

	btnGenKey := widget.NewButton("🔑 GENERISI KLJUC", func() {
		l.generateKey()
	})
	btnGenKey.Importance = widget.HighImportance

	btnGenKeyFile := widget.NewButton("💾 GENERISI I SACUVAJ", func() {
		l.generateKeyFile()
	})

	btnCopyHex := widget.NewButton("📋 Kopiraj HEX", func() {
		l.parent.Clipboard().SetContent(l.hexKeyLabel.Text)
		dialog.ShowInformation("Uspeh", "HEX kljuc kopiran", l.parent)
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("🔑 LEA Key Management", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Velicina kljuca (biti):"),
			l.keySizeSelect,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Sacuvaj u fajl:"),
			l.keyFileEntry,
			btnSelectOutput,
		),
		container.NewHBox(btnGenKey, btnGenKeyFile),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Generisani kljuc:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewVBox(
			widget.NewLabel("HEX format:"),
			container.NewBorder(nil, nil, nil, btnCopyHex, l.hexKeyLabel),
			widget.NewLabel("Raw format:"),
			l.rawKeyLabel,
		),
		l.statusLabel,
	)

	return content
}

func (l *LEAWindow) generateKey() {
	l.statusLabel.SetText("🔄 Generisem kljuc...")

	go func() {
		cmd := exec.Command("./crypto-cli",
			"lea", "genkey", l.keySizeSelect.Selected,
		)

		output, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				l.statusLabel.SetText("❌ Greska pri generisanju")
				dialog.ShowError(fmt.Errorf(string(output)), l.parent)
				return
			}

			lines := strings.Split(string(output), "\n")

			var hexKey, rawKey string
			for i, line := range lines {
				if strings.Contains(line, "Hex:") {
					hexKey = strings.TrimSpace(strings.TrimPrefix(line, "Hex:"))
				}
				if i > 3 && line != "" && !strings.Contains(line, "Raw bytes") && !strings.Contains(line, "crypto-cli") {
					rawKey = line
				}
			}

			l.hexKeyLabel.SetText(hexKey)
			l.rawKeyLabel.SetText(rawKey)
			l.statusLabel.SetText("✅ Kljuc uspesno generisan")
		})
	}()
}

func (l *LEAWindow) generateKeyFile() {
	if l.keyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Unesite ime fajla"), l.parent)
		return
	}

	l.statusLabel.SetText("🔄 Generisem i cuvam kljuc...")

	go func() {
		cmd := exec.Command("./crypto-cli",
			"lea", "genkey-file",
			"--size", l.keySizeSelect.Selected,
			"--output", l.keyFileEntry.Text,
		)

		output, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				l.statusLabel.SetText("❌ Greska pri generisanju")
				dialog.ShowError(fmt.Errorf(string(output)), l.parent)
				return
			}

			l.statusLabel.SetText(fmt.Sprintf("✅ Kljuc sacuvan u %s", l.keyFileEntry.Text))
			dialog.ShowInformation("Uspeh", string(output), l.parent)
			l.generateKey()
		})
	}()
}
