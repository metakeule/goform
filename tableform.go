package goform

import (
	h "github.com/metakeule/goh4"
	"github.com/metakeule/pgsql"
	"github.com/metakeule/typeconverter"
)

type TableForm struct {
	*FormHandler
	afterCreation []func(*TableForm)
}

func NewTableForm(fields []*pgsql.Field) (ø *TableForm) {
	ø = &TableForm{afterCreation: []func(*TableForm){}}
	ø.FormHandler = NewForm()
	for _, f := range fields {
		var field *Field
		if f.Is(pgsql.NullAllowed) {
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

// sets all true and false values to the given text
func (ø *TableForm) SetBoolTexts(trueText string, falseText string) {
	for n, f := range ø.Fields {
		if f.Type == Bool {
			elem := ø.FieldElement(n)
			elem.Any(h.And(h.Attr("value", "true"), h.Tag("option"))).Set(trueText)
			elem.Any(h.And(h.Attr("value", "false"), h.Tag("option"))).Set(falseText)
		}
	}
}

func (ø *TableForm) SetValues(row *pgsql.Row) {
	props := row.AsStrings()
	for k, v := range props {
		if !ø.HasFieldDefinition(k) {
			continue
		}
		elem := ø.FieldElement(k)
		if elem.Tag() == "select" {
			option := elem.Any(h.And(h.Attr("value", v), h.Tag("option")))
			if option != nil {
				option.Add(h.Attr("selected", "selected"))
			}
		} else {
			elem.Add(h.Attr("value", v))
		}
	}
}

func (ø *TableForm) SetSaveAction(row *pgsql.Row, id int) {
	ø.Action = func(f *FormHandler) (err error) {
		err = row.Fill(f.Map())
		if err != nil {
			return err
		}
		if id != 0 {
			row.Set(row.Table.PrimaryKey, id)
		}

		err = row.Save()
		return
	}
}

func (ø *TableForm) getType(in pgsql.Type) (out Type) {
	switch in {
	case pgsql.IntType:
		return Int
	case pgsql.FloatType:
		return Float
	case pgsql.BoolType:
		return Bool
	}
	return String
}

func (ø *TableForm) SetLabels(o ...string) {
	for i := 0; i < len(o); i = i + 2 {
		ø.Label(o[i]).Add(h.Span(o[i+1]))
	}
}

func (ø *TableForm) SetLabelMap(m map[string]string) {
	for k, v := range m {
		l := ø.Label(k)
		if l == nil {
			continue
		}
		l.Add(h.Span(v))
	}
}

func (ø *TableForm) getElement(in *pgsql.Field) (out *h.Element) {
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
	if pgsql.IsVarChar(in.Type) {
		return h.Input()
	}
	if in.Type == pgsql.TextType {
		return h.Textarea()
	}

	if in.Type == pgsql.BoolType {
		return h.Select(
			h.Option(h.Attr("value", "true"), "true"),
			h.Option(h.Attr("value", "false"), "false"),
		)
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
	e = ø.Any(h.And(h.Tag("label"), h.Attr("for", fld)))
	return
}

func (ø *TableForm) Select(fld string) (e *h.Element) {
	e = ø.Any(h.And(h.Tag("select"), h.Id(fld)))
	return
}

func (ø *TableForm) FieldElement(fld string) *h.Element {
	e := ø.Label(fld)
	return e.Fields()[0]
}

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
