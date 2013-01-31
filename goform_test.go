package goform

import (
	"fmt"
	"github.com/metakeule/goform/examples/person"
	"testing"
)

func err(t *testing.T, msg string, is interface{}, shouldbe interface{}) {
	t.Errorf(msg+": is %s, should be %s\n", is, shouldbe)
}

func TestT(t *testing.T) {
	fmt.Println("just a test")
}

func handleErrors(err error, generalValidationErrs []error, fieldErrors map[*Field][]error) {
	if len(fieldErrors) > 0 {
		for k, ee := range fieldErrors {
			fmt.Printf("\nErrors in field %s\n", k.Name)
			for _, eee := range ee {
				fmt.Printf("\t- %s\n", eee)
			}
		}
	}

	if len(generalValidationErrs) > 0 {
		for _, ee := range generalValidationErrs {
			fmt.Printf("\nValidation Error %s\n", ee)
		}
	}

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func TestNewTableForm(t *testing.T) {
	// if you don't want all the field, take only some
	f := NewTableForm(person.TABLE.Fields, handleErrors)
	f.RemoveField("Id")
	f.Unrequire("LastName")
	/*


		f.Require("firstname")
		f.Label("firstname").AddAtPosition(0, h.Span(h.Text("Your First Name:")))
		f.Label("lastname").Add(h.Span(h.Text("Your Last Name")))
		f.FieldElement("firstname").Add(h.Attr("value", "Benny"))

		f.Select("vita").AddClass("vitesse")

		for _, option := range f.FieldElement("vita").All(h.Tag("option")) {
			option.Set(h.Text("option " + option.Attributes()["value"]))
		}

		//f.Selection("vita", "c", "d", "e")

		f.Selection("object_id", 1, 2, 3, 200)
	*/
	vals := map[string]string{
		"Id":        "12",
		"FirstName": "Donald",
		"LastName":  "Duck",
		"Age":       "144",
		"ObjectId":  "200",
		"Vita":      "a",
	}

	fmt.Println(f)

	_ = f.Parse(vals)

	//fmt.Println(Form(Id("testform"), Attr("action", "blah"), f))

}
