package windows

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type LogsWindow struct {
	parent fyne.Window

	logList     *widget.List
	logEntries  []string
	filterEntry *widget.Entry
}

func NewLogsWindow(parent fyne.Window) *LogsWindow {
	return &LogsWindow{
		parent:     parent,
		logEntries: []string{},
	}
}

func (l *LogsWindow) Build() *fyne.Container {

	l.filterEntry = widget.NewEntry()
	l.filterEntry.SetPlaceHolder("Filter po aktivnosti...")

	btnFilter := widget.NewButton("🔍 Filtriraj", func() {
		l.loadLogs()
	})

	btnRefresh := widget.NewButton("⟲ Osveži", func() {
		l.loadLogs()
	})

	btnClear := widget.NewButton("🧹 Obriši", func() {
		dialog.NewConfirm("Brisanje logova",
			"Da li ste sigurni da želite da obrišete sve logove?",
			func(confirm bool) {
				if confirm {
					l.clearLogs()
				}
			}, l.parent).Show()
	})

	btnStats := widget.NewButton("📊 Statistika", func() {
		l.showStats()
	})

	l.logList = widget.NewList(
		func() int { return len(l.logEntries) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(l.logEntries[len(l.logEntries)-1-id])
		},
	)

	toolbar := container.NewHBox(
		l.filterEntry,
		btnFilter,
		btnRefresh,
		btnClear,
		btnStats,
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("📜 Pregled logova", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			toolbar,
		),
		nil, nil, nil,
		l.logList,
	)

	l.loadLogs()

	return content
}

func (l *LogsWindow) loadLogs() {
	logFile := "./logs/crypto-app.log"

	file, err := os.Open(logFile)
	if err != nil {
		l.logEntries = []string{"Log fajl nije pronađen"}
		l.logList.Refresh()
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if l.filterEntry.Text != "" {
			if !strings.Contains(strings.ToLower(line), strings.ToLower(l.filterEntry.Text)) {
				continue
			}
		}

		lines = append(lines, line)
	}

	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}

	l.logEntries = lines
	l.logList.Refresh()
	l.logList.ScrollTo(len(l.logEntries) - 1)
}

func (l *LogsWindow) clearLogs() {
	cmd := exec.Command("./crypto-cli", "logs", "clear", "--yes")
	_, err := cmd.CombinedOutput()

	if err != nil {
		dialog.ShowError(err, l.parent)
		return
	}

	l.logEntries = []string{"Logovi obrisani"}
	l.logList.Refresh()
	dialog.ShowInformation("Uspeh", "Logovi su obrisani", l.parent)
}

func (l *LogsWindow) showStats() {
	cmd := exec.Command("./crypto-cli", "logs", "stats")
	output, err := cmd.CombinedOutput()

	if err != nil {
		dialog.ShowError(err, l.parent)
		return
	}

	dialog.ShowInformation("Statistika logova", string(output), l.parent)
}
