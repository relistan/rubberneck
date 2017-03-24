package rubberneck

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	expr *regexp.Regexp
)

const (
	AddLineFeed = iota
	NoAddLineFeed = iota
)

func init() {
	expr = regexp.MustCompile("^[a-z]")
}

// Conforms to the signature used by fmt.Printf and log.Printf among
// many functions available in other packages.
type printerFunc func(format string, v ...interface{})

// Printer defines the signature of a function that can be
// used to display the configuration. This signature is used
// by fmt.Printf, log.Printf, various logging output levels
// from the logrus package, and others.
type Printer struct {
	Show  printerFunc
}

func addLineFeed(fn printerFunc) printerFunc {
	return func(format string, v ...interface{}) {
		format = format + "\n"
		fn(format, v...)
	}
}

// NewDefaultPrinter returns a Printer configured to write to stdout.
func NewDefaultPrinter() *Printer {
	return &Printer{
		Show: func(format string, v ...interface{}) {
			fmt.Printf(format+"\n", v...)
		},
	}
}

// NewPrinter returns a Printer configured to use the supplied function
// to output to the supplied function.
func NewPrinter(fn printerFunc, lineFeed int) *Printer {
	p := &Printer{Show: fn}

	if lineFeed == AddLineFeed {
		p.Show = addLineFeed(fn)
	}

	return p
}

// Print attempts to pretty print the contents of obj in a format suitable
// for displaying the configuration of an application on startup.
func (p *Printer) Print(obj interface{}) {
	p.Show("Settings %s", strings.Repeat("=", 40))
	p.processOne(reflect.ValueOf(obj), 0)
	p.Show("%s", strings.Repeat("=", 49))
}

func (p *Printer) processOne(value reflect.Value, indent int) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	t := value.Type()

	for i := 0; i < value.NumField(); i++ {
		name := t.Field(i).Name

		// Other methods of detecting unexported fields seem unreliable
		// or different between Go versions and Go compilers (gc vs gccgo)
		if expr.MatchString(name) {
			continue
		}

		field := value.Field(i)
		realKind := field.Kind()

		if realKind == reflect.Ptr {
			realKind = field.Elem().Kind()
		}

		switch realKind {
		case reflect.Struct:
			p.Show("%s * %s:", strings.Repeat("  ", indent), name)
			p.processOne(reflect.ValueOf(field.Interface()), indent+1)
		default:
			p.Show("%s * %s: %v", strings.Repeat("  ", indent), name, field.Interface())
		}
	}
}

// Print configures a default printer to output to stdout and
// then prints the object.
func Print(obj interface{}) {
	NewDefaultPrinter().Print(obj)
}
