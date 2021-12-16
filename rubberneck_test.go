package rubberneck_test

import (
	"fmt"

	. "github.com/relistan/rubberneck"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rubberneck", func() {
	Describe("NewDefaultPrinter()", func() {
		It("returns a properly configured Printer", func() {
			printer := NewDefaultPrinter()
			Expect(printer.Show).NotTo(BeNil())
			Expect(printer.Mask).To(BeNil())
		})
	})

	Describe("NewPrinter()", func() {
		var didRun bool
		var didAddLineFeed bool

		var printFunc func(format string, v ...interface{})
		var printable struct{ Content string }
		var receivedFormat string

		// This ridiculous line is required to get Go to compile tests.
		// Without it, it says receivedFormat is not used. If I remove
		// receivedFormat it complains about it in the BeforeEach block,
		// because it's _obviously_ not defined.
		if receivedFormat == "" { receivedFormat = "" }

		BeforeEach(func() {
			didRun = false
			didAddLineFeed = false

			printFunc = func(format string, v ...interface{}) {
				didRun = true
				if format[len(format)-1] == '\n' {
					didAddLineFeed = true
				}
				receivedFormat = format
			}

			printable = struct{ Content string }{"grendel"}
		})

		It("returns a properly configured Printer without line feed", func() {
			printer := NewPrinter(printFunc, NoAddLineFeed)
			Expect(printer.Show).NotTo(BeNil())

			printer.Print(printable)
			Expect(didRun).To(BeTrue())
			Expect(didAddLineFeed).To(BeFalse())
		})

		It("returns a properly configured Printer with line feed", func() {
			printer := NewPrinter(printFunc, AddLineFeed)
			Expect(printer.Show).NotTo(BeNil())

			printer.Print(printable)
			Expect(didRun).To(BeTrue())
			Expect(didAddLineFeed).To(BeTrue())
		})
	})

	Describe("NewPrinterWithKeyMasking()", func() {
		var didPrint bool
		var didMask bool

		var printFunc func(format string, v ...interface{})
		var maskFunc func(argument string) *string
		var printable struct{ Content string }
		var printedValue = ""

		BeforeEach(func() {
			didPrint = false
			didMask = false

			printFunc = func(format string, v ...interface{}) {
				didPrint = true
				printedValue += fmt.Sprintf(format, v...)
			}
			maskFunc = func(argument string) *string {
				if argument == "Content" {
					didMask = true
					mask := "MASKED"
					return &mask
				}
				return nil
			}

			printable = struct{ Content string }{"grendel"}
		})

		It("returns a properly configured Printer", func() {
			printer := NewPrinterWithKeyMasking(printFunc, maskFunc, AddLineFeed)
			Expect(printer.Show).NotTo(BeNil())
			Expect(printer.Mask).NotTo(BeNil())

			printer.Print(printable)
			Expect(didPrint).To(BeTrue())
			Expect(didMask).To(BeTrue())
			Expect(printedValue).NotTo(HaveLen(0))
			Expect(printedValue).To(ContainSubstring("MASKED"))
		})
	})

	Describe("when printing with", func() {
		var printFunc func(format string, v ...interface{})
		var printable struct {
			Content []string
			Another struct{ Included string }
			private bool
		}
		var output string
		var printer *Printer

		BeforeEach(func() {
			output = ""

			printFunc = func(format string, v ...interface{}) {
				output += fmt.Sprintf(format, v...)
			}

			printable = struct {
				Content []string
				Another struct{ Included string }
				private bool
			}{
				[]string{"njal", "groenlendinga"},
				struct{ Included string }{"leif"},
				true,
			}

			printer = NewPrinter(printFunc, AddLineFeed)
		})

		Describe("PrintWithLabel()", func() {
			It("generates correct output", func() {
				printer.PrintWithLabel("saga", printable)
				Expect(output).To(ContainSubstring("saga ----"))
				Expect(output).To(ContainSubstring("Content: [njal groenlendinga]"))
				Expect(output).To(MatchRegexp("\\* Another:\n\\s+\\* Included: leif"))
			})

			It("generates correct output when passed a pointer", func() {
				printer.PrintWithLabel("saga", &printable)
				Expect(output).To(ContainSubstring("saga ----"))
				Expect(output).To(ContainSubstring("Content: [njal groenlendinga]"))
				Expect(output).To(MatchRegexp("\\* Another:\n\\s+\\* Included: leif"))
			})

			It("excludes private struct members", func() {
				printer.PrintWithLabel("saga", &printable)
				Expect(output).NotTo(ContainSubstring("private"))
			})

			It("handles labels longer than 50", func() {
				printer.PrintWithLabel("0123456789012345678901234567890123456789012345678901234567891234", &printable)
				Expect(output).To(ContainSubstring("------------------------------------------------------"))
			})
		})

		Describe("Print()", func() {
			It("complains when passed a string", func() {
				printer.Print("saga", printable)
				Expect(output).To(ContainSubstring("Expected to print a struct"))
				Expect(output).NotTo(ContainSubstring("Content: [njal groenlendinga]"))
				Expect(output).NotTo(MatchRegexp("\\* Another:\n\\s+\\* Included: leif"))
			})
		})
	})

	Describe("handling values", func() {
		var printFunc func(format string, v ...interface{})
		var printable struct {
			Bad *int
		}
		var output string
		var printer *Printer

		BeforeEach(func() {
			output = ""

			printFunc = func(format string, v ...interface{}) {
				output += fmt.Sprintf(format, v...)
			}

			printable = struct {
				Bad *int
			}{}

			printer = NewPrinter(printFunc, AddLineFeed)
		})

		It("handles nil pointers", func() {
			test := func() { printer.Print(printable) }

			Expect(test).NotTo(Panic())
		})
	})
})
