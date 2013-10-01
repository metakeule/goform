package goform

import (
	_ "fmt"
	h "github.com/metakeule/goh4"
	. "github.com/metakeule/goh4/tag"
	"github.com/metakeule/typeconverter"
)

type Filler interface {
	Fill(m map[string]interface{}) error
	Validate() error
}

func NewForm(objects ...interface{}) (f *FormHandler) {
	f = &FormHandler{
		Types:    map[*Field]Type{},
		Fields:   map[string]*Field{},
		Order:    []h.Stringer{},
		required: []*Field{},
		Element:  FORM(ATTR("method", "POST", "enctype", "multipart/form-data")),
	}

	for _, obj := range objects {
		switch v := obj.(type) {
		case []*Field:
			for _, field := range v {
				f.AddField(field)
			}
		case *Field:
			f.AddField(v)
		default:
			f.AddHtml(v.(h.Stringer))
		}
	}
	f.Reset()
	return
}

func Selection(Field *Field, vals ...interface{}) *Field {
	switch Field.Type {
	case Int:
		allowed := []int{}
		for _, v := range vals {
			var i int
			typeconverter.Convert(v, &i)
			allowed = append(allowed, i)
		}
		Field.Selection = allowed
	case Float:
		allowed := []float32{}
		for _, v := range vals {
			iv, ok := v.(float32)
			if !ok {
				iv = float32(v.(float64))
			}

			allowed = append(allowed, iv)
		}
		Field.Selection = allowed
	case String:
		allowed := []string{}
		for _, v := range vals {
			allowed = append(allowed, v.(string))
		}
		Field.Selection = allowed
	}
	if sel := Field.Element.Any(h.Tag("select")); sel != nil {
		options := sel.All(h.Tag("option"))
		for i, opt := range options {
			var str string
			typeconverter.Convert(vals[i], &str)
			opt.Add(h.Attr("value", str))
		}
	}
	return Field
}

func Required(name string, t Typer, html ...interface{}) (ø *Field) {
	e := h.NewElement(h.Tag("form"), h.WithoutDecoration)
	e.Add(html...)
	ø = &Field{Name: name, Type: t.Type(), Element: e, Required: true}
	if c, ok := t.(Constructor); ok {
		ø.Constructor = c
	}
	ø.setFieldInfos()
	return
}

func Optional(name string, t Typer, html ...interface{}) (ø *Field) {
	e := h.NewElement(h.Tag("form"), h.WithoutDecoration)
	e.Add(html...)
	ø = &Field{Name: name, Type: t.Type(), Element: e, Required: false}
	if c, ok := t.(Constructor); ok {
		ø.Constructor = c
	}
	ø.setFieldInfos()
	return
}
