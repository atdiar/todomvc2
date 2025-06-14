package main

import (
	ui "github.com/atdiar/particleui"
	. "github.com/atdiar/particleui/drivers/js"
)

func App() *Document {

	var AppSection *ui.Element
	var MainSection *ui.Element
	var MainFooter *ui.Element
	var todosinput *ui.Element
	var ToggleAllInput *ui.Element
	var TodosList *ui.Element
	var TodoCount *ui.Element
	var FilterList *ui.Element
	var ClearCompleteButton *ui.Element

	toggleallhandler := ui.NewEventHandler(func(evt ui.Event) bool {
		var ischecked bool
		v, ok := evt.Target().Get("data", "checked")
		if !ok {
			nativeinput, ok := JSValue(evt.Target())
			if !ok {
				panic("no js value")
			}
			ischecked = nativeinput.Get("checked").Bool()
			evt.Target().SetUI("checked", ui.Bool(ischecked))

		} else {
			ischecked = !v.(ui.Bool).Bool()
			evt.Target().SetUI("checked", ui.Bool(ischecked))
		}

		evt.Target().SyncUISetData("checked", ui.Bool(ischecked))
		evt.Target().TriggerEvent("toggleall", ui.Bool(ischecked))

		return false
	})

	ClearCompleteHandler := ui.NewEventHandler(func(evt ui.Event) bool {
		ClearCompleteButton := evt.Target()
		ClearCompleteButton.TriggerEvent("clear", ui.Bool(true))
		return false
	})

	document := NewDocument("Todo-App", EnableScrollRestoration())

	document.Head().AppendChild(
		E(document.Link.WithID("todocss").
			SetRel("stylesheet").
			SetHref("./assets/styles/todomvc.css"),
		),
	)

	E(document.Body(),
		Children(
			E(AriaChangeAnnouncerFor(document)),
			E(document.Section.WithID("todoapp"),
				Ref(&AppSection),
				Class("todoapp"),
				Children(
					E(document.Header.WithID("header"),
						Class("header"),
						Children(
							E(document.H1.WithID("apptitle").SetText("Todo")),
							E(NewTodoInput(document, "new-todo"),
								Ref(&todosinput),
								Class("new-todo"),
							),
						),
					),
					E(document.Section.WithID("main"),
						Ref(&MainSection),
						Class("main"),
						Children(
							E(document.Input.WithID("toggle-all", "checkbox"),
								Ref(&ToggleAllInput),
								Class("toggle-all"),
								Listen("click", toggleallhandler),
							),
							E(document.Label().For(&ToggleAllInput)),
							E(NewTodoList(document, "todo-list", EnableLocalPersistence()),
								Ref(&TodosList),
								InitRouter(Hijack("/", "/all"), ui.TrailingSlashMatters),
							),
						),
					),
					E(document.Footer.WithID("footer"),
						Ref(&MainFooter),
						Class("footer"),
						Children(
							E(NewTodoCount(document, "todo-count"), Ref(&TodoCount)),
							E(NewFilterList(document, "filters"), Ref(&FilterList)),
							E(ClearCompleteBtn(document, "clear-complete"),
								Ref(&ClearCompleteButton),
								Listen("click", ClearCompleteHandler),
							),
						),
					),
				),
			),
			E(document.Footer(),
				Class("info"),
				Children(
					E(document.Paragraph().SetText("Double-click to edit a todo")),
					E(document.Paragraph().SetText("Created with: "),
						Children(
							E(document.Anchor().SetHref("https://zui.dev").SetText("zui")),
						),
					),
				),
			),
		),
	)

	// COMPONENTS DATA RELATIONSHIPS

	// 4. Watch for new todos to insert
	AppSection.WatchEvent("newtodo", todosinput.AsElement(), ui.OnMutation(func(evt ui.MutationEvent) bool {
		tlist := TodoListFromRef(TodosList)
		tdl := tlist.GetList()

		s, ok := evt.NewValue().(ui.String)
		if !ok || s == "" {
			panic("BAD TODO")
		}
		t := NewTodo(s)
		tdl = tdl.MakeCopy().Append(t).Commit()
		tlist.SetList(tdl)

		return false
	}))

	AppSection.WatchEvent("clear", ClearCompleteButton.AsElement(), ui.OnMutation(func(evt ui.MutationEvent) bool {
		tlist := TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		ntdl := ui.NewList()
		for _, todo := range tdl.UnsafelyUnwrap() {
			t := todo.(Todo)
			c, _ := t.Get("completed")
			cpl := c.(ui.Bool)
			if !cpl {
				ntdl = ntdl.Append(todo)
			}
		}

		tlist.SetList(ntdl.Commit())
		return false
	}))

	AppSection.WatchEvent("toggleall", ToggleAllInput, ui.OnMutation(func(evt ui.MutationEvent) bool {
		status := evt.NewValue().(ui.Bool)

		tlist := TodoListFromRef(TodosList)

		tdl := tlist.GetList()
		ntdl := tdl.MakeCopy()

		for i, todo := range tdl.UnsafelyUnwrap() {
			t := todo.(Todo)
			t = t.MakeCopy().Set("completed", status).Commit()
			ntdl.Set(i, t)
		}
		tlist.SetList(ntdl.Commit())

		return false
	}))

	AppSection.Watch("ui", "todoslist", TodosList, ui.OnMutation(func(evt ui.MutationEvent) bool {
		tlist := TodoListFromRef(TodosList)
		l := tlist.GetList()

		if len(l.UnsafelyUnwrap()) == 0 {
			SetInlineCSS(MainFooter.AsElement(), "display:none")
		} else {
			SetInlineCSS(MainFooter.AsElement(), "display:block")
		}

		countcomplete := 0
		allcomplete := len(l.UnsafelyUnwrap()) > 0

		for _, todo := range l.UnsafelyUnwrap() {
			t := todo.(Todo)
			completed, ok := t.Get("completed")
			if !ok {
				panic("todo should have a completed property")
			}
			c := completed.(ui.Bool)
			if !c {
				allcomplete = false
			} else {
				countcomplete++
			}
		}

		tc := TodoCountFromRef(TodoCount)
		var itemsleft = len(l.UnsafelyUnwrap()) - countcomplete
		tc.SetCount(itemsleft)

		if countcomplete == 0 {
			SetInlineCSS(ClearCompleteButton.AsElement(), "display:none")
		} else {
			SetInlineCSS(ClearCompleteButton.AsElement(), "display:block")
		}

		if allcomplete {
			ToggleAllInput.SetDataSetUI("checked", ui.Bool(true))
		} else {
			ToggleAllInput.SetDataSetUI("checked", ui.Bool(false))
		}
		return false
	}))

	AppSection.WatchEvent("mounted", MainFooter, ui.OnMutation(func(evt ui.MutationEvent) bool {

		tlist := TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl.UnsafelyUnwrap()) == 0 {
			SetInlineCSS(MainFooter.AsElement(), "display : none")
		} else {
			SetInlineCSS(MainFooter.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	AppSection.Watch("data", "filterslist", TodosList, ui.OnMutation(func(evt ui.MutationEvent) bool {
		FilterList.AsElement().SetUI("filterslist", evt.NewValue())
		return false
	}).RunASAP())

	MainSection.WatchEvent("renderlist", TodosList, ui.OnMutation(func(evt ui.MutationEvent) bool {
		tlist := TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl.UnsafelyUnwrap()) == 0 {
			SetInlineCSS(MainSection.AsElement(), "display : none")
		} else {
			SetInlineCSS(MainSection.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	return document

}

func main() {
	ListenAndServe := NewBuilder(App)
	ListenAndServe(nil)
}
