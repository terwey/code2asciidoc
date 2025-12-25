package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type docOutput struct {
	title        string
	funcName     string
	body         []string
	path         string
	goFilename   string
	jsonFilename string
	apidocs      bool
	postTitle    string
	post         []string
}

func newDocOutput(filePath string, goFilename string) docOutput {
	base := strings.Split(goFilename, "_test.go")
	return docOutput{
		path:         filePath,
		goFilename:   goFilename,
		jsonFilename: fmt.Sprintf("%s.apisamples", base[0]),
	}
}

func (d *docOutput) getGoInclude() string {
	includePath := d.getIncludePath(d.goFilename)
	return fmt.Sprintf("include::%s[tag=%s,indent=0]", includePath, d.funcName)
}

func (d *docOutput) getJsonInclude() string {
	includePath := d.getIncludePath(d.jsonFilename)
	return fmt.Sprintf("include::%s[tag=%s]", includePath, d.funcName)
}

func (d *docOutput) getIncludePath(filename string) string {
	fullPath := path.Join(d.path, filename)

	// Antora mode: use example$ prefix with just the filename
	if antora {
		return fmt.Sprintf("example$%s", filename)
	}

	// Include prefix mode: prepend custom prefix to filename
	if includePrefix != "" {
		return fmt.Sprintf("%s%s", includePrefix, filename)
	}

	// Relative path mode: compute relative path from specified directory
	if relativeTo != "" {
		relPath, err := filepath.Rel(relativeTo, fullPath)
		if err != nil {
			// If we can't compute relative path, fall back to absolute
			return fullPath
		}
		return relPath
	}

	// Default: absolute path (original behavior)
	return fullPath
}

func (d *docOutput) functionName() string {
	return fmt.Sprintf("Test_%s", d.funcName)
}

func (d *docOutput) processPost() {
	var body []string
	found := false
	end := false
	for _, line := range d.body {
		if strings.Contains(line, "startpostdocs") {
			t := strings.TrimPrefix(line, "startpostdocs")
			d.postTitle = strings.Trim(t, " ")
			found = true
			continue
		}
		if strings.Contains(line, "endpostdocs") {
			end = true
			continue
		}
		if found && !end {
			d.post = append(d.post, line)
			continue
		}

		body = append(body, line)
	}
	d.body = body
}

func (d *docOutput) getAsciiDoc() string {
	d.processPost()
	var out strings.Builder

	// Outer tag wrapper (optional)
	if !noOuterTags {
		out.WriteString("// tag::" + d.funcName + "[]\n")
	}

	// Page break (optional)
	if !noPageBreaks {
		out.WriteString("<<<\n")
	}

	// Section heading (optional)
	if !noHeadings {
		out.WriteString("== ")
		out.WriteString(d.title)
		out.WriteString("\n")
	}

	// Documentation body
	for _, line := range d.body {
		out.WriteString(line)
		out.WriteString("\n")
	}

	// Go code section
	out.WriteString("\n[#" + strings.ToLower(d.title) + "_" + strings.ToLower(d.funcName) + "_go]")
	out.WriteString("\n.Go ")
	out.WriteString(d.title)
	// out.WriteString("\n[%collapsible]\n")
	// out.WriteString("====\n")
	out.WriteString("\n[source,go]\n")
	out.WriteString("----\n")
	out.WriteString(d.getGoInclude())
	out.WriteString("\n----\n")
	// out.WriteString("====\n")

	// JSON section (optional - only if apidocs is set AND skipJSON is false)
	if d.apidocs && !skipJSON {
		out.WriteString("\n[#" + strings.ToLower(d.title) + "_" + strings.ToLower(d.funcName) + "_json]")
		out.WriteString("\n.JSON ")
		out.WriteString(d.title)
		// out.WriteString("\n[%collapsible]\n")
		// out.WriteString("====\n")
		out.WriteString("\n[source,json]\n")
		out.WriteString("----\n")
		out.WriteString(d.getJsonInclude())
		out.WriteString("\n----\n")
		// out.WriteString("====\n")
	}

	// Post-documentation section
	if d.postTitle != `` {
		out.WriteString("\n=== ")
		out.WriteString(d.postTitle)
		out.WriteString("\n")
		for _, line := range d.post {
			out.WriteString(line)
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	// Outer tag wrapper end (optional)
	if !noOuterTags {
		out.WriteString("// end::" + d.funcName + "[]\n")
	}

	return out.String()
}

var (
	flags         *pflag.FlagSet
	sourceFile    string
	outFile       string
	overwrite     bool
	runTests      bool
	noheader      bool
	skipJSON      bool
	antora        bool
	noPageBreaks  bool
	noHeadings    bool
	noOuterTags   bool
	relativeTo    string
	includePrefix string
	dryRun        bool
)

func init() {
	/*
		Used by documentation for the manpage
		tag::options[]
		*--source string*
			Source file to parse into AsciiDoc, recommended is to set the absolute path.

		*--out*
			File to write to, if left empty writes to stdout

		*--f*
			Overwrite the existing out file

		*--run*
			Run the tests to produce the output file for the JSON samples.
			The JSON samples need to be written to a file called the same as
			the source Go file minus the _test with a .apisamples extension

		*--no-header*
			Do not set a document header and ToC

		*--skip-json*
			Skip JSON sample sections in output (useful when .apisamples file doesn't exist)

		*--antora*
			Use Antora-compatible include paths (example$ prefix instead of absolute paths)

		*--no-page-breaks*
			Do not insert page break markers (<<<) before sections

		*--no-headings*
			Do not generate section headings from tag names

		*--no-outer-tags*
			Do not wrap generated content in outer tag markers

		*--relative-to string*
			Make include paths relative to the specified directory

		*--include-prefix string*
			Prepend a prefix to all include paths (e.g., 'example$')

		*--dry-run*
			Preview output to stdout without writing files
		end::options[]
	*/
	flags = pflag.NewFlagSet("AsciiDoc Generator for Distrib", pflag.ContinueOnError)
	flags.StringVar(&sourceFile, "source", "", "Source file to parse into AsciiDoc, recommended is to set the absolute path.")
	flags.StringVar(&outFile, "out", "", "File to write to, if left empty writes to stdout")
	flags.BoolVar(&overwrite, "f", false, "Overwrite the existing out file")
	flags.BoolVar(&runTests, "run", false, "Run the tests to produce the output file for the JSON samples. "+
		"The JSON samples need to be written to a file called the same as the source Go file minus the _test with a .apisamples extension")
	flags.BoolVar(&noheader, "no-header", false, "Do not set a document header and ToC")
	flags.BoolVar(&skipJSON, "skip-json", false, "Skip JSON sample sections in output (useful when .apisamples file doesn't exist)")
	flags.BoolVar(&antora, "antora", false, "Use Antora-compatible include paths (example$ prefix instead of absolute paths)")
	flags.BoolVar(&noPageBreaks, "no-page-breaks", false, "Do not insert page break markers (<<<) before sections")
	flags.BoolVar(&noHeadings, "no-headings", false, "Do not generate section headings from tag names")
	flags.BoolVar(&noOuterTags, "no-outer-tags", false, "Do not wrap generated content in outer tag markers")
	flags.StringVar(&relativeTo, "relative-to", "", "Make include paths relative to the specified directory")
	flags.StringVar(&includePrefix, "include-prefix", "", "Prepend a prefix to all include paths (e.g., 'example$')")
	flags.BoolVar(&dryRun, "dry-run", false, "Preview output to stdout without writing files")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if sourceFile == "" {
		fmt.Fprint(os.Stderr, "Sourcefile must be set\n")
		flags.PrintDefaults()
		os.Exit(100)
	}

	// Validate mutually exclusive path-handling flags
	pathFlagsSet := 0
	if antora {
		pathFlagsSet++
	}
	if relativeTo != "" {
		pathFlagsSet++
	}
	if includePrefix != "" {
		pathFlagsSet++
	}
	if pathFlagsSet > 1 {
		fmt.Fprint(os.Stderr, "Error: --antora, --relative-to, and --include-prefix are mutually exclusive. Use only one.\n")
		os.Exit(100)
	}

	filePath, goFilename := path.Split(sourceFile)
	data, err := ioutil.ReadFile(path.Join(filePath, goFilename))
	if err != nil {
		exitError("Could not read source file", err)
	}
	var f []string
	f = append(f, strings.Split(string(data), "\n")...)

	var funcs []int
	var docbuf []docOutput

	// first find all tests in the file
	for i, line := range f {
		if strings.Contains(line, "Test_") {
			funcs = append(funcs, i)
		}
	}

	// now we know where the funcs are we can test the read file for our docs
	for _, i := range funcs {
		doc := newDocOutput(filePath, goFilename)

		split := strings.Split(f[i], "Test_")
		split = strings.Split(split[1], "(")
		doc.funcName = split[0]

		search := i

		if search+1 == len(f) {
			break
		}

		// check if it's the first after the func declare, if it means this func has no docs
		if !strings.Contains(f[search+1], "startdocs") && !strings.Contains(f[search+1], "startapidocs") {
			continue
		}

		for {
			search++
			if search == len(f) {
				break
			}

			buf := f[search]
			buf = strings.Trim(buf, "\t")

			// title found
			if strings.Contains(buf, "startdocs") || strings.Contains(buf, "startapidocs") {
				if strings.Contains(buf, "startapidocs") {
					doc.apidocs = true
				}
				buf = strings.TrimPrefix(buf, "// startdocs")
				buf = strings.TrimPrefix(buf, "// startapidocs")
				doc.title = strings.Trim(buf, " ")
				continue
			}

			// end
			if strings.Contains(buf, "// enddocs") {
				break
			}

			// contents
			if strings.Contains(buf, "//") {
				if strings.Contains(buf, "// tag::") {
					continue
				}
				if strings.Contains(buf, "// end::") {
					continue
				}
				buf = strings.TrimPrefix(buf, "//")
				doc.body = append(doc.body, strings.Trim(buf, " "))
				continue
			}
		}

		if doc.title != "" {
			docbuf = append(docbuf, doc)
		}
	}

	_, title := path.Split(sourceFile)
	title = strings.Replace(title, "test.go", "", 1)
	title = strings.ReplaceAll(title, "_", " ")

	// Handle dry-run mode
	if dryRun {
		outFile = ""
	}

	if outFile != `` {
		if stat, _ := os.Stat(outFile); stat != nil {
			if !overwrite {
				exitError("File already exists", nil)
			}
			if err = os.Remove(outFile); err != nil {
				exitError("Could not delete file", err)
			}
		}
		o, err := os.Create(outFile)
		if err != nil {
			exitError("Could not create file for writing: "+outFile, err)
		}

		if !noheader {
			o.WriteString("= " + title + "\n")
			o.WriteString(":toc: left\n\n")
			o.WriteString("// THIS FILE IS GENERATED. DO NOT EDIT.\n")
		}
		for _, out := range docbuf {
			_, err = o.WriteString(out.getAsciiDoc())
			if err != nil {
				exitError("Could not write string to file: "+outFile, err)
			}
		}
	} else {
		if !noheader {
			fmt.Printf("= " + title + "\n")
			fmt.Printf(":toc: left\n\n")
			fmt.Printf("// THIS FILE IS GENERATED. DO NOT EDIT.\n")
		}
		for _, out := range docbuf {
			fmt.Printf("%s\n", out.getAsciiDoc())
		}
	}

	if runTests {
		var testsToRun []string
		for _, out := range docbuf {
			testsToRun = append(testsToRun, out.functionName())
		}

		tests := strings.Join(testsToRun, "|")
		packagedir, _ := path.Split(sourceFile)
		// command := "/usr/local/go/bin/go"
		args := []string{
			"test", "-timeout", "30s",
			packagedir,
			"-run",
			fmt.Sprintf("^(%s)$", tests),
			"-count", "1",
		}
		// fmt.Printf("Executing: \n%s %s\n", command, args)
		cmd := exec.Command("go", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("\nFailed to run for sourcefile: %s\n", sourceFile)
			fmt.Printf("Command: %s ", "go")
			for _, arg := range args {
				fmt.Printf("%s ", arg)
			}
			fmt.Printf("\n")
			exitError("Could not run tests", err)
		}
	}

	os.Exit(0)
}

func exitError(reason string, err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, reason+": "+err.Error()+"\n")
	} else {
		fmt.Fprint(os.Stderr, reason+"\n")
	}
	os.Exit(1)
}
