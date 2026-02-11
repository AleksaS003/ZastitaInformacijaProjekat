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

type FoursquareWindow struct {
	parent fyne.Window

	// Mode
	modeSelect *widget.RadioGroup

	// Input
	inputTypeSelect *widget.RadioGroup
	textEntry       *widget.Entry
	fileEntry       *widget.Entry
	btnSelectFile   *widget.Button

	// Keys
	key1Entry *widget.Entry
	key2Entry *widget.Entry

	// Output
	outputFileEntry *widget.Entry
	btnSelectOutput *widget.Button

	// Rezultati
	resultText  *widget.Entry
	statusLabel *widget.Label
}

func NewFoursquareWindow(parent fyne.Window) *FoursquareWindow {
	return &FoursquareWindow{parent: parent}
}

func (f *FoursquareWindow) Build() *fyne.Container {
	f.createWidgets()
	return f.createLayout()
}

func (f *FoursquareWindow) createWidgets() {
	// Mode: Encrypt/Decrypt
	f.modeSelect = widget.NewRadioGroup(
		[]string{"Enkripcija", "Dekripcija"},
		func(s string) { f.onModeChange() },
	)
	f.modeSelect.SetSelected("Enkripcija")
	f.modeSelect.Horizontal = true

	// Input type: Text/File
	f.inputTypeSelect = widget.NewRadioGroup(
		[]string{"Tekst", "Fajl"},
		func(s string) { f.onInputTypeChange() },
	)
	f.inputTypeSelect.SetSelected("Tekst")
	f.inputTypeSelect.Horizontal = true

	// Text input
	f.textEntry = widget.NewMultiLineEntry()
	f.textEntry.SetPlaceHolder("Unesite tekst za enkripciju/dekripciju...")
	f.textEntry.Wrapping = fyne.TextWrapWord

	// File input
	f.fileEntry = widget.NewEntry()
	f.fileEntry.SetPlaceHolder("Izaberi fajl...")

	f.btnSelectFile = widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				f.fileEntry.SetText(reader.URI().Path())
			}
		}, f.parent).Show()
	})

	// Ključevi
	f.key1Entry = widget.NewEntry()
	f.key1Entry.SetText("keyword")
	f.key1Entry.SetPlaceHolder("Prvi ključ...")

	f.key2Entry = widget.NewEntry()
	f.key2Entry.SetText("example")
	f.key2Entry.SetPlaceHolder("Drugi ključ...")

	// Output
	f.outputFileEntry = widget.NewEntry()
	f.outputFileEntry.SetPlaceHolder("Output fajl (opciono)...")

	f.btnSelectOutput = widget.NewButton("💾 Pregledaj", func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				f.outputFileEntry.SetText(writer.URI().Path())
			}
		}, f.parent).Show()
	})

	// Rezultat
	f.resultText = widget.NewMultiLineEntry()
	f.resultText.SetPlaceHolder("Rezultat će se prikazati ovde...")
	f.resultText.Wrapping = fyne.TextWrapWord
	f.resultText.Disable()

	// Status
	f.statusLabel = widget.NewLabel("")

	f.updateVisibility()
}

func (f *FoursquareWindow) createLayout() *fyne.Container {
	btnExecute := widget.NewButton("▶ IZVRŠI", func() {
		f.execute()
	})
	btnExecute.Importance = widget.HighImportance

	btnCopy := widget.NewButton("📋 Kopiraj", func() {
		if f.resultText.Text != "" {
			f.parent.Clipboard().SetContent(f.resultText.Text)
			dialog.ShowInformation("Uspeh", "Rezultat kopiran u clipboard", f.parent)
		}
	})

	btnClear := widget.NewButton("🧹 Obriši", func() {
		f.resultText.SetText("")
	})

	inputFileRow := container.NewGridWithColumns(2,
		f.fileEntry,
		f.btnSelectFile,
	)

	outputFileRow := container.NewGridWithColumns(3,
		widget.NewLabel("Output fajl:"),
		f.outputFileEntry,
		f.btnSelectOutput,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("🔣 Foursquare Cipher", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("Mod:"),
			f.modeSelect,
		),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("Input tip:"),
			f.inputTypeSelect,
		),
		f.textEntry,
		inputFileRow,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Ključ 1:"),
				f.key1Entry,
			),
			container.NewVBox(
				widget.NewLabel("Ključ 2:"),
				f.key2Entry,
			),
		),
		widget.NewSeparator(),
		outputFileRow,
		container.NewCenter(btnExecute),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Rezultat:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		f.resultText,
		container.NewHBox(btnCopy, btnClear),
		f.statusLabel,
	)

	return content
}

func (f *FoursquareWindow) onModeChange() {}

func (f *FoursquareWindow) onInputTypeChange() {
	f.updateVisibility()
}

func (f *FoursquareWindow) updateVisibility() {
	if f.textEntry == nil || f.fileEntry == nil || f.btnSelectFile == nil {
		return
	}

	isText := f.inputTypeSelect.Selected == "Tekst"
	f.textEntry.Hidden = !isText
	f.fileEntry.Hidden = isText
	f.btnSelectFile.Hidden = isText
}

func (f *FoursquareWindow) execute() {
	if f.inputTypeSelect.Selected == "Tekst" && f.textEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Unesite tekst"), f.parent)
		return
	}

	if f.inputTypeSelect.Selected == "Fajl" && f.fileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Izaberite fajl"), f.parent)
		return
	}

	mode := "encrypt"
	if f.modeSelect.Selected == "Dekripcija" {
		mode = "decrypt"
	}

	f.statusLabel.SetText("🔄 Obrada u toku...")

	go func() {
		var cmd *exec.Cmd

		if f.inputTypeSelect.Selected == "Tekst" {
			if f.outputFileEntry.Text != "" {
				cmd = exec.Command("./crypto-cli",
					"foursquare", mode,
					"--text", f.textEntry.Text,
					"--key1", f.key1Entry.Text,
					"--key2", f.key2Entry.Text,
					"--output", f.outputFileEntry.Text,
				)
			} else {
				cmd = exec.Command("./crypto-cli",
					"foursquare", mode,
					"--text", f.textEntry.Text,
					"--key1", f.key1Entry.Text,
					"--key2", f.key2Entry.Text,
				)
			}
		} else {
			args := []string{"foursquare", mode,
				"--file", f.fileEntry.Text,
				"--key1", f.key1Entry.Text,
				"--key2", f.key2Entry.Text,
			}

			if f.outputFileEntry.Text != "" {
				args = append(args, "--output", f.outputFileEntry.Text)
			}

			cmd = exec.Command("./crypto-cli", args...)
		}

		output, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				f.statusLabel.SetText("❌ Greška")
				dialog.ShowError(fmt.Errorf(string(output)), f.parent)
				return
			}

			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				f.resultText.SetText(strings.Join(lines[1:], "\n"))
			} else {
				f.resultText.SetText(string(output))
			}

			f.statusLabel.SetText("✅ Obrada uspešna!")
		})
	}()
}
