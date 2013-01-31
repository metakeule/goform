package goform

import (
	"encoding/json"
	"fmt"
	h "github.com/metakeule/goh4"
	"strconv"
	"strings"
)

type FormHandler struct {
	*h.Element
	Ints          map[*Field]int
	Floats        map[*Field]float32
	Strings       map[*Field]string
	IntArrays     map[*Field][]int
	StringArrays  map[*Field][]string
	FloatArrays   map[*Field][]float32
	JsonMaps      map[*Field]map[string]interface{}
	JsonStructs   map[*Field]interface{}
	JsonsOriginal map[*Field]string
	Fills         map[*Field]Filler
	Types         map[*Field]Type
	Fields        map[string]*Field
	errorHandler  func(error, []error, map[*Field][]error)
	Order         []h.Stringer
	required      []*Field
	Validation    func(*FormHandler)       // may call AddFieldError and AddValidationError
	Action        func(*FormHandler) error // stops on the first error and returns it

	BeforeParsing func(*FormHandler) // will be executed before Parsing
	AfterParsing  func(*FormHandler) // will be executed after Parsing

	BeforeValidation func(*FormHandler) // will be executed before Validation
	AfterValidation  func(*FormHandler) // will be executed after Validation

	BeforeAction func(*FormHandler) // will be executed before Action
	AfterAction  func(*FormHandler) // will be executed after Action

	FieldErrors             map[*Field][]error // to collect all the errors of the Fields
	GeneralValidationErrors []error            // to collect all validation errors that are a result of different Field values
}

func (ø *FormHandler) resetElement() {
	ø.Element = h.Form()
	for _, s := range ø.Order {
		if f, ok := s.(*Field); ok {
			ø.Element.Add(f.Element)
		} else {
			ø.Element.Add(s)
		}
	}
}

func (ø *FormHandler) Reset() {
	ø.Ints = map[*Field]int{}
	ø.Floats = map[*Field]float32{}
	ø.Strings = map[*Field]string{}
	ø.IntArrays = map[*Field][]int{}
	ø.StringArrays = map[*Field][]string{}
	ø.FloatArrays = map[*Field][]float32{}
	ø.JsonMaps = map[*Field]map[string]interface{}{}
	ø.FieldErrors = map[*Field][]error{}
	ø.JsonStructs = map[*Field]interface{}{}
	ø.JsonsOriginal = map[*Field]string{}
	ø.GeneralValidationErrors = []error{}
	ø.Fills = map[*Field]Filler{}
}

func (ø *FormHandler) AddFieldError(Field *Field, err error) {
	if ø.FieldErrors[Field] == nil {
		ø.FieldErrors[Field] = []error{err}
	} else {
		ø.FieldErrors[Field] = append(ø.FieldErrors[Field], err)
	}
}

func (ø *FormHandler) AddValidationError(err error) {
	ø.GeneralValidationErrors = append(ø.GeneralValidationErrors, err)
}

func (ø *FormHandler) IsNil(field *Field) (is bool) {
	is = false
	switch field.Type {
	case Int:
		if ø.Ints[field] == 0 {
			return true
		}
	case String:
		if ø.Strings[field] == "" {
			return true
		}
	case Float:
		if ø.Floats[field] == 0.0 {
			return true
		}
	case IntArray:
		if ø.IntArrays[field] == nil {
			return true
		}
	case FloatArray:
		if ø.FloatArrays[field] == nil {
			return true
		}
	case StringArray:
		if ø.StringArrays[field] == nil {
			return true
		}
	case Struct:
		if ø.JsonsOriginal[field] == "" {
			return true
		}
	case Map:
		if ø.JsonsOriginal[field] == "" {
			return true
		}
	case Fill:
		if ø.Fills[field] == nil {
			return true
		}
	}
	return
}

func (ø *FormHandler) removeFieldFromOrder(f *Field) {
	newOrder := []h.Stringer{}
	for _, s := range ø.Order {
		if fld, ok := s.(*Field); ok && fld == f {
			continue
		}
		newOrder = append(newOrder, s)
	}
	ø.Order = newOrder
}

func (ø *FormHandler) RemoveFieldFromRequired(f *Field) {
	newRequired := []*Field{}
	for _, fld := range ø.required {
		if fld == f {
			continue
		}
		newRequired = append(newRequired, fld)
	}
	ø.required = newRequired
}

func (ø *FormHandler) AddFieldToRequired(f *Field) {
	ø.required = append(ø.required, f)
}

func (ø *FormHandler) RemoveField(fld string) {
	field := ø.Field(fld)
	switch field.Type {
	case Int:
		delete(ø.Ints, field)
	case String:
		delete(ø.Strings, field)
	case Float:
		delete(ø.Floats, field)
	case IntArray:
		delete(ø.IntArrays, field)
	case FloatArray:
		delete(ø.FloatArrays, field)
	case StringArray:
		delete(ø.StringArrays, field)
	case Struct:
		delete(ø.JsonStructs, field)
		delete(ø.JsonsOriginal, field)
	case Map:
		delete(ø.JsonMaps, field)
		delete(ø.JsonsOriginal, field)
	case Fill:
		delete(ø.Fills, field)
		delete(ø.JsonsOriginal, field)
	}
	ø.removeFieldFromOrder(field)
	if field.Required {
		ø.RemoveFieldFromRequired(field)
	}
	ø.resetElement()
	delete(ø.FieldErrors, field)
	delete(ø.Fields, fld)
}

func (ø *FormHandler) Validate() {
	for _, field := range ø.required {
		if ø.IsNil(field) {
			ø.AddFieldError(field, fmt.Errorf("required"))
		}
	}

	for _, field := range ø.Fields {
		field.CheckAllowed(ø)
	}

	if ø.Validation != nil {
		ø.Validation(ø)
	}

	for field, fill := range ø.Fills {
		if err := fill.Validate(); err != nil {
			ø.AddFieldError(field, err)
		}
	}

}

func (ø *FormHandler) Parse(vals map[string]string) (err error) {
	if ø.BeforeParsing != nil {
		ø.BeforeParsing(ø)
	}

	for kk, v := range vals {
		if ø.Fields[kk] == nil {
			continue
		}

		k := ø.Fields[kk]

		switch ø.Types[k] {
		case Int:
			i, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				ø.AddFieldError(k, fmt.Errorf("%#v is no int", v))
			}
			ø.Ints[k] = int(i)
		case String:
			ø.Strings[k] = v
		case Float:
			fl, err := strconv.ParseFloat(v, 32)
			if err != nil {
				ø.AddFieldError(k, fmt.Errorf("%#v is no float", v))
			}
			ø.Floats[k] = float32(fl)
		case IntArray:
			m := []int{}
			a := strings.Split(v, ",")
			for _, str := range a {
				trimmed := strings.Trim(str, " ")
				i, err := strconv.ParseInt(trimmed, 0, 32)
				if err != nil {
					ø.AddFieldError(k, fmt.Errorf("%#v is no int", str))
				}
				m = append(m, int(i))
			}
			ø.IntArrays[k] = m
		case FloatArray:
			m := []float32{}
			a := strings.Split(v, ",")
			for _, str := range a {
				trimmed := strings.Trim(str, " ")
				i, err := strconv.ParseFloat(trimmed, 32)
				if err != nil {
					ø.AddFieldError(k, fmt.Errorf("%#v is no float", str))
				}
				m = append(m, float32(i))
			}
			ø.FloatArrays[k] = m

		case StringArray:
			m := []string{}
			a := strings.Split(v, ",")
			for _, str := range a {
				i := strings.Trim(str, " ")
				m = append(m, i)
			}
			ø.StringArrays[k] = m
		case Struct:
			ø.JsonsOriginal[k] = v
			ø.JsonStructs[k] = k.Constructor()
			i := ø.JsonStructs[k]

			dec := json.NewDecoder(strings.NewReader(v))
			err = dec.Decode(i)
			if err != nil {
				ø.AddFieldError(k, fmt.Errorf("%#v could not be parsed: %s", v, err))
			}
		case Map:
			ø.JsonsOriginal[k] = v
			var ii map[string]interface{}
			err = json.Unmarshal([]byte(v), &ii)
			ø.JsonMaps[k] = ii
			if err != nil {
				ø.AddFieldError(k, fmt.Errorf("%#v could not be parsed: %s", v, err))
			}
		case Fill:
			ø.JsonsOriginal[k] = v
			var ii map[string]interface{}
			err = json.Unmarshal([]byte(v), &ii)
			if err != nil {
				ø.AddFieldError(k, fmt.Errorf("%#v could not be parsed: %s", v, err))
			}
			ø.Fills[k].Fill(ii)
		}
	}
	if ø.AfterParsing != nil {
		ø.AfterParsing(ø)
	}

	if ø.BeforeValidation != nil {
		ø.BeforeValidation(ø)
	}

	ø.Validate()

	if ø.AfterValidation != nil {
		ø.AfterValidation(ø)
	}

	if len(ø.FieldErrors) == 0 && len(ø.GeneralValidationErrors) == 0 {
		if ø.Action != nil {
			if ø.BeforeAction != nil {
				ø.BeforeAction(ø)
			}

			err = ø.Action(ø)
			if err == nil && ø.AfterAction != nil {
				ø.AfterAction(ø)
			}
		}
		return
	}

	if len(ø.FieldErrors) > 0 || len(ø.GeneralValidationErrors) > 0 || err != nil {
		ø.errorHandler(err, ø.GeneralValidationErrors, ø.FieldErrors)
	}

	if len(ø.FieldErrors) > 0 && len(ø.GeneralValidationErrors) > 0 {
		err = fmt.Errorf("Field errors and general validation errors")
		return
	}

	if len(ø.FieldErrors) > 0 {
		err = fmt.Errorf("Field errors")
		return
	}

	if len(ø.GeneralValidationErrors) > 0 {
		err = fmt.Errorf("general validation errors")
	}

	return
}

func (ø *FormHandler) HasFieldDefinition(field string) (has bool) {
	has = false
	for _, f := range ø.Fields {
		if f.Name == field {
			return true
		}
	}
	return
}

func (ø *FormHandler) AddHtml(s h.Stringer) {
	ø.Order = append(ø.Order, s)
	ø.Element.Add(s)
}

func (ø *FormHandler) AddField(f *Field) {
	if ø.HasFieldDefinition(f.Name) {
		panic("Field " + f.Name + " already defined")
	}
	ø.Order = append(ø.Order, f)
	ø.Types[f] = f.Type
	ø.Fields[f.Name] = f
	if f.Required {
		ø.required = append(ø.required, f)
	}
	ø.Element.Add(f.Element)
}

func (ø *FormHandler) Field(fld string) (f *Field) {
	return ø.Fields[fld]
}

func (ø *FormHandler) Get(field string) interface{} {
	if ø.Fields[field] == nil {
		panic("field " + field + " does not exist")
	}
	k := ø.Fields[field]
	switch ø.Types[k] {
	case Int:
		return ø.Ints[k]
	case String:
		return ø.Strings[k]
	case Float:
		return ø.Floats[k]
	case IntArray:
		return ø.IntArrays[k]
	case FloatArray:
		return ø.FloatArrays[k]
	case StringArray:
		return ø.StringArrays[k]
	case Struct:
		return ø.JsonStructs[k]
	case Map:
		return ø.JsonMaps[k]
	case Fill:
		return ø.Fills[k]
	}
	panic("can't get field " + field + ": unknown type")
}