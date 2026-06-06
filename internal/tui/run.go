package tui

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Zero Bubble Tea shell and returns a process-style exit code.
func Run(ctx context.Context, options Options) int {
	externalSink := options.RuntimeMessageSink
	var program *tea.Program
	options.RuntimeMessageSink = func(msg tea.Msg) {
		if externalSink != nil {
			externalSink(msg)
		}
		if program != nil {
			program.Send(msg)
		}
	}

	programOpts := []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	}
	if options.Skin == "zenline" {
		// enable mouse so the centered permission modal buttons are clickable
		programOpts = append(programOpts, tea.WithMouseCellMotion())
	}
	program = tea.NewProgram(newModel(ctx, options), programOpts...)

	if _, err := program.Run(); err != nil {
		return 1
	}
	return 0
}
