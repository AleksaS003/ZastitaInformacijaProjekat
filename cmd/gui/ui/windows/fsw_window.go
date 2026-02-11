package windows

import (
	"bufio"
	"fmt"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type FSWWindow struct {
	parent fyne.Window

	// Polja
	watchDirEntry   *widget.Entry
	outputDirEntry  *widget.Entry
	keyFileEntry    *widget.Entry
	algorithmSelect *widget.Select

	// Status
	statusLabel *widget.Label
	logList     *widget.List
	events      []string

	// Proces
	cmd       *exec.Cmd
	isRunning bool
	stopChan  chan bool
}

func NewFSWWindow(parent fyne.Window) *FSWWindow {
	return &FSWWindow{
		parent:   parent,
		events:   []string{},
		stopChan: make(chan bool),
	}
}

func (f *FSWWindow) Build() *fyne.Container {
	f.createWidgets()
	return f.createLayout()
}

func (f *FSWWindow) createWidgets() {
	// Watch direktorijum
	f.watchDirEntry = widget.NewEntry()
	f.watchDirEntry.SetText("./watch")

	// Output direktorijum
	f.outputDirEntry = widget.NewEntry()
	f.outputDirEntry.SetText("./encrypted")

	// Key fajl
	f.keyFileEntry = widget.NewEntry()
	f.keyFileEntry.SetPlaceHolder("Izaberi key fajl...")

	// Algoritam
	f.algorithmSelect = widget.NewSelect([]string{"LEA", "LEA-PCBC"}, func(s string) {})
	f.algorithmSelect.SetSelected("LEA-PCBC")

	// Status
	f.statusLabel = widget.NewLabel("Status: Zaustavljen")

	// Log lista
	f.logList = widget.NewList(
		func() int { return len(f.events) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if len(f.events) > 0 && id < len(f.events) {
				item.(*widget.Label).SetText(f.events[len(f.events)-1-id])
			}
		},
	)
}

func (f *FSWWindow) createLayout() *fyne.Container {
	// Dugmad za pregled direktorijuma
	btnSelectWatch := widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				f.watchDirEntry.SetText(uri.Path())
			}
		}, f.parent).Show()
	})

	btnSelectOutput := widget.NewButton("📂 Pregledaj", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				f.outputDirEntry.SetText(uri.Path())
			}
		}, f.parent).Show()
	})

	btnSelectKey := widget.NewButton("🔑 Pregledaj", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				f.keyFileEntry.SetText(reader.URI().Path())
			}
		}, f.parent).Show()
	})

	// Kontrolna dugmad
	btnStart := widget.NewButton("▶ POKRENI FSW", func() {
		f.startFSW()
	})
	btnStart.Importance = widget.HighImportance

	btnStop := widget.NewButton("⏹ ZAUSTAVI", func() {
		f.stopFSW()
	})
	btnStop.Importance = widget.DangerImportance

	btnEncryptExisting := widget.NewButton("📁 Enkriptuj postojeće", func() {
		f.encryptExisting()
	})

	// Kontrole
	controls := container.NewVBox(
		widget.NewLabelWithStyle("📁 File System Watcher", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			widget.NewLabel("Watch direktorijum:"),
			f.watchDirEntry,
			btnSelectWatch,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Output direktorijum:"),
			f.outputDirEntry,
			btnSelectOutput,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Key fajl:"),
			f.keyFileEntry,
			btnSelectKey,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Algoritam:"),
			f.algorithmSelect,
		),
		container.NewHBox(btnStart, btnStop, btnEncryptExisting),
		f.statusLabel,
	)

	// Glavni sadržaj
	content := container.NewBorder(
		controls,
		nil,
		nil,
		nil,
		container.NewBorder(
			widget.NewLabelWithStyle("📋 Događaji:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			f.logList,
		),
	)

	return content
}

func (f *FSWWindow) startFSW() {
	if f.isRunning {
		return
	}

	if f.keyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), f.parent)
		return
	}

	f.statusLabel.SetText("Status: Pokrećem FSW...")
	f.isRunning = true

	go func() {
		f.cmd = exec.Command("./crypto-cli",
			"fsw", "start",
			"--watch", f.watchDirEntry.Text,
			"--output", f.outputDirEntry.Text,
			"--keyfile", f.keyFileEntry.Text,
			"--algo", f.algorithmSelect.Selected,
		)

		stdout, err := f.cmd.StdoutPipe()
		if err != nil {
			fyne.Do(func() {
				f.statusLabel.SetText("Status: Greška pri pokretanju")
				f.isRunning = false
			})
			return
		}

		err = f.cmd.Start()
		if err != nil {
			fyne.Do(func() {
				f.statusLabel.SetText("Status: Greška pri pokretanju")
				f.isRunning = false
			})
			return
		}

		fyne.Do(func() {
			f.statusLabel.SetText("Status: FSW aktivan - nadgleda: " + f.watchDirEntry.Text)
		})

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fyne.Do(func() {
				f.events = append(f.events, line)
				f.logList.Refresh()
				f.logList.ScrollTo(len(f.events) - 1)
			})
		}
	}()
}

func (f *FSWWindow) stopFSW() {
	if !f.isRunning {
		return
	}

	if f.cmd != nil && f.cmd.Process != nil {
		f.cmd.Process.Kill()
	}

	f.isRunning = false
	f.statusLabel.SetText("Status: Zaustavljen")
}

func (f *FSWWindow) encryptExisting() {
	if f.keyFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), f.parent)
		return
	}

	go func() {
		cmd := exec.Command("./crypto-cli",
			"fsw", "encrypt-existing",
			"--watch", f.watchDirEntry.Text,
			"--output", f.outputDirEntry.Text,
			"--keyfile", f.keyFileEntry.Text,
			"--algo", f.algorithmSelect.Selected,
		)

		output, err := cmd.CombinedOutput()

		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf(string(output)), f.parent)
				return
			}
			dialog.ShowInformation("Uspeh", string(output), f.parent)
		})
	}()
}
