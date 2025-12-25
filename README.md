# code2asciidoc

## About

code2asciidoc solves the problem where code samples for API's are not properly
tested in most documentation.
It leverages the Go testing framework by writing the sample in there, using
the AsciiDoc tagging system and this tool joins it all together.

When using gRPC best use it in conjunction with [proto2asciidoc](https://github.com/productsupcom/proto2asciidoc)

## Installation

### Using go install (Recommended)

```bash
go install github.com/productsupcom/code2asciidoc@latest
```

This installs the `code2asciidoc` binary to your `$GOPATH/bin` (or `$GOBIN` if set).

### Using go tool

You can also run it directly without installing using `go tool`:

```bash
go tool github.com/productsupcom/code2asciidoc --source examples/examples_test.go --dry-run
```

### Building from Source

```bash
git clone https://github.com/productsupcom/code2asciidoc.git
cd code2asciidoc
go build -o code2asciidoc .
```

## Usage

### Basic Flags

**--source string**
Source file to parse into AsciiDoc. Can be absolute or relative path.

**--out string**
File to write to. If left empty, writes to stdout.

**--f**
Overwrite the existing output file if it exists.

**--run**
Run the tests to produce the output file for the JSON samples.
The JSON samples need to be written to a file called the same as
the source Go file minus the \_test with a .apisamples extension.

**--dry-run**
Preview output to stdout without writing files. Useful for testing before committing.

### Output Control Flags

**--no-header**
Do not set a document header (title, `:toc:`) and the "THIS FILE IS GENERATED" comment.
Use this when generating partials to be included in other documents.

**--skip-json**
Skip JSON sample sections in output. Useful when the `.apisamples` file doesn't exist
or when you only want to show Go code examples without JSON output.

**--no-page-breaks**
Do not insert page break markers (`<<<`) before sections.

**--no-headings**
Do not generate section headings (`== Title`) from tag names.

**--no-outer-tags**
Do not wrap generated content in outer `// tag::` and `// end::` markers.

### Path Handling Flags (Mutually Exclusive)

These flags control how `include::` paths are generated. Only one can be used at a time.

**--antora**
Use Antora-compatible include paths with `example$` prefix instead of absolute paths.
Example: `include::example$file.go[...]` instead of `include::/absolute/path/file.go[...]`

**--include-prefix string**
Prepend a custom prefix to all include paths.
Example: `--include-prefix "partials/"` generates `include::partials/file.go[...]`

**--relative-to string**
Make include paths relative to the specified directory.
Example: `--relative-to /home/user/project` converts absolute paths to relative ones.

## Output

You can view the output of the example below in Asciidoc here:
[Examples Documentation in AsciiDoc](https://github.com/productsupcom/code2asciidoc/blob/master/docs/generated/examples.adoc)

Or as a Markdown version (generated from Asciidoc through Pandoc output) here:
[Examples Documentation in Markdown](https://github.com/productsupcom/code2asciidoc/blob/master/docs/generated/examples.md)

## Example

The following file shows a complete example, it’s provided inside the `examples/`
directory including a Protobuf definition plus the compiled output.

The output can be seen at `docs/generated/examples.adoc`.

**Test Example.**

``` go
package examples

import (
    "testing"

    "github.com/productsupcom/code2asciidoc/sample"
)

func Test_ExampleSample1(t *testing.T) {
    // startapidocs Example Hello World
    // A simple demonstration to show the syntax of code2asciidoc.
    // You can use any form of asciidoc inside the comments, as it will be
    // properly parsed by AsciiDoc later.
    //
    // What is important is that the function of the test method is the same
    // as the Tags set below.
    // Also the name for the CreateJsonSample must match (the name of the current)
    // package!
    //
    // Be sure to place it inside a filename you can later ingest into your
    // API docs. Recommended is to keep it per gRPC Service.
    //
    // [NOTE.small]
    // ====
    // We can even use blocks in here if we want.
    // ====
    // startpostdocs Response
    // And in some cases you also want to have docs behind the samples to make
    // data more clear.
    // endpostdocs
    // enddocs
    // tag::ExampleSample1[]
    ex := Example{
        SomeString: "Foo",
        SomeInt:    1,
    }
    // end::ExampleSample1[]

    f, err := sample.Setup("examples.apisamples")
    if err != nil {
        t.Errorf("%v", err)
    }
    err = sample.CreateJsonSample(&ex, "ExampleSample1", f)
    if err != nil {
        t.Errorf("%v", err)
    }
}
```

You can produce the same output by doing:

``` shell
code2asciidoc --source examples/examples_test.go --out docs/generated/examples.adoc --run --f
```

Or using `go tool` without installing:

``` shell
go tool github.com/productsupcom/code2asciidoc --source examples/examples_test.go --out docs/generated/examples.adoc --run --f
```

The `--run` causes the tool to call the Go test-suite which will produce the
output files.

## Antora Integration

When integrating with Antora documentation, you need to handle file locations and include paths carefully.

### Directory Structure

For an Antora module, your structure typically looks like:

```
src/my-service/
├── examples/
│   └── client_example.go          # Source examples (Go module)
├── docs/
│   ├── antora.yml
│   └── modules/
│       └── ROOT/
│           ├── examples/
│           │   └── client_example.go    # Copy for Antora resolution
│           ├── partials/
│           │   └── generated/
│           │       └── client_example.adoc  # Generated partial
│           └── pages/
│               └── api-guide.adoc       # Main documentation
```

### Workflow

1. **Write your example** in `src/my-service/examples/client_example.go`:

```go
func Test_ClientSetup(t *testing.T) {
    // startapidocs Client Setup
    // This example shows how to create a new client.
    // enddocs
    // tag::ClientSetup[]
    client := NewClient()
    // end::ClientSetup[]
}
```

2. **Generate the AsciiDoc partial** using the `--antora` flag:

```shell
code2asciidoc \
  --source src/my-service/examples/client_example.go \
  --out src/my-service/docs/modules/ROOT/partials/generated/client_example.adoc \
  --antora \
  --no-header \
  --skip-json \
  --f
```

This generates includes like: `include::example$client_example.go[tag=ClientSetup,indent=0]`

3. **Copy the source file** to Antora's examples directory:

```shell
cp src/my-service/examples/client_example.go \
   src/my-service/docs/modules/ROOT/examples/
```

4. **Include the partial** in your documentation:

```asciidoc
= API Guide

\include::partial$generated/client_example.adoc[tag=ClientSetup]
```

### Recommended Flags for Antora

For generating **partials** to be included in other documents:
```shell
--antora --no-header --skip-json --no-outer-tags
```

For generating **standalone pages**:
```shell
--antora --skip-json
```

### Automation with go:generate

You can automate generation using `go:generate` directives in your test files.

**Option 1: Using installed binary**

```go
//go:generate sh -c "code2asciidoc --source client_example.go --out ../../docs/modules/ROOT/partials/generated/client_example.adoc --antora --no-header --skip-json --f && cp client_example.go ../../docs/modules/ROOT/examples/"

func Test_ClientSetup(t *testing.T) {
    // ...
}
```

**Option 2: Using go tool (no installation required)**

```go
//go:generate sh -c "go tool github.com/productsupcom/code2asciidoc --source client_example.go --out ../../docs/modules/ROOT/partials/generated/client_example.adoc --antora --no-header --skip-json --f && cp client_example.go ../../docs/modules/ROOT/examples/"

func Test_ClientSetup(t *testing.T) {
    // ...
}
```

Then run:
```shell
cd src/my-service/examples
go generate
```

**Note:** Using `go tool` ensures everyone on the team uses the same version without requiring manual installation.

### Using as a Go Tool

You can also use `code2asciidoc` as a Go tool in your `tools.go` file for version pinning:

1. Create a `tools.go` file in your project:

```go
//go:build tools
// +build tools

package tools

import (
    _ "github.com/productsupcom/code2asciidoc"
)
```

2. Add it to your `go.mod`:

```shell
go mod tidy
```

3. Run it using `go tool`:

```shell
go tool github.com/productsupcom/code2asciidoc --source examples/example.go --antora --dry-run
```

This approach ensures:
- Version is tracked in `go.mod`
- Everyone on the team uses the same version
- No manual installation required
- Works in CI/CD environments

### Troubleshooting

**Error: "target of include not found: example$file.go"**
- Make sure you copied the source file to `docs/modules/ROOT/examples/`
- Verify the filename matches exactly (case-sensitive)

**Error: "target of include not found: example$file.apisamples"**
- Use `--skip-json` flag if you don't need JSON samples
- Or run with `--run` to generate the `.apisamples` file

**Absolute paths in generated includes**
- Make sure you're using `--antora` flag
- Check that you're not also using `--relative-to` or `--include-prefix` (mutually exclusive)
