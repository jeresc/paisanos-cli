package steps

type Steps struct {
	Steps map[string]StepSchema
}

// step represents a single installation step.
type StepSchema struct {
	StepName, Headers, Field string
	Options                  []Item
}

type Item struct {
	Flag, Title, Desc string
}

func InitSteps() *Steps {
	steps := &Steps{
		Steps: map[string]StepSchema{
			"setup": {
				StepName: "setup",
				Headers:  "Editor de texto",
				Field:    "Editor",
				Options: []Item{
					{Flag: "neovim", Title: "Neovim", Desc: "Ninja 🥷"},
					{Flag: "cursor", Title: "Cursor.ai", Desc: "AI Assisted"},
					{Flag: "vscode", Title: "VSCode", Desc: "Get shit done"},
				},
			},
		},
	}

	return steps
}
