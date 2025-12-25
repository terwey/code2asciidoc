# code2asciidoc Improvements Demo

This document demonstrates the fixes for the critical issues identified in `code2asciidoc-problems.md`.

## Issue #1: Mandatory JSON Sample Sections (FIXED)

**Problem:** JSON sections were always generated, breaking builds when `.apisamples` files didn't exist.

**Before:**
```bash
code2asciidoc --source examples/examples_test.go --out output.adoc
# Always includes JSON section, even without .apisamples file
```

**After:**
```bash
code2asciidoc --source examples/examples_test.go --out output.adoc --skip-json
# JSON section completely omitted
```

**Result:** No more Antora build errors for missing `.apisamples` files.

---

## Issue #2: Absolute Paths in Generated Includes (FIXED)

**Problem:** Generated includes used absolute paths, breaking portability and Antora resolution.

**Before:**
```asciidoc
include::/home/terwey/architecture/src/identity-integration-service/examples/client_setup_example.go[tag=ClientSetup,indent=0]
```

**After (with --antora flag):**
```asciidoc
include::example$client_setup_example.go[tag=ClientSetup,indent=0]
```

**Result:** Portable, Antora-compatible includes that work across machines and in CI/CD.

---

## Issue #3: Document Header in Partials (FIXED)

**Problem:** `--no-header` didn't suppress the "THIS FILE IS GENERATED" comment, breaking partial includes.

**Before (with --no-header):**
```asciidoc
// THIS FILE IS GENERATED. DO NOT EDIT.
// tag::ExampleSample1[]
<<<
== Example Hello World
...
```

**After (with --no-header):**
```asciidoc
// tag::ExampleSample1[]
<<<
== Example Hello World
...
```

**Result:** True partials that can be cleanly included in other documents.

---

## Recommended Antora Workflow

### 1. Generate Partial for Inclusion

```bash
code2asciidoc \
  --source src/my-service/examples/client_example.go \
  --out src/my-service/docs/modules/ROOT/partials/generated/client_example.adoc \
  --antora \
  --no-header \
  --skip-json \
  --no-outer-tags \
  --f
```

**Flags explained:**
- `--antora`: Use `example$` prefix for Antora compatibility
- `--no-header`: Suppress document header (it's a partial)
- `--skip-json`: Skip JSON sections (no `.apisamples` file)
- `--no-outer-tags`: Remove wrapper tags for cleaner includes
- `--f`: Overwrite existing file

### 2. Copy Source to Antora Examples

```bash
cp src/my-service/examples/client_example.go \
   src/my-service/docs/modules/ROOT/examples/
```

### 3. Include in Documentation

```asciidoc
= API Guide

\include::partial$generated/client_example.adoc[]
```

---

## Additional Improvements

### Suppress Page Breaks and Headings

For minimal output when you want full control:

```bash
code2asciidoc \
  --source examples/examples_test.go \
  --no-header \
  --no-page-breaks \
  --no-headings \
  --no-outer-tags \
  --skip-json \
  --dry-run
```

### Preview Before Writing

```bash
code2asciidoc --source examples/examples_test.go --dry-run | less
```

### Custom Include Prefixes

```bash
code2asciidoc --source examples/examples_test.go --include-prefix "partials/generated/"
# Generates: include::partials/generated/examples_test.go[...]
```

---

## Validation

The tool now validates mutually exclusive flags:

```bash
$ code2asciidoc --source file.go --antora --include-prefix "foo"
Error: --antora, --relative-to, and --include-prefix are mutually exclusive. Use only one.
```

---

## Summary of Fixes

| Issue | Status | Solution |
|-------|--------|----------|
| #1: Mandatory JSON sections | ✅ FIXED | `--skip-json` flag |
| #2: Absolute paths | ✅ FIXED | `--antora` flag |
| #3: Headers in partials | ✅ FIXED | `--no-header` now fully suppresses headers |
| #5: Page breaks | ✅ FIXED | `--no-page-breaks` flag |
| #6: Section headings | ✅ FIXED | `--no-headings` flag |
| #7: Double tag wrapping | ✅ FIXED | `--no-outer-tags` flag |
| #17: No dry-run mode | ✅ FIXED | `--dry-run` flag |

All fixes are **backward compatible** - default behavior unchanged.
