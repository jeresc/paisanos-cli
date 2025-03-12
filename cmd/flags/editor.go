package flags

import (
	"fmt"
	"slices"
	"strings"
)

type Editor string

const (
	Vscode Editor = "vscode"
	Nvim   Editor = "nvim"
	Cursor Editor = "cursor"
	Xcode  Editor = "xcode"
	None   Editor = "none"
)

var AllowedEditors = []string{string(Vscode), string(Nvim), string(Cursor), string(Xcode), string(None)}

func (f Editor) String() string {
	return string(f)
}

func (f *Editor) Type() string {
	return "Editor"
}

func (f *Editor) Set(value string) error {
	if slices.Contains(AllowedEditors, value) {
		*f = Editor(value)
		return nil
	}

	return fmt.Errorf("Editor to use. Allowed values: %s", strings.Join(AllowedEditors, ", "))
}
