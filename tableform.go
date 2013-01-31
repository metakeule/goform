package goform

import (
	h "github.com/metakeule/goh4"
	"github.com/metakeule/pgdb"
	"github.com/metakeule/typeconverter"
)

type TableForm struct {
	*FormHandler
	afterCreation []func(*TableForm)
}

func NewTableForm(fields []*pgdb.Field, errorHandler func(error, []error, map[*Field][]error)) (ø *TableForm) {
	ø = &TableForm{afterCreation: []func(*TableForm){}}
	ø.FormHandler = NewForm(errorHandler)
	for _, f := range fields {
		var field *Field
		if f.Is(pgdb.NullAllowed) {
			field = Optional(
				f.Name,
				ø.getType(f.Type),
				h.Label(ø.getElement(f)),
			)
		} else {
			field = Required(
				f.Name,
				ø.getType(f.Type),
				h.Label(ø.getElement(f)),
			)
		}
		ø.AddField(field)
	}
	for _, fk := range ø.afterCreation {
		fk(ø)
	}
	return
}

func (ø *TableForm) getType(in pgdb.Type) (out Type) {
	switch in {
	case pgdb.IntType:
		return Int
	case pgdb.FloatType:
		return Float
	}
	return String
}

func (ø *TableForm) getElement(in *pgdb.Field) (out *h.Element) {
	if in.Selection != nil {
		ø.afterCreation = append(ø.afterCreation, func(tf *TableForm) {
			sell := []interface{}{}
			for _, ss := range in.Selection {
				sell = append(sell, ss)
			}
			tf.Selection(in.Name, sell...)
		})
		return h.Select()
	}
	if pgdb.IsVarChar(in.Type) {
		return h.Input()
	}
	if in.Type == pgdb.TextType {
		return h.Textarea()
	}

	return h.Input()
}

func (ø *TableForm) Unrequire(fld string) {
	e := ø.Label(fld)
	e.RemoveClass("required")
	ø.FieldElement(fld).RemoveAttribute("required")
	field := ø.Field(fld)
	if field.Required {
		ø.RemoveFieldFromRequired(field)
		field.Required = false
	}
}

func (ø *TableForm) Label(fld string) (e *h.Element) {
	_, e = ø.Any(h.And(h.Tag("label"), h.Attr("for", fld)))
	return
}

func (ø *TableForm) Select(fld string) (e *h.Element) {
	_, e = ø.Any(h.And(h.Tag("select"), h.Id(fld)))
	return
}

func (ø *TableForm) FieldElement(fld string) *h.Element {
	e := ø.Label(fld)
	return e.Fields()[0]
}

/*
pattern
placeholder
required als attribut ver
*/

func (ø *TableForm) Require(fld string) {
	e := ø.Label(fld)
	e.AddClass("required")
	ø.FieldElement(fld).Add(h.Attr("required", "required"))
	field := ø.Field(fld)
	if !field.Required {
		ø.AddFieldToRequired(field)
		field.Required = true
	}
}

func (ø *TableForm) Selection(fld string, vals ...interface{}) {
	field := ø.Field(fld)
	Selection(field, vals...)
	label := ø.Label(fld)
	sel := ø.FieldElement(fld)

	if sel.Tag() != "select" {
		innerSelect := h.Select()
		label.Set(innerSelect)
		field.Element = label
		field.setFieldInfos()
		sel = innerSelect
	}
	sel.Clear()
	for _, v := range vals {
		r := ""
		typeconverter.Convert(v, &r)
		sel.Add(h.Option(h.Attr("value", r), h.Text(r)))
	}
}
