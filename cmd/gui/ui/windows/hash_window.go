package windows

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type HashWindow struct {
	parent fyne.Window

	modeSelect *widget.RadioGroup

	inputTypeSelect *widget.RadioGroup
	textEntry       *widget.Entry
	fileEntry       *widget.Entry
	btnSelectFile   *widget.Button

	hashEntry     *widget.Entry
	saveFileEntry *widget.Entry
	btnSelectSave *widget.Button

	verifyFileEntry     *widget.Entry
	btnSelectVerifyFile *widget.Button
	verifyHashEntry     *widget.Entry
	btnSelectHashFile   *widget.Button
	verifyResult        *widget.Label

	// Status
	statusLabel *widget.Label

	// Sekcije
	hashSection   *fyne.Container
	verifySection *fyne.Container
}

func NewHashWindow(parent fyne.Window) *HashWindow {
	return &HashWindow{parent: parent}
}

func (h *HashWindow) Build() *fyne.Container {
	h.createWidgets()
	return h.createLayout()
}

func (h *HashWindow) createWidgets() {
	// Mode: Hash/Verify
	h.modeSelect = widget.NewRadioGroup(
		[]string{"Izračunaj hash", "Verifikuj hash"},
		func(s string) { h.onModeChange() },
	)
	h.modeSelect.SetSelected("Izračunaj hash")
	h.modeSelect.Horizontal = true

	// ----- HASH SEKCIJA -----
	h.inputTypeSelect = widget.NewRadioGroup(
		[]string{"Tekst", "Fajl"},
		func(s string) { h.onInputTypeChange() },
	)
	h.inputTypeSelect.SetSelected("Tekst")
	h.inputTypeSelect.Horizontal = true

	h.textEntry = widget.NewMultiLineEntry()
	h.textEntry.SetPlaceHolder("Unesite tekst za hashovanje...")
	h.textEntry.Wrapping = fyne.TextWrapWord

	h.fileEntry = widget.NewEntry()
	h.fileEntry.SetPlaceHolder("Izaberi fajl za hashovanje...")

	h.btnSelectFile = widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				h.fileEntry.SetText(reader.URI().Path())
				h.calculateHash()
			}
		}, h.parent).Show()
	})

	h.hashEntry = widget.NewEntry()
	h.hashEntry.SetPlaceHolder("SHA-256 hash će se prikazati ovde...")
	h.hashEntry.Disable()
	h.hashEntry.TextStyle = fyne.TextStyle{Monospace: true}

	h.saveFileEntry = widget.NewEntry()
	h.saveFileEntry.SetPlaceHolder("Sacuvaj hash u fajl...")

	h.btnSelectSave = widget.NewButton("💾 Pregledaj", func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				h.saveFileEntry.SetText(writer.URI().Path())
			}
		}, h.parent).Show()
	})

	// ----- VERIFICATION SEKCIJA -----
	h.verifyFileEntry = widget.NewEntry()
	h.verifyFileEntry.SetPlaceHolder("Izaberi fajl za verifikaciju...")

	h.btnSelectVerifyFile = widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				h.verifyFileEntry.SetText(reader.URI().Path())
			}
		}, h.parent).Show()
	})

	h.verifyHashEntry = widget.NewEntry()
	h.verifyHashEntry.SetPlaceHolder("Unesite ocekivani hash...")
	h.verifyHashEntry.TextStyle = fyne.TextStyle{Monospace: true}

	h.btnSelectHashFile = widget.NewButton("📂 Ucitaj hash", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				go func() {
					content, err := os.ReadFile(reader.URI().Path())
					if err == nil {
						h.verifyHashEntry.SetText(strings.TrimSpace(string(content)))
					}
				}()
			}
		}, h.parent).Show()
	})

	h.verifyResult = widget.NewLabel("")
	h.verifyResult.TextStyle = fyne.TextStyle{Bold: true}

	h.statusLabel = widget.NewLabel("")

	h.updateVisibility()
}

func (h *HashWindow) createLayout() *fyne.Container {
	btnCalculate := widget.NewButton("🔢 IZRACUNAJ HASH", func() {
		h.calculateHash()
	})
	btnCalculate.Importance = widget.HighImportance

	btnSaveHash := widget.NewButton("💾 Sacuvaj hash", func() {
		h.saveHash()
	})

	btnVerify := widget.NewButton("✓ VERIFIKUJ", func() {
		h.verifyHash()
	})
	btnVerify.Importance = widget.HighImportance

	h.hashSection = container.NewVBox(
		widget.NewLabelWithStyle("🔢 SHA-256 Hash", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("Input tip:"),
			h.inputTypeSelect,
		),
		h.textEntry,
		container.NewGridWithColumns(2,
			h.fileEntry,
			h.btnSelectFile,
		),
		btnCalculate,
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabelWithStyle("Rezultat:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			h.hashEntry,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Sacuvaj u fajl:"),
			h.saveFileEntry,
			container.NewHBox(h.btnSelectSave, btnSaveHash),
		),
	)

	h.verifySection = container.NewVBox(
		widget.NewLabelWithStyle("✓ Verifikacija hash-a", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Fajl za proveru:"),
			container.NewHBox(h.verifyFileEntry, h.btnSelectVerifyFile),
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Očekivani hash:"),
			h.verifyHashEntry,
			h.btnSelectHashFile,
		),
		btnVerify,
		h.verifyResult,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("🔐 SHA-256 Alat", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		h.modeSelect,
		widget.NewSeparator(),
		container.NewStack(
			h.hashSection,
			h.verifySection,
		),
		h.statusLabel,
	)

	h.updateVisibility()
	return content
}

func (h *HashWindow) onModeChange() {
	h.updateVisibility()
}

func (h *HashWindow) onInputTypeChange() {
	h.updateVisibility()
}

func (h *HashWindow) updateVisibility() {
	if h.hashSection == nil || h.verifySection == nil {
		return
	}

	isHashMode := h.modeSelect.Selected == "Izracunaj hash"
	h.hashSection.Hidden = !isHashMode
	h.verifySection.Hidden = isHashMode

	if isHashMode {
		if h.textEntry == nil || h.fileEntry == nil || h.btnSelectFile == nil {
			return
		}

		isText := h.inputTypeSelect.Selected == "Tekst"
		h.textEntry.Hidden = !isText
		h.fileEntry.Hidden = isText
		h.btnSelectFile.Hidden = isText
	}
}

func (h *HashWindow) calculateHash() {
	if h.inputTypeSelect.Selected == "Tekst" && h.textEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Unesite tekst"), h.parent)
		return
	}

	if h.inputTypeSelect.Selected == "Fajl" && h.fileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Izaberite fajl"), h.parent)
		return
	}

	h.statusLabel.SetText("🔄 Racunam hash...")

	go func() {
		var cmd *exec.Cmd

		if h.inputTypeSelect.Selected == "Tekst" {
			cmd = exec.Command("./crypto-cli",
				"sha256", "hash",
				"--text", h.textEntry.Text,
			)
		} else {
			cmd = exec.Command("./crypto-cli",
				"sha256", "hash",
				"--file", h.fileEntry.Text,
			)
		}

		output, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				h.statusLabel.SetText("❌ Greska")
				dialog.ShowError(fmt.Errorf(string(output)), h.parent)
				return
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "SHA-256 hash:") {
					hash := strings.TrimSpace(strings.TrimPrefix(line, "SHA-256 hash:"))
					h.hashEntry.SetText(hash)
					break
				}
			}

			h.statusLabel.SetText("✅ Hash izracunat")
		})
	}()
}

func (h *HashWindow) saveHash() {
	if h.hashEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Nema hash-a za cuvanje"), h.parent)
		return
	}

	if h.saveFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Izaberite fajl za cuvanje"), h.parent)
		return
	}

	go func() {
		err := os.WriteFile(h.saveFileEntry.Text, []byte(h.hashEntry.Text), 0644)

		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, h.parent)
				return
			}

			dialog.ShowInformation("Uspeh",
				fmt.Sprintf("Hash sačuvan u %s", h.saveFileEntry.Text),
				h.parent)
		})
	}()
}

func (h *HashWindow) verifyHash() {
	if h.verifyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Izaberite fajl za verifikaciju"), h.parent)
		return
	}

	if h.verifyHashEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Unesite ocekivani hash"), h.parent)
		return
	}

	h.verifyResult.SetText("🔄 Proveravam...")

	go func() {
		cmd := exec.Command("./crypto-cli",
			"sha256", "verify",
			"--file", h.verifyFileEntry.Text,
			"--hash", h.verifyHashEntry.Text,
		)

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		fyne.Do(func() {
			if strings.Contains(outputStr, "match") || strings.Contains(outputStr, "Hashes match") {
				h.verifyResult.SetText("✅ Hash se poklapa - fajl je validan!")
			} else {
				h.verifyResult.SetText("❌ Hash se NE poklapa - fajl je ostecen!")
			}

			if err != nil {
			}
		})
	}()
}
