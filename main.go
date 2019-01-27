package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt.go"
	"github.com/segmentio/terraform-docs/internal/pkg/doc"
	"github.com/segmentio/terraform-docs/internal/pkg/print"
	"github.com/segmentio/terraform-docs/internal/pkg/print/json"
	markdown_document "github.com/segmentio/terraform-docs/internal/pkg/print/markdown/document"
	markdown_table "github.com/segmentio/terraform-docs/internal/pkg/print/markdown/table"
	"github.com/segmentio/terraform-docs/internal/pkg/print/pretty"
	"github.com/segmentio/terraform-docs/internal/pkg/settings"
)

var version = "dev"

const usage = `
  Usage:
    terraform-docs [--no-required] [--no-sort | --sort-inputs-by-required] [--with-aggregate-type-defaults] [--follow-modules] [json | markdown | md] [document | table] <path>...
    terraform-docs -h | --help

  Examples:

    # View inputs and outputs
    $ terraform-docs ./my-module

    # View inputs and outputs for variables.tf and outputs.tf only
    $ terraform-docs variables.tf outputs.tf

    # Generate a JSON of inputs and outputs
    $ terraform-docs json ./my-module

    # Generate Markdown tables of inputs and outputs
    $ terraform-docs md ./my-module

    # Generate Markdown tables of inputs and outputs
    $ terraform-docs md table ./my-module

    # Generate Markdown document of inputs and outputs
    $ terraform-docs md document ./my-module

    # Generate Markdown tables of inputs and outputs, but don't print "Required" column
    $ terraform-docs --no-required md ./my-module

    # Generate Markdown tables of inputs and outputs for the given module and ../config.tf
    $ terraform-docs md ./my-module ../config.tf

    # Generate markdown tables of inputs, outputs and used local modules
    $ terraform-docs --follow-modules md ./my-stack

  Options:
    -h, --help                       show help information
    --no-required                    omit "Required" column when generating Markdown
    --no-sort                        omit sorted rendering of inputs and outputs
    --sort-inputs-by-required        sort inputs by name and prints required inputs first
    --with-aggregate-type-defaults   print default values of aggregate types
    --follow-modules                 follow local modules in stacks (ignored when selected output is JSON)
    --version                        print version

  Types of Markdown:
    document                         generate Markdown document of inputs and outputs
    table                            generate Markdown tables of inputs and outputs (default)

`

func main() {
	parser := &docopt.Parser{
		HelpHandler:   docopt.PrintHelpAndExit,
		OptionsFirst:  true,
		SkipHelpFlags: false,
	}

	args, err := parser.ParseArgs(usage, nil, version)
	if err != nil {
		log.Fatal(err)
	}

	paths := args["<path>"].([]string)

	document, err := doc.CreateFromPaths(paths)
	if err != nil {
		log.Fatal(err)
	}

	var printSettings settings.Settings
	if !args["--no-required"].(bool) {
		printSettings.Add(print.WithRequired)
	}

	if !args["--no-sort"].(bool) {
		printSettings.Add(print.WithSortByName)
	}

	if args["--sort-inputs-by-required"].(bool) {
		printSettings.Add(print.WithSortInputsByRequired)
	}

	if args["--with-aggregate-type-defaults"].(bool) {
		printSettings.Add(print.WithAggregateTypeDefaults)
	}

	// construct the final output from multiple sources
	var out strings.Builder
	// most functions have double output (string, err)
	// which can not be used as input directly
	var tempstring string

	// get the main output (formatted)
	tempstring, err = doPrint(args, document, printSettings)

	if err != nil {
		log.Fatal(err)
	}
	// add the formatted document to the output string
	out.WriteString(tempstring)

	// done with the standard stuff, modules follow
	// no chance to use JSON as the logic of the program had to be changed
	if args["--follow-modules"].(bool) && !args["json"].(bool) && document.HasModules() {

		for _, module := range document.Modules {
			paths := []string{module.Source}
			isExternal := false

			modulepath := filepath.Join(module.GetBasepath(), paths[0])
			if _, err := os.Stat(modulepath); os.IsNotExist(err) {
				// the path does not exists, so the module will be either
				// git based or registry based and can not be loaded
				modulepath = paths[0]
				isExternal = true
			}

			document, err := doc.CreateFromPaths([]string{modulepath})
			if err != nil {
				log.Fatal(err)
			}

			tempstring, err = doPrint(args, document, printSettings)

			// print the Module name as header
			switch {
			case args["markdown"].(bool):
				fallthrough
			case args["md"].(bool):
				out.WriteString(fmt.Sprintf("\n----\n# Module: %s\n\n", module.Name))
				if isExternal {
					out.WriteString(fmt.Sprintf("external module\n"))
				}
			default:
				format := "\n\033[4m\033[1mModule:\033[21m %s\033[0m\n"
				out.WriteString(fmt.Sprintf(format, module.Name))
				if isExternal {
					format := "external module (%s)\n"
					out.WriteString(fmt.Sprintf(format, module.Source))
				}
			}

			out.WriteString(tempstring)
		}
	}

	// finally print the result
	fmt.Println(out.String())

}

// helper function to save code on switch()
func doPrint(args map[string]interface{}, document *doc.Doc, printSettings settings.Settings) (string, error) {
	switch {
	case args["markdown"].(bool), args["md"].(bool):
		if args["document"].(bool) {
			return markdown_document.Print(document, printSettings)
		} else {
			return markdown_table.Print(document, printSettings)
		}
	case args["json"].(bool):
		return json.Print(document, printSettings)
	default:
		return pretty.Print(document, printSettings)
	}
}
