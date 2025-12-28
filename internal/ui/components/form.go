package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Field struct {
	Label string
	Value string
	Type  string // "text", "textarea"
}

type Form struct {
	fields       []Field
	focusedField int
}

func NewForm(fields []Field) *Form {
	return &Form{fields: fields}
}

func (f *Form) Update(msg tea.KeyMsg) {
	switch msg.String() {
	case "tab", "down":
		f.focusedField = (f.focusedField + 1) % len(f.fields)

	case "shift+tab", "up":
		f.focusedField--
		if f.focusedField < 0 {
			f.focusedField = len(f.fields) - 1
		}

	case "backspace":
		if len(f.fields[f.focusedField].Value) > 0 {
			v := f.fields[f.focusedField].Value
			f.fields[f.focusedField].Value = v[:len(v)-1]
		}

	default:
		if msg.Type == tea.KeyRunes {
			f.fields[f.focusedField].Value += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			f.fields[f.focusedField].Value += " "
		}
	}
}

func (f *Form) View() string {
	var sections []string
	for i, field := range f.fields {
		isActive := i == f.focusedField
		sections = append(sections, f.renderField(field, isActive))
	}
	return "\n" + lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (f *Form) renderField(field Field, isActive bool) string {
	inputStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("7"))

	if isActive {
		inputStyle = inputStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4"))
	}

	return fmt.Sprintf("%s\n%s", field.Label, inputStyle.Render(field.Value))
}

func (f *Form) GetValues() map[string]string {
	result := make(map[string]string)
	for _, field := range f.fields {
		result[field.Label] = field.Value
	}
	return result
}
