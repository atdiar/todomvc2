package main

import (
	. "github.com/atdiar/particleui"
	doc "github.com/atdiar/particleui/drivers/js"
)

type TodosListElement struct {
	*Element
}

func (t TodosListElement) GetList() List {
	var tdl List
	res, ok := t.AsElement().Get("ui", "todoslist")
	if !ok {
		tdl = NewList().Commit()
	}
	tdl, ok = res.(List)
	if !ok {
		tdl = NewList().Commit()
	}
	return tdl
}

func (t TodosListElement) SetList(tdl List) TodosListElement {
	t.SetDataSetUI("todoslist", tdl)
	return t
}

func TodoListFromRef(ref *Element) TodosListElement {
	return TodosListElement{ref}
}

func (t TodosListElement) AsViewElement() ViewElement {
	return ViewElement{t.AsElement()}
}

func displayWhen(filter string) func(Value) bool {
	return func(v Value) bool {
		o := v.(Todo)
		cplte, _ := o.Get("completed")
		complete := cplte.(Bool)

		if filter == "active" {
			if complete {
				return false
			}
			return true
		}

		if filter == "completed" {
			if !complete {
				return false
			}
			return true
		}
		return true
	}
}

func newTodoListElement(document *doc.Document, id string, options ...string) *Element {
	t := document.Ul.WithID(id, options...)
	doc.AddClass(t.AsElement(), "todo-list")

	tview := NewViewElement(t.AsElement(), NewView("all"), NewView("active"), NewView("completed"))
	t.OnRouterMounted(func(r *Router) {
		names := NewList(String("all"), String("active"), String("completed")).Commit()
		links := NewList(
			String(r.NewLink("all").URI()),
			String(r.NewLink("active").URI()),
			String(r.NewLink("completed").URI()),
		).Commit()
		filterslist := NewObject()
		filterslist.Set("names", names)
		filterslist.Set("urls", links)

		t.AsElement().SetDataSetUI("filterslist", filterslist.Commit())
	})

	tview.AsElement().Watch("ui", "filter", tview, NewMutationHandler(func(evt MutationEvent) bool {
		evt.Origin().TriggerEvent("renderlist")
		return false
	}))

	tview.AsElement().Watch("ui", "todoslist", tview, NewMutationHandler(func(evt MutationEvent) bool {
		newlist := evt.NewValue().(List)

		for _, v := range newlist.UnsafelyUnwrap() {
			o := v.(Todo)
			ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()), o)
			if !ok {
				ntd = TodosListElement{evt.Origin()}.NewTodo(o)
			} else {
				ntd.SetDataSetUI("todo", o)
			}
		}

		//TodosListElement{evt.Origin()}.signalUpdate()
		evt.Origin().TriggerEvent("renderlist")
		return false
	}))

	tview.OnActivated("all", NewMutationHandler(func(evt MutationEvent) bool {
		evt.Origin().SetUI("filter", String("all"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-all")
		//evt.Origin().TriggerEvent("renderlist")
		return false
	}))
	tview.OnActivated("active", NewMutationHandler(func(evt MutationEvent) bool {
		evt.Origin().SetUI("filter", String("active"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-active")
		//evt.Origin().TriggerEvent("renderlist")
		return false
	}))
	tview.OnActivated("completed", NewMutationHandler(func(evt MutationEvent) bool {
		evt.Origin().SetUI("filter", String("completed"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-completed")
		//evt.Origin().TriggerEvent("renderlist")
		return false
	}))

	t.WatchEvent("renderlist", t, NewMutationHandler(func(evt MutationEvent) bool {
		t := evt.Origin()

		// Retrieve current filter
		filterval, ok := t.Get("ui", "filter")
		var filter string
		if !ok {
			filter = "all"
		} else {
			filter = string(filterval.(String))
		}

		var todos List
		tlist, ok := t.Get("ui", "todoslist")
		if ok {
			todos = tlist.(List)
		} else {
			todos = NewList().Commit()
		}

		length := len(todos.UnsafelyUnwrap())
		var newChildren = make([]*Element, 0, length)

		todos.Range(func(i int, v Value) bool {
			o := v.(Todo)
			if displayWhen(filter)(o) {
				ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()), o)
				if !ok {
					panic("todo not found for rendering...")
				}
				newChildren = append(newChildren, ntd.AsElement())
			}
			return false
		})

		t.SetChildren(newChildren...)

		return false
	}))

	return t.AsElement()
}

func NewTodoList(d *doc.Document, id string, options ...string) TodosListElement {
	return TodosListElement{newTodoListElement(d, id, options...)}
}

func (t TodosListElement) NewTodo(o Todo) TodoElement {

	ntd := newTodoElement(doc.GetDocument(t.AsElement()), o)
	id, _ := o.Get("id")
	idstr := id.(String)

	t.Watch("ui", "todo", ntd, NewMutationHandler(func(evt MutationEvent) bool { // escalates back to the todolist the data changes issued at the todo Element level
		var tdl List
		res, ok := t.GetUI("todoslist")
		if !ok {
			tdl = NewList().Commit()
		} else {
			tdl = res.(List)
		}

		newval := evt.NewValue()

		rawlist := tdl.UnsafelyUnwrap()

		for i, rawtodo := range rawlist {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")

			if oldid == idstr {
				rawlist[i] = newval
				t.SetList(NewListFrom(rawlist))
				break
			}
		}
		return false
	}))

	t.WatchEvent("delete", ntd, NewMutationHandler(func(evt MutationEvent) bool {
		var tdl List
		res, ok := t.AsElement().GetUI("todoslist")
		if !ok {
			tdl = NewList().Commit()
		} else {
			tdl = res.(List)
		}
		ntdl := NewList()
		var i int

		for _, rawtodo := range tdl.UnsafelyUnwrap() {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")
			if oldid == idstr {
				t, ok := FindTodoElement(doc.GetDocument(evt.Origin()), rawtodo.(Todo))
				if ok {
					Delete(t.AsElement())
				}
				continue
			}
			ntdl = ntdl.Append(rawtodo)
			i++
		}

		t.SetList(ntdl.Commit())
		return false
	}))

	return ntd
}
