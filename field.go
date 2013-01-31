package goform

import (
	"fmt"
	h "github.com/metakeule/goh4"
)

type Field struct {
	*h.Element
	Name        string
	Type        Type
	Required    bool
	Constructor Constructor // only for struct Fields, should return a pointer to a struct
	Selection   interface{} // if only certain values are allowed, should be an array of things that are of the same type as value
}

// sets the infos of the inner Field tag
func (ø *Field) setFieldInfos() {
	fs := ø.Element.Fields()
	if len(fs) == 0 {
		panic("got no form Field in " + ø.Element.String())
	}
	fs[0].Add(h.Class("field"), h.Id(ø.Name), h.Attr("name", ø.Name))
	if ø.Required {
		fs[0].Add(h.Attr("required", "required"))
	}
	ø.setLabelInfos()
}

func (ø *Field) setLabelInfos() {
	_, label := ø.Element.Any(h.Tag("label"))
	if label != nil {
		label.Add(h.Attr("for", ø.Name))
		if ø.Required {
			label.Add(h.Class("required"))
		}
	}
}

func (ø *Field) CheckAllowed(form *FormHandler) {
	if ø.Selection == nil {
		return
	}

	switch ø.Type {
	case Int:
		a := ø.Selection.([]int)
		val := form.Ints[ø]
		if !ø.hasInt(a, val) {
			form.AddFieldError(ø, fmt.Errorf("%#v not in %+v", val, a))
		}
	case Float:
		a := ø.Selection.([]float32)
		val := form.Floats[ø]
		if !ø.hasFloat(a, val) {
			form.AddFieldError(ø, fmt.Errorf("%#v not in %+v", val, a))
		}
	case String:
		a := ø.Selection.([]string)
		val := form.Strings[ø]
		if !ø.hasString(a, val) {
			form.AddFieldError(ø, fmt.Errorf("%#v not in %+v", val, a))
		}
	}
}

func (ø *Field) hasInt(a []int, i int) (has bool) {
	has = false
	for _, e := range a {
		if e == i {
			return true
		}
	}
	return
}

func (ø *Field) hasFloat(a []float32, i float32) (has bool) {
	has = false
	for _, e := range a {
		if e == i {
			return true
		}
	}
	return
}

func (ø *Field) hasString(a []string, i string) (has bool) {
	has = false
	for _, e := range a {
		if e == i {
			return true
		}
	}
	return
}
