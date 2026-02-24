package ui

import (
	"errors"
	"strings"
)

type EditorCommand struct {
	Executable string
	Args       []string
}

var ErrNoEditorFound = errors.New(
	"no editor found. please set the EDITOR environment variable or install 'code' or 'vi'",
)

// resolveEditor는 사용할 외부 에디터를 결정한다.
// lookupEnv는 환경변수 조회 함수, commandExists는 명령 존재 확인 함수로,
// 테스트 시 주입하여 모킹할 수 있다.
func resolveEditor(
	lookupEnv func(string) (string, bool),
	commandExists func(string) bool,
) (EditorCommand, error) {
	editorEnv, ok := lookupEnv("EDITOR")
	if ok && editorEnv != "" {
		tokens := strings.Fields(editorEnv)
		return EditorCommand{
			Executable: tokens[0],
			Args:       tokens[1:],
		}, nil
	}

	if commandExists("code") {
		return EditorCommand{
			Executable: "code",
			Args:       []string{"--wait"},
		}, nil
	}

	if commandExists("vi") {
		return EditorCommand{
			Executable: "vi",
		}, nil
	}

	return EditorCommand{}, ErrNoEditorFound
}
