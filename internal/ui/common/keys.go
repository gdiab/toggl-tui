package common

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines global key bindings.
type KeyMap struct {
	Quit      key.Binding
	Help      key.Binding
	Start     key.Binding
	Manual    key.Binding
	Refresh   key.Binding
	Stop      key.Binding
	Up        key.Binding
	Down      key.Binding
	Submit    key.Binding
	Cancel    key.Binding
	NextField key.Binding
	PrevField key.Binding
}

// Keys is the default key map.
var Keys = KeyMap{
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Start:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start timer")),
	Manual:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "manual entry")),
	Refresh:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Stop:      key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "stop timer")),
	Up:        key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:      key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Submit:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	NextField: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	PrevField: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
}
