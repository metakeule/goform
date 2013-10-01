package goform

import (
	"fmt"
	h "github.com/metakeule/goh4"
	. "github.com/metakeule/goh4/tag"
	"github.com/metakeule/pgsql"
	"github.com/metakeule/typeconverter"
	"time"
	// "html"
)

type TableForm struct {
	*FormHandler
	afterCreation []func(*TableForm)
}

func getType(in pgsql.Type) (out Type) {
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

func GetElement(in *pgsql.Field) (out *h.Element) {
	if in.Selection != nil {
		return SELECT()
	}
	if pgsql.IsVarChar(in.Type) {
		return INPUT(h.Attr("type", "text"))
	}

	if in.Type == pgsql.BoolType {
		return SELECT(
			OPTION(h.Attr("value", "true"), "true"),
			OPTION(h.Attr("value", "false"), "false"),
		)
	}

	switch in.Type {
	case pgsql.TextType:
		return TEXTAREA()
	case pgsql.XmlType:
		return TEXTAREA(h.Class("xml"))
	case pgsql.HtmlType:
		return TEXTAREA(h.Class("html"))
	case pgsql.IntType:
		return INPUT(h.Attr("type", "number"))
	case pgsql.UuidType:
		if in.ForeignKey != nil {
			return INPUT(h.Attr("type", "text", "fkey", in.ForeignKey.Table.Name), h.Class("foreign-key"))
		}
		return INPUT(h.Attr("type", "text"))
	case pgsql.DateType:
		return INPUT(h.Attr("type", "text"), h.Class("date"))
	case pgsql.TimeType:
		return INPUT(h.Attr("type", "time"))
	}
	return INPUT(h.Attr("type", "text"))
}

func TableField(f *pgsql.Field, e *h.Element) (field *Field) {
	if f.Is(pgsql.NullAllowed) {
		field = Optional(f.Name, getType(f.Type), LABEL(e))
	} else {
		field = Required(f.Name, getType(f.Type), LABEL(e))
	}
	return
}

func NewTableForm(fields []*pgsql.Field) (ø *TableForm) {
	ø = &TableForm{afterCreation: []func(*TableForm){}}
	ø.FormHandler = NewForm()
	for _, f := range fields {
		var field *Field
		if f.Is(pgsql.NullAllowed) {
			field = Optional(
				f.Name,
				getType(f.Type),
				LABEL(ø.getElement(f)),
			)
		} else {
			field = Required(
				f.Name,
				getType(f.Type),
				LABEL(ø.getElement(f)),
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
			trueOpt := elem.Any(h.And_(h.Attr("value", "true"), h.Tag("option")))
			if trueOpt != nil {
				trueOpt.SetContent(trueText)
			}
			falseOpt := elem.Any(h.And_(h.Attr("value", "false"), h.Tag("option")))
			if falseOpt != nil {
				falseOpt.SetContent(falseText)
			}
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
			option := elem.Any(h.And_(h.Attr("value", v), h.Tag("option")))
			if option != nil {
				option.Add(h.Attr("selected", "selected"))
			}
		} else {
			if elem.Tag() == "textarea" {
				elem.Add(v)
			} else {
				//tp := elem.Attribute("type")
				//if tp == "date" {
				if elem.HasClass("date") {
					var tme time.Time
					field := row.Table.Field(k)
					row.Get(field, &tme)
					year, month, day := tme.Date()
					// %02.0f.%02.0f.%4.0f
					v = fmt.Sprintf("%4.0f-%02.0f-%02.0f", float64(year), float64(int(month)), float64(day))
				}
				elem.Add(h.Attr("value", v))
			}
		}
	}
}

func (ø *TableForm) SetSaveAction(row *pgsql.Row, id string) {
	ø.Action = func(f *FormHandler) (err error) {
		err = row.Fill(f.Map())
		if err != nil {
			return err
		}
		if id != "new" && len(row.Table.PrimaryKey) == 1 {
			row.Set(row.Table.PrimaryKey[0], id)
		}

		err = row.Save()
		return
	}
}

func (ø *TableForm) SetLabels(o ...string) {
	for i := 0; i < len(o); i = i + 2 {
		ø.Label(o[i]).AddAtPosition(0, SPAN(o[i+1]))
	}
}

func (ø *TableForm) SetLabelMap(m map[string]string) {
	for k, v := range m {
		l := ø.Label(k)
		if l == nil {
			continue
		}
		l.AddAtPosition(0, SPAN(v))
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
		return SELECT()
	}
	if pgsql.IsVarChar(in.Type) {
		return INPUT(h.Attr("type", "text"))
	}

	if in.Type == pgsql.BoolType {
		return SELECT(
			OPTION(h.Attr("value", "true"), "true"),
			OPTION(h.Attr("value", "false"), "false"),
		)
	}

	switch in.Type {
	case pgsql.TextType:
		return TEXTAREA()
	case pgsql.XmlType:
		return TEXTAREA(h.Class("xml"))
	case pgsql.HtmlType:
		return TEXTAREA(h.Class("html"))
	case pgsql.UuidType:
		if in.ForeignKey != nil {
			return INPUT(h.Attr("type", "text", "fkey", in.ForeignKey.Table.Name), h.Class("foreign-key"))
		}
		return INPUT(h.Attr("type", "text"))
	case pgsql.IntType:
		return INPUT(h.Attr("type", "number"))
	case pgsql.DateType:
		return INPUT(h.Class("date"), h.Attr("type", "text"))
	case pgsql.TimeType:
		return INPUT(h.Attr("type", "time"))
	}
	return INPUT(h.Attr("type", "text"))

	/*
	 input[type="password"]:focus,
	 input[type="datetime"]:focus,
	 input[type="datetime-local"]:focus,
	 input[type="date"]:focus,
	 input[type="month"]:focus,
	 input[type="time"]:focus,
	 input[type="week"]:focus,
	 input[type="number"]:focus,
	 input[type="email"]:focus,
	 input[type="url"]:focus,
	 input[type="search"]:focus,
	 input[type="tel"]:focus,
	 input[type="color"]:focus,
	 .uneditable-input:focus
	*/

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
	e = ø.Any(h.And_(h.Tag("label"), h.Attr("for", fld)))
	return
}

func (ø *TableForm) Select(fld string) (e *h.Element) {
	e = ø.Any(h.And_(h.Tag("select"), h.Id(fld)))
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
		innerSelect := SELECT()
		label.SetContent(innerSelect)
		field.Element = label
		field.setFieldInfos()
		sel = innerSelect
	}
	sel.Clear()
	for _, v := range vals {
		r := ""
		typeconverter.Convert(v, &r)
		sel.Add(OPTION(h.Attr("value", r), h.Text(r)))
	}
}
