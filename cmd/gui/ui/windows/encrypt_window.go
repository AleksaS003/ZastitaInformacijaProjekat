package windows

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type EncryptWindow struct {
	parent fyne.Window

	inputFileEntry  *widget.Entry
	keyFileEntry    *widget.Entry
	outputFileEntry *widget.Entry
	algorithmSelect *widget.Select
	statusLabel     *widget.Label

	resultLabel   *widget.Label
	metadataLabel *widget.Label
}

func NewEncryptWindow(parent fyne.Window) *EncryptWindow {
	return &EncryptWindow{parent: parent}
}

func (e *EncryptWindow) Build() *fyne.Container {

	e.inputFileEntry = widget.NewEntry()
	e.inputFileEntry.SetPlaceHolder("Izaberi fajl za enkripciju...")

	btnSelectInput := widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				e.inputFileEntry.SetText(reader.URI().Path())
			}
		}, e.parent).Show()
	})

	e.keyFileEntry = widget.NewEntry()
	e.keyFileEntry.SetPlaceHolder("Izaberi key fajl...")

	btnSelectKey := widget.NewButton("🔑 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				e.keyFileEntry.SetText(reader.URI().Path())
			}
		}, e.parent).Show()
	})

	e.outputFileEntry = widget.NewEntry()
	e.outputFileEntry.SetPlaceHolder("Output fajl (opciono)...")

	btnSelectOutput := widget.NewButton("💾 Pregledaj", func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				e.outputFileEntry.SetText(writer.URI().Path())
			}
		}, e.parent).Show()
	})

	e.algorithmSelect = widget.NewSelect([]string{"LEA", "LEA-PCBC"}, func(s string) {})
	e.algorithmSelect.SetSelected("LEA-PCBC")

	btnEncrypt := widget.NewButton("🔒 ENKRIPTUJ", func() {
		e.runEncryption()
	})
	btnEncrypt.Importance = widget.HighImportance

	e.statusLabel = widget.NewLabel("")
	e.resultLabel = widget.NewLabel("")
	e.metadataLabel = widget.NewLabel("")

	form := container.NewVBox(
		widget.NewLabelWithStyle("🔐 Enkripcija fajla", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			widget.NewLabel("Input fajl:"),
			e.inputFileEntry,
			btnSelectInput,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Key fajl:"),
			e.keyFileEntry,
			btnSelectKey,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Output fajl:"),
			e.outputFileEntry,
			btnSelectOutput,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Algoritam:"),
			e.algorithmSelect,
		),
		container.NewCenter(btnEncrypt),
		widget.NewSeparator(),
		container.NewVBox(
			e.statusLabel,
			e.resultLabel,
			e.metadataLabel,
		),
	)

	return container.NewVBox(form)
}

func (e *EncryptWindow) runEncryption() {
	if e.inputFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati fajl"), e.parent)
		return
	}

	if e.keyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), e.parent)
		return
	}

	e.statusLabel.SetText("🔄 Enkripcija u toku...")

	go func() {
		output := e.outputFileEntry.Text
		if output == "" {
			output = e.inputFileEntry.Text + ".enc"
		}

		cmd := exec.Command("./crypto-cli",
			"encrypt-file",
			"--file", e.inputFileEntry.Text,
			"--keyfile", e.keyFileEntry.Text,
			"--algo", e.algorithmSelect.Selected,
			"--output", output,
		)

		_, err := cmd.CombinedOutput()
		if err != nil {
			e.statusLabel.SetText("❌ Greška pri enkripciji")
			dialog.ShowError(err, e.parent)
			return
		}

		e.statusLabel.SetText("✅ Enkripcija uspešna!")
		e.resultLabel.SetText(fmt.Sprintf("Output: %s", output))
		e.metadataLabel.SetText("Original filename: " + filepath.Base(e.inputFileEntry.Text))
	}()
}
