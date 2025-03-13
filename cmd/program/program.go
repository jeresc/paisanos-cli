package program

import (
	"errors"
	"log"
	"os"
	"paisanos-cli/cmd/flags"
	"paisanos-cli/utils"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

type Project struct {
	Editors   []string
	Exit      bool
	EditorMap map[flags.Editor]Editor
	OSCheck   map[string]bool
	Username  string
	HomeDir   string
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

func (p *Project) CreateEditorMap() {
	p.EditorMap = make(map[flags.Editor]Editor)
	p.EditorMap[flags.Cursor] = Editor{DisplayName: "Cursor"}
	p.EditorMap[flags.Vscode] = Editor{DisplayName: "VSCode"}
	p.EditorMap[flags.Nvim] = Editor{DisplayName: "Neovim"}
	p.EditorMap[flags.Xcode] = Editor{DisplayName: "Xcode"}
	p.EditorMap[flags.None] = Editor{DisplayName: "None"}
}

func (p *Project) Run() error {
	p.CheckOS()
	p.CreateEditorMap()
	currentuser, _ := utils.GetCurrentUser()
	p.HomeDir = currentuser.HomeDir
	p.Username = currentuser.Username

	if !p.OSCheck["darwin"] {
		return errors.New("lo lamentamos, este comando solo funciona en macOS")
	}

	return nil
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
