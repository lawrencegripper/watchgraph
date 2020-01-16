package main

import (
	"fmt"
	"io/ioutil"
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
	fullReload := false  // true to indicate each execution of the command gives a full set of data (vs single new datapoint)
	fileContent := false // true to indicate that data should be loaded from a file
	args := os.Args[1:]
	if args[0] == "--full-reload" {
		fullReload = true
		args = args[1:]
	}
	if args[0] == "--file" {
		fileContent = true
		args = args[1:]
	}

	commandsToRun := args

	fmt.Println(commandsToRun[0])

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
			var text string
			command := commandsToRun[0]
			if fileContent {
				buf, err := ioutil.ReadFile(command)
				if err != nil {
					errorText = err.Error()
					textArea.SetText(err.Error())
				}
				text = string(buf)
			} else {
				out, err := exec.Command("/bin/bash", "-c", command).Output()
				if err != nil {
					errorText = err.Error()
					textArea.SetText(err.Error())
				}
				text = string(out)
			}

			if fullReload {
				valueStrings := strings.Split(text, "\n")
				newValues := []float64{}
				for i := 0; i < len(valueStrings); i++ {
					temp := strings.TrimSpace(valueStrings[i])
					if temp != "" {
						value, err := strconv.Atoi(strings.TrimSpace(valueStrings[i]))
						if err != nil {
							errorText = err.Error()
							textArea.SetText(err.Error())
							break
						}
						newValues = append(newValues, float64(value))
					}
				}
				values = newValues
			} else {
				value, err := strconv.Atoi(strings.TrimSpace(text))
				if err != nil {
					errorText = err.Error()
					textArea.SetText(err.Error())
				}
				values = append(values, float64(value))
			}

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
