package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Field struct {
	Label       string
	Value       string
	Type        string
	Placeholder string
}

type Form struct {
	fields       []Field
	focusedField int
	CursorPos    int
	ViewWidth    int
}

func NewForm(fields []Field) *Form {
	return &Form{
		fields:       fields,
		focusedField: 0,
		CursorPos:    0,
		ViewWidth:    40,
	}
}

func (f *Form) SetViewWidth(width int) {
	if width > 10 {
		f.ViewWidth = width - 4
	}
}

func (f *Form) isTextarea() bool {
	if f.focusedField >= 0 && f.focusedField < len(f.fields) {
		return f.fields[f.focusedField].Type == "textarea"
	}
	return false
}

func (f *Form) HandleKey(key string, runes []rune) (bool, bool) {
	submit := false
	cancel := false

	if f.focusedField < 0 || f.focusedField >= len(f.fields) {
		return submit, cancel
	}

	field := &f.fields[f.focusedField]
	isTextarea := f.isTextarea()

	if f.CursorPos > len(field.Value) {
		f.CursorPos = len(field.Value)
	}
	if f.CursorPos < 0 {
		f.CursorPos = 0
	}

	switch key {
	case "tab":
		f.focusedField = (f.focusedField + 1) % len(f.fields)
		f.CursorPos = len(f.fields[f.focusedField].Value)

	case "shift+tab":
		f.focusedField--
		if f.focusedField < 0 {
			f.focusedField = len(f.fields) - 1
		}
		f.CursorPos = len(f.fields[f.focusedField].Value)

	case "down":
		if isTextarea {
			f.moveCursorToNextLine(field.Value)
		} else {
			f.focusedField = (f.focusedField + 1) % len(f.fields)
			f.CursorPos = len(f.fields[f.focusedField].Value)
		}

	case "up":
		if isTextarea {
			f.moveCursorToPrevLine(field.Value)
		} else {
			f.focusedField--
			if f.focusedField < 0 {
				f.focusedField = len(f.fields) - 1
			}
			f.CursorPos = len(f.fields[f.focusedField].Value)
		}

	case "escape":
		cancel = true

	case "enter":
		if isTextarea {
			// Check if cursor is at end of text - if so, submit
			if f.CursorPos >= len(field.Value) {
				submit = true
			} else {
				// Otherwise add newline
				field.Value = f.insertAtCursor(field.Value, "\n")
				f.CursorPos++
			}
		} else {
			submit = true
		}

	case "ctrl+enter", "ctrl+j":
		submit = true

	case "backspace":
		if f.CursorPos > 0 {
			field.Value = field.Value[:f.CursorPos-1] + field.Value[f.CursorPos:]
			f.CursorPos--
		}

	case "left":
		if f.CursorPos > 0 {
			f.CursorPos--
		}

	case "right":
		if f.CursorPos < len(field.Value) {
			f.CursorPos++
		}

	case "ctrl+a":
		f.CursorPos = 0

	case "ctrl+e":
		f.CursorPos = len(field.Value)

	case "ctrl+u":
		field.Value = ""
		f.CursorPos = 0

	default:
		if len(runes) > 0 {
			field.Value = f.insertAtCursor(field.Value, string(runes))
			f.CursorPos += len(runes)
		}
	}

	if f.CursorPos < 0 {
		f.CursorPos = 0
	}
	if f.CursorPos > len(field.Value) {
		f.CursorPos = len(field.Value)
	}

	return submit, cancel
}

func (f *Form) insertAtCursor(s, insert string) string {
	if f.CursorPos >= len(s) {
		return s + insert
	}
	return s[:f.CursorPos] + insert + s[f.CursorPos:]
}

func (f *Form) moveCursorToNextLine(value string) {
	before := value[:f.CursorPos]
	newlineIdx := strings.Index(before, "\n")
	if newlineIdx == -1 {
		f.CursorPos = len(value)
	} else {
		f.CursorPos = len(value)
	}
}

func (f *Form) moveCursorToPrevLine(value string) {
	before := value[:f.CursorPos]
	newlineIdx := strings.LastIndex(before, "\n")
	if newlineIdx == -1 {
		f.CursorPos = 0
	} else {
		f.CursorPos = newlineIdx
	}
}

func (f *Form) Update(keyMsg tea.KeyMsg) (bool, bool) {
	return f.HandleKey(keyMsg.String(), keyMsg.Runes)
}

func (f *Form) FocusedField() int {
	return f.focusedField
}

func (f *Form) View() string {
	var lines []string

	for i, field := range f.fields {
		isFocused := i == f.focusedField
		lines = append(lines, f.renderField(field, isFocused))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (f *Form) renderField(field Field, isFocused bool) string {
	isTextarea := field.Type == "textarea"

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5"))

	if isFocused {
		labelStyle = labelStyle.Underline(true)
	}

	var inputContent string
	if isTextarea {
		inputContent = f.renderTextarea(field, isFocused)
	} else {
		inputContent = f.renderSingleLine(field, isFocused)
	}

	return labelStyle.Render(field.Label) + "\n" + inputContent
}

func (f *Form) renderSingleLine(field Field, isFocused bool) string {
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	if isFocused {
		fieldStyle = fieldStyle.Foreground(lipgloss.Color("15"))
	}

	if field.Value == "" && field.Placeholder != "" {
		return fieldStyle.Copy().Foreground(lipgloss.Color("8")).Render(field.Placeholder)
	}

	if field.Value == "" {
		if isFocused {
			return lipgloss.NewStyle().
				Reverse(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("5")).
				Render(" ")
		}
		return " "
	}

	cursorPos := f.CursorPos
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(field.Value) {
		cursorPos = len(field.Value)
	}

	before := field.Value[:cursorPos]
	after := field.Value[cursorPos:]

	if isFocused {
		cursor := lipgloss.NewStyle().
			Reverse(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("5")).
			Render(" ")

		if cursorPos >= len(field.Value) {
			return fieldStyle.Render(before) + cursor
		}
		return fieldStyle.Render(before) + cursor + fieldStyle.Render(after)
	}

	return fieldStyle.Render(field.Value)
}

func (f *Form) renderTextarea(field Field, isFocused bool) string {
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	lines := strings.Split(field.Value, "\n")

	cursorPos := f.CursorPos
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(field.Value) {
		cursorPos = len(field.Value)
	}

	if len(lines) == 1 && lines[0] == "" {
		if field.Value == "" && field.Placeholder != "" {
			return fieldStyle.Copy().Foreground(lipgloss.Color("8")).Render(field.Placeholder)
		}
		if isFocused {
			return lipgloss.NewStyle().
				Reverse(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("5")).
				Render(" ")
		}
		return " "
	}

	var renderedLines []string
	currentPos := 0
	cursorLine := 0

	for i, line := range lines {
		if currentPos+len(line) >= cursorPos {
			cursorLine = i
			break
		}
		currentPos += len(line) + 1
	}

	for i, line := range lines {
		displayLine := line
		if displayLine == "" {
			displayLine = " "
		}

		if isFocused && i == cursorLine {
			relPos := 0
			if i < len(lines)-1 || (i == len(lines)-1 && cursorPos < len(field.Value)) {
				relPos = cursorPos - currentPos
				if i > 0 {
					for j := 0; j < i; j++ {
						relPos += len(lines[j]) + 1
					}
				}
			}

			if relPos < 0 {
				relPos = 0
			}
			if relPos > len(line) {
				relPos = len(line)
			}

			before := line[:relPos]
			after := line[relPos:]

			cursor := lipgloss.NewStyle().
				Reverse(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("5")).
				Render(" ")

			renderedLines = append(renderedLines, fieldStyle.Render(before)+cursor+fieldStyle.Render(after))
		} else {
			renderedLines = append(renderedLines, fieldStyle.Render(displayLine))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderedLines...)
}

func (f *Form) GetValues() map[string]string {
	result := make(map[string]string)
	for _, field := range f.fields {
		result[field.Label] = field.Value
	}
	return result
}

func (f *Form) SetValue(label string, value string) {
	for i := range f.fields {
		if f.fields[i].Label == label {
			f.fields[i].Value = value
			break
		}
	}
}

func (f *Form) FieldCount() int {
	return len(f.fields)
}
