package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	twstock "github.com/miles170/twstock-go/twstock"
)

func main() {
	w, err := os.Create("./twstock/securities_GENERATED.go")
	if err != nil {
		log.Fatal(w)
	}
	defer w.Close()

	client := twstock.NewClient()
	securities, err := client.Security.Download()
	if err != nil {
		log.Fatal(w)
	}

	fmt.Fprintf(w, "// Code generated security DO NOT EDIT.\n\n")
	fmt.Fprintf(w, "//go:build !codeanalysis\n\n")
	fmt.Fprintf(w, "package %s\n\n", "twstock")
	fmt.Fprintf(w, "var Securities = map[string]Security{\n")

	fields := reflect.VisibleFields(reflect.TypeOf(struct{ twstock.Security }{}))
	for _, s := range securities {
		fmt.Fprintf(w, "\t\"%s\":{", s.Code)
		v := reflect.ValueOf(s)
		for i := range fields {
			if i == 0 {
				continue
			}
			fmt.Fprintf(w, "\"%s\",", reflect.Indirect(v).Field(i-1))
		}
		fmt.Fprint(w, "},\n")
	}

	fmt.Fprint(w, "}\n")
}
