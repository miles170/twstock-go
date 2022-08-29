package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/golang-sql/civil"
	twstock "github.com/miles170/twstock-go/twstock"
)

func format(w *os.File, v reflect.Value) {
	switch v.Interface().(type) {
	case civil.Date:
		d := v.Interface().(civil.Date)
		fmt.Fprintf(w, "civil.Date{Year: %d, Month: %d, Day: %d},", d.Year, d.Month, d.Day)
	default:
		fmt.Fprintf(w, "\"%s\",", v)
	}
}

func main() {
	w, err := os.Create("/tmp/securities_GENERATED.go")
	if err != nil {
		log.Fatal(w)
	}
	defer w.Close()

	client := twstock.NewClient()
	securities, err := client.Security.Download()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "// Code generated security DO NOT EDIT.\n\n")
	fmt.Fprintf(w, "//go:build !codeanalysis\n\n")
	fmt.Fprintf(w, "package %s\n\n", "twstock")
	fmt.Fprint(w, "import \"github.com/golang-sql/civil\"\n\n")
	fmt.Fprintf(w, "var Securities = map[string]Security{\n")

	fields := reflect.VisibleFields(reflect.TypeOf(struct{ twstock.Security }{}))
	for _, s := range securities {
		fmt.Fprintf(w, "\t\"%s\":{", s.Code)
		v := reflect.ValueOf(s)
		for i := range fields {
			if i == 0 {
				continue
			}
			format(w, reflect.Indirect(v).Field(i-1))
		}
		fmt.Fprint(w, "},\n")
	}

	fmt.Fprint(w, "}\n")
}
