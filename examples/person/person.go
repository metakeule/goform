package person

import (
	. "github.com/metakeule/gopersona"
	. "github.com/metakeule/pgdb"
)

var OBJECTS = NewTable("objects", NewField("Id", IntType, PrimaryKey|Serial))

var ObjectId = NewField("ObjectId", IntType, OBJECTS.Field("Id"), OnDeleteCascade)
var Id = NewField("Id", IntType, PrimaryKey|Serial)
var FirstName = NewField("FirstName", VarChar(123), NullAllowed)
var LastName = NewField("LastName", VarChar(125))
var Age = NewField("Age", IntType)
var Vita = NewField("Vita", TextType, NullAllowed, Selection{"a", "b"})

var TABLE = NewTable("person", Id, FirstName, LastName, Age, Vita, ObjectId)

type ROW struct{ *Persona } // should include values map[*Field]interface{}

func New(db DB) *ROW                           { return &ROW{Persona: NewPersona(db, TABLE, Id, ObjectId)} }
func (ø *ROW) Id() int                         { return ø.GetInt(Id) }
func (ø *ROW) FirstName() string               { return ø.GetString(FirstName) }
func (ø *ROW) LastName() string                { return ø.GetString(LastName) }
func (ø *ROW) Age() int                        { return ø.GetInt(Age) }
func (ø *ROW) Vita() string                    { return ø.GetString(Vita) }
func (ø *ROW) SetFirstName(f *Field, s string) { ø.SetString(FirstName, s) }
func (ø *ROW) SetLastName(f *Field, s string)  { ø.SetString(LastName, s) }
func (ø *ROW) SetAge(f *Field, s int)          { ø.SetInt(Age, s) }
func (ø *ROW) SetVita(f *Field, s string)      { ø.SetString(Vita, s) }

// should be inherited: load all fields from the underlying table
//func (ø *ROW) Load(objectId int) (err error) { return ø.Load(objectId, FIELDS...) }
// should be inherited: insert/update  everything inside ø.values
// func (ø *ROW) Save() (err error)     { return ø.Persona.Save() }
// should be inherited
// func (ø *ROW) Validate() (err error) { return ø.Persona.Validate() }
