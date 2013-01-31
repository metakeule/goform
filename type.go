package goform

const (
	_        = iota
	Int Type = 1 << iota
	String
	Float
	IntArray
	StringArray
	FloatArray
	Map
	Struct
	Fill
)

type Type int

func (ø Type) Type() Type { return ø }

type Typer interface {
	Type() Type
}

type Constructor func() interface{}

func (ø Constructor) Type() Type { return Struct }
