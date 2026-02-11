package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type NetworkWindow struct {
	parent fyne.Window

	modeSelect *widget.RadioGroup

	serverAddressEntry *widget.Entry
	serverOutputEntry  *widget.Entry
	serverKeyEntry     *widget.Entry

	clientAddressEntry *widget.Entry
	clientFileEntry    *widget.Entry
	clientKeyEntry     *widget.Entry
	clientAlgoSelect   *widget.Select

	serverStatusLabel *widget.Label
	clientStatusLabel *widget.Label
	serverLogLabel    *widget.Label

	serverCmd       *exec.Cmd
	serverIsRunning bool
	serverStopChan  chan bool

	serverBox *fyne.Container
	clientBox *fyne.Container
}

func NewNetworkWindow(parent fyne.Window) *NetworkWindow {
	return &NetworkWindow{
		parent:         parent,
		serverStopChan: make(chan bool),
	}
}

func (n *NetworkWindow) Build() *fyne.Container {
	n.createWidgets()
	return n.createLayout()
}

func (n *NetworkWindow) createWidgets() {
	n.modeSelect = widget.NewRadioGroup(
		[]string{"Server", "Client"},
		func(s string) { n.onModeChange(s) },
	)
	n.modeSelect.SetSelected("Server")
	n.modeSelect.Horizontal = true

	n.serverAddressEntry = widget.NewEntry()
	n.serverAddressEntry.SetText(":8080")

	n.serverOutputEntry = widget.NewEntry()
	n.serverOutputEntry.SetText("./received")

	n.serverKeyEntry = widget.NewEntry()
	n.serverKeyEntry.SetPlaceHolder("Izaberi key fajl...")

	n.serverStatusLabel = widget.NewLabel("Status: Server nije pokrenut")
	n.serverLogLabel = widget.NewLabel("")

	n.clientAddressEntry = widget.NewEntry()
	n.clientAddressEntry.SetText("localhost:8080")

	n.clientFileEntry = widget.NewEntry()
	n.clientFileEntry.SetPlaceHolder("Izaberi fajl za slanje...")

	n.clientKeyEntry = widget.NewEntry()
	n.clientKeyEntry.SetPlaceHolder("Izaberi key fajl...")

	n.clientAlgoSelect = widget.NewSelect([]string{"LEA", "LEA-PCBC"}, func(s string) {})
	n.clientAlgoSelect.SetSelected("LEA-PCBC")

	n.clientStatusLabel = widget.NewLabel("Status: Nije povezan")
}

func (n *NetworkWindow) createLayout() *fyne.Container {
	btnServerSelectOutput := widget.NewButton("📂 Output", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				n.serverOutputEntry.SetText(uri.Path())
			}
		}, n.parent).Show()
	})

	btnServerSelectKey := widget.NewButton("🔑 Key", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				n.serverKeyEntry.SetText(reader.URI().Path())
			}
		}, n.parent).Show()
	})

	btnServerStart := widget.NewButton("▶ POKRENI SERVER", func() {
		n.startServer()
	})
	btnServerStart.Importance = widget.HighImportance

	btnServerStop := widget.NewButton("⏹ ZAUSTAVI SERVER", func() {
		n.stopServer()
	})
	btnServerStop.Importance = widget.DangerImportance

	n.serverBox = container.NewVBox(
		widget.NewLabelWithStyle("🖥️ TCP Server", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(3,
			widget.NewLabel("Adresa:"),
			n.serverAddressEntry,
			widget.NewLabel("(host:port)"),
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Output dir:"),
			n.serverOutputEntry,
			btnServerSelectOutput,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Key fajl:"),
			n.serverKeyEntry,
			btnServerSelectKey,
		),
		container.NewHBox(btnServerStart, btnServerStop),
		n.serverStatusLabel,
		widget.NewSeparator(),
		n.serverLogLabel,
	)

	btnClientSelectFile := widget.NewButton("📂 Fajl", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				n.clientFileEntry.SetText(reader.URI().Path())
			}
		}, n.parent).Show()
	})

	btnClientSelectKey := widget.NewButton("🔑 Key", func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				n.clientKeyEntry.SetText(reader.URI().Path())
			}
		}, n.parent).Show()
	})

	btnClientSend := widget.NewButton("📤 POŠALJI FAJL", func() {
		n.sendFile()
	})
	btnClientSend.Importance = widget.HighImportance

	n.clientBox = container.NewVBox(
		widget.NewLabelWithStyle("📤 TCP Client", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(3,
			widget.NewLabel("Server adresa:"),
			n.clientAddressEntry,
			widget.NewLabel(""),
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Fajl za slanje:"),
			n.clientFileEntry,
			btnClientSelectFile,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("Key fajl:"),
			n.clientKeyEntry,
			btnClientSelectKey,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Algoritam:"),
			n.clientAlgoSelect,
		),
		btnClientSend,
		n.clientStatusLabel,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("🌐 TCP Server/Client", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		n.modeSelect,
		widget.NewSeparator(),
		container.NewStack(
			n.serverBox,
			n.clientBox,
		),
	)

	n.updateVisibility()
	return content
}

func (n *NetworkWindow) onModeChange(mode string) {
	n.updateVisibility()
}

func (n *NetworkWindow) updateVisibility() {
	if n.serverBox == nil || n.clientBox == nil {
		return
	}

	isServer := n.modeSelect.Selected == "Server"
	n.serverBox.Hidden = !isServer
	n.clientBox.Hidden = isServer
}

func (n *NetworkWindow) startServer() {
	if n.serverKeyEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), n.parent)
		return
	}

	if n.serverIsRunning {
		dialog.ShowInformation("Info", "Server je vec pokrenut", n.parent)
		return
	}

	n.serverStatusLabel.SetText("Status: Pokrecem server...")

	n.serverCmd = exec.Command("./crypto-cli",
		"server",
		"--address", n.serverAddressEntry.Text,
		"--output", n.serverOutputEntry.Text,
		"--keyfile", n.serverKeyEntry.Text,
	)

	err := n.serverCmd.Start()
	if err != nil {
		n.serverStatusLabel.SetText("Status: Greska pri pokretanju")
		dialog.ShowError(err, n.parent)
		return
	}

	n.serverIsRunning = true
	n.serverStatusLabel.SetText(fmt.Sprintf("Status: Server radi na %s", n.serverAddressEntry.Text))
	n.serverLogLabel.SetText("Server pokrenut, cekam konekcije...")

	go func() {
		waitErr := n.serverCmd.Wait()
		fyne.Do(func() {
			n.serverIsRunning = false
			n.serverStatusLabel.SetText("Status: Server zaustavljen")

			if waitErr != nil {
				n.serverLogLabel.SetText(fmt.Sprintf("Server izasao sa greskom: %v", waitErr))
			} else {
				n.serverLogLabel.SetText("Server normalno zavrsio")
			}
		})
	}()
}

func (n *NetworkWindow) stopServer() {
	if !n.serverIsRunning || n.serverCmd == nil || n.serverCmd.Process == nil {
		dialog.ShowInformation("Info", "Server nije pokrenut", n.parent)
		return
	}

	n.serverStatusLabel.SetText("Status: Zaustavljam server...")

	err := n.serverCmd.Process.Signal(os.Interrupt)
	if err != nil {
		err = n.serverCmd.Process.Kill()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Greska pri zaustavljanju servera: %v", err), n.parent)
			return
		}
	}

	n.serverLogLabel.SetText("Signal za zaustavljanje poslat serveru")
	dialog.ShowInformation("Uspeh", "Signal za zaustavljanje poslat serveru", n.parent)
}

func (n *NetworkWindow) sendFile() {
	if n.clientFileEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati fajl"), n.parent)
		return
	}

	if n.clientKeyEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("Morate izabrati key fajl"), n.parent)
		return
	}

	n.clientStatusLabel.SetText("📤 Saljem fajl...")

	go func() {
		cmd := exec.Command("./crypto-cli",
			"client",
			"--address", n.clientAddressEntry.Text,
			"--file", n.clientFileEntry.Text,
			"--keyfile", n.clientKeyEntry.Text,
			"--algo", n.clientAlgoSelect.Selected,
		)

		output, err := cmd.CombinedOutput()
		_ = err
		outputStr := string(output)

		fyne.Do(func() {
			if strings.Contains(outputStr, "File sent successfully") {
				n.clientStatusLabel.SetText("✅ Fajl uspesno poslat!")
				dialog.ShowInformation("Uspeh",
					fmt.Sprintf("Fajl '%s' je uspesno poslat na server %s",
						filepath.Base(n.clientFileEntry.Text),
						n.clientAddressEntry.Text),
					n.parent)

			} else {
				n.clientStatusLabel.SetText("❌ Greska pri slanju")
				dialog.ShowError(fmt.Errorf(outputStr), n.parent)
			}
		})
	}()
}
