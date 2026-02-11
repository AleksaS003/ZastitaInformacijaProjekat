package windows

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type DecryptWindow struct {
	parent fyne.Window

	// Polja
	inputFileEntry  *widget.Entry
	keyFileEntry    *widget.Entry
	outputFileEntry *widget.Entry

	// Status
	statusLabel   *widget.Label
	resultLabel   *widget.Label
	metadataLabel *widget.Label
}

func NewDecryptWindow(parent fyne.Window) *DecryptWindow {
	return &DecryptWindow{parent: parent}
}

func (d *DecryptWindow) Build() *fyne.Container {
	d.createWidgets()
	return d.createLayout()
}

func (d *DecryptWindow) createWidgets() {
	// Input fajl (.enc)
	d.inputFileEntry = widget.NewEntry()
	d.inputFileEntry.SetPlaceHolder("Izaberi .enc fajl za dekripciju...")

	// Key fajl
	d.keyFileEntry = widget.NewEntry()
	d.keyFileEntry.SetPlaceHolder("Izaberi key fajl...")

	// Output fajl
	d.outputFileEntry = widget.NewEntry()
	d.outputFileEntry.SetPlaceHolder("Output fajl (opciono)...")

	// Status i rezultati
	d.statusLabel = widget.NewLabel("")
	d.resultLabel = widget.NewLabel("")
	d.metadataLabel = widget.NewLabel("")
}

func (d *DecryptWindow) createLayout() *fyne.Container {
	btnSelectInput := widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := reader.URI().Path()
				d.inputFileEntry.SetText(path)

				if strings.HasSuffix(path, ".enc") {
					d.outputFileEntry.SetText(path[:len(path)-4])
				}
			}
		}, d.parent).Show()
	})

	btnSelectKey := widget.NewButton("🔑 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				d.keyFileEntry.SetText(reader.URI().Path())
			}
		}, d.parent).Show()
	})

	btnSelectOutput := widget.NewButton("💾 Pregledaj", func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				d.outputFileEntry.SetText(writer.URI().Path())
			}
		}, d.parent).Show()
	})

	btnDecrypt := widget.NewButton("🔓 DEKRIPTUJ", func() {
		d.runDecryption()
	})
	btnDecrypt.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabelWithStyle("🔓 Dekripcija fajla", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			widget.NewLabel("Enkriptovani fajl:"),
			d.inputFileEntry,
			btnSelectInput,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Key fajl:"),
			d.keyFileEntry,
			btnSelectKey,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Output fajl:"),
			d.outputFileEntry,
			btnSelectOutput,
		),
		container.NewCenter(btnDecrypt),
		widget.NewSeparator(),
		container.NewVBox(
			d.statusLabel,
			d.resultLabel,
			d.metadataLabel,
		),
	)

	return container.NewVBox(form)
}

func (d *DecryptWindow) runDecryption() {
	if d.inputFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati enkriptovani fajl"), d.parent)
		return
	}

	if d.keyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), d.parent)
		return
	}

	d.statusLabel.SetText("🔄 Dekripcija u toku...")

	go func() {
		output := d.outputFileEntry.Text
		if output == "" {
			output = d.inputFileEntry.Text
			if strings.HasSuffix(output, ".enc") {
				output = output[:len(output)-4]
			} else {
				output = output + ".dec"
			}
		}

		cmd := exec.Command("./crypto-cli",
			"decrypt-file",
			"--file", d.inputFileEntry.Text,
			"--keyfile", d.keyFileEntry.Text,
			"--output", output,
		)

		_, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				d.statusLabel.SetText("❌ Greška pri dekripciji")
				dialog.ShowError(err, d.parent)
				return
			}

			d.statusLabel.SetText("✅ Dekripcija uspešna!")
			d.resultLabel.SetText(fmt.Sprintf("Output: %s", output))
			d.metadataLabel.SetText("✓ Hash verifikovan\n✓ Originalni fajl: " + filepath.Base(d.inputFileEntry.Text))
		})
	}()
}
