package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/guptarohit/asciigraph"
)

type boxL struct {
	views.BoxLayout
}

var app = &views.Application{}
var box = &boxL{}

func (m *boxL) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyCtrlC {
			app.Quit()
			return true
		}
	}
	return m.BoxLayout.HandleEvent(ev)
}

var errorText = "bob"

func main() {
	commandsToRun := os.Args[1:]

	title := &views.TextBar{}
	title.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorYellow).
		Foreground(tcell.ColorBlack))
	title.SetCenter("ASCII graph version of watch", tcell.StyleDefault)
	title.SetLeft("CTRLC to exit", tcell.StyleDefault.
		Background(tcell.ColorBlue).
		Foreground(tcell.ColorWhite))
	title.SetRight("==>X", tcell.StyleDefault)

	inner := views.NewBoxLayout(views.Vertical)

	textArea := views.NewText()
	textArea.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorLime))
	inner.AddWidget(textArea, 1)

	values := []float64{}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorText = fmt.Sprintf("Error1:%v", r)
				app.Quit()
			}
		}()

		for {
			command := commandsToRun[0]

			out, err := exec.Command("/bin/bash", "-c", commandsToRun[0]).Output()
			if err != nil {
				errorText = err.Error()
				textArea.SetText(err.Error())
			}

			value, err := strconv.Atoi(strings.TrimSpace(string(out)))
			if err != nil {
				errorText = err.Error()
				textArea.SetText(err.Error())
			}

			values = append(values, float64(value))

			// w, _ := textArea.Size()
			graph := asciigraph.Plot(values, asciigraph.Height(20), asciigraph.Width(80))

			textArea.SetText("Command: '" + command + "' \n" + graph + "\n" + time.Now().String())
			app.Refresh()

			time.Sleep(1 * time.Second)
		}
	}()

	box.SetOrientation(views.Vertical)
	box.AddWidget(title, 0)
	box.AddWidget(inner, 1)
	app.SetRootWidget(box)
	if e := app.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		fmt.Println(errorText)
		os.Exit(1)
	}
}
