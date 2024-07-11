package main

import (
	"strings"

	ui "github.com/atdiar/particleui"
	doc "github.com/atdiar/particleui/drivers/js"
)

func NewTodoInput(document *doc.Document, id string, options ...string) doc.InputElement {
	todosinput := document.Input.WithID(id, "text", options...)
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")

	doc.Autofocus(todosinput.AsElement())

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		v, ok := evt.Value().(ui.Object).Get("value")
		if !ok {
			todosinput.SyncUISetData("value", ui.String(""))
			return false
		}
		s := v.(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.SyncUISetData("value", ui.String(str))
		return false
	}))

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		todosinput := doc.InputElement{evt.CurrentTarget()}

		v := evt.(doc.KeyboardEvent).Key()
		if v == "Enter" {
			evt.PreventDefault()

			val, ok := evt.Value().(ui.Object).Get("value")
			if !ok {
				// TODO clear input? panic?
				return false
			}

			if !ui.Equal(val, ui.String("")) {
				todosinput.TriggerEvent("newtodo", val)
			}
			//todosinput.SyncUISetData("value", ui.String(""))
			todosinput.Clear()

			// todo: apply mutation to state and watch state instead of todo event?
			// require clearing state because of idempotency.
		}
		return false
	}))

	return todosinput
}
