// Package main implements a parser for our config structure
// that emits help topics based on its struct tags.
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"core"
)

type output struct {
	Preamble string            `json:"preamble"`
	Topics   map[string]string `json:"topics"`
}

// ExampleValue returns an example value for a config field based on its type.
func ExampleValue(f reflect.Value, name string, t reflect.Type) string {
	if t.Kind() == reflect.Slice {
		return ExampleValue(f, name, t.Elem()) + fmt.Sprintf("\n\n%s can be repeated", name)
	}
	// Special case some fields of unusual types.
	if name == "version" {
		return core.PleaseVersion.String() // keep it up to date!
	}
	if t.Kind() == reflect.String {
		if f.String() != "" {
			return f.String()
		}
		return "<str>"
	} else if t.Kind() == reflect.Bool {
		return "true | false | yes | no | on | off"
	} else if t.Kind() == reflect.Int || t.Kind() == reflect.Int64 {
		if f.Int() != 0 {
			return fmt.Sprintf("%d", f.Int())
		}
		return "42"
	} else if t.Kind() == reflect.Uint64 {
		return fmt.Sprintf("%d", f.Uint())
	} else if t.Name() == "BuildLabel" {
		return "//src/core:core"
	}
	panic(fmt.Sprintf("Unknown type: %s", t.Kind()))
}

func main() {
	o := output{
		Preamble: "%s is a config setting defined in the .plzconfig file. See `plz help plzconfig` for more information.",
		Topics:   map[string]string{},
	}
	config := core.DefaultConfiguration()
	v := reflect.ValueOf(*config)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		sectname := strings.ToLower(t.Field(i).Name)
		if f.Type().Kind() == reflect.Struct {
			for j := 0; j < f.Type().NumField(); j++ {
				subf := f.Field(j)
				subt := t.Field(i).Type.Field(j)
				if help := subt.Tag.Get("help"); help != "" {
					name := strings.ToLower(subt.Name)
					preamble := fmt.Sprintf("[%s]\n%s = %s\n\n", sectname, name, ExampleValue(subf, name, subt.Type))
					help = strings.Replace(help, "\\n", "\n", -1)
					o.Topics[name] = preamble + help
				}
			}
		}
	}
	b, _ := json.Marshal(o)
	fmt.Println(string(b))
}