package cli

import "bytes"

func runCommand(app *App, args []string) (string, error) {
	cmd := NewRootCommand(app)
	cmd.SetArgs(args)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out.String(), err
}
