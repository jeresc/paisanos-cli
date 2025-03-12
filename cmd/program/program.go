package program

import (
	"log"
	"os"
	"paisanos-cli/cmd/flags"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

type Project struct {
	Editors   []string
	Exit      bool
	EditorMap map[flags.Editor]Editor
	OSCheck   map[string]bool
}

type Editor struct {
	DisplayName string
}

func (p *Project) ExitCLI(tprogram *tea.Program) {
	if p.Exit {
		if err := tprogram.ReleaseTerminal(); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}

func (p *Project) CheckOS() {
	p.OSCheck = make(map[string]bool)

	if runtime.GOOS != "windows" {
		p.OSCheck["UnixBased"] = true
	}
	if runtime.GOOS == "linux" {
		p.OSCheck["linux"] = true
	}
	if runtime.GOOS == "darwin" {
		p.OSCheck["darwin"] = true
	}
}
