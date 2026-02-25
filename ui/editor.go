package ui

import (
	"errors"
	"strings"
)

type editorCommand struct {
	Executable string
	Args       []string
}

var errNoEditorFound = errors.New(
	"no editor found. please set the EDITOR environment variable or install 'code' or 'vi'",
)

// resolveEditor determines which external editor to use based on the following
// precedence:
//
//  1. If the EDITOR environment variable is set and not empty, use that.
//  2. If the 'code' command (VS Code) is available on the system, use that with
//     the '--wait' flag.
//  3. If the 'vi' command is available on the system, use that.
//  4. If none of the above are available, return an error indicating that no
//     editor was found.
func resolveEditor(
	lookupEnv func(string) (string, bool),
	commandExists func(string) bool,
) (editorCommand, error) {
	editorEnv, ok := lookupEnv("EDITOR")
	if ok && editorEnv != "" {
		tokens := strings.Fields(editorEnv)
		return editorCommand{
			Executable: tokens[0],
			Args:       tokens[1:],
		}, nil
	}

	if commandExists("code") {
		return editorCommand{
			Executable: "code",
			Args:       []string{"--wait"},
		}, nil
	}

	if commandExists("vi") {
		return editorCommand{
			Executable: "vi",
		}, nil
	}

	return editorCommand{}, errNoEditorFound
}
