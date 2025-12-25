#!/bin/bash
set -e

echo "=== Testing code2asciidoc new features ==="
echo ""

echo "1. Testing --skip-json flag..."
./code2asciidoc --source examples/examples_test.go --skip-json --dry-run | grep -q "\.JSON" && echo "FAIL: JSON section found" || echo "PASS: JSON section skipped"
echo ""

echo "2. Testing --antora flag..."
./code2asciidoc --source examples/examples_test.go --antora --dry-run | grep -q "include::example\\\$" && echo "PASS: Antora paths generated" || echo "FAIL: Antora paths not found"
echo ""

echo "3. Testing --no-header flag..."
./code2asciidoc --source examples/examples_test.go --no-header --dry-run | grep -q "THIS FILE IS GENERATED" && echo "FAIL: Header comment found" || echo "PASS: All headers suppressed"
echo ""

echo "4. Testing --no-page-breaks flag..."
./code2asciidoc --source examples/examples_test.go --no-page-breaks --dry-run | grep -q "<<<" && echo "FAIL: Page breaks found" || echo "PASS: Page breaks suppressed"
echo ""

echo "5. Testing --no-headings flag..."
./code2asciidoc --source examples/examples_test.go --no-headings --dry-run | grep -q "^== " && echo "FAIL: Headings found" || echo "PASS: Headings suppressed"
echo ""

echo "6. Testing --no-outer-tags flag..."
./code2asciidoc --source examples/examples_test.go --no-outer-tags --dry-run | grep -q "^// tag::ExampleSample1" && echo "FAIL: Outer tags found" || echo "PASS: Outer tags suppressed"
echo ""

echo "7. Testing --include-prefix flag..."
./code2asciidoc --source examples/examples_test.go --include-prefix "custom/" --dry-run | grep -q "include::custom/" && echo "PASS: Custom prefix applied" || echo "FAIL: Custom prefix not found"
echo ""

echo "8. Testing mutually exclusive flags..."
./code2asciidoc --source examples/examples_test.go --antora --include-prefix "foo" --dry-run 2>&1 | grep -q "mutually exclusive" && echo "PASS: Validation working" || echo "FAIL: Validation not working"
echo ""

echo "9. Testing --dry-run flag..."
./code2asciidoc --source examples/examples_test.go --out /tmp/should_not_exist.adoc --dry-run > /dev/null 2>&1
[ ! -f /tmp/should_not_exist.adoc ] && echo "PASS: Dry-run doesn't write files" || echo "FAIL: File was written"
echo ""

echo "10. Testing combined flags (Antora workflow)..."
OUTPUT=$(./code2asciidoc --source examples/examples_test.go --antora --no-header --skip-json --no-outer-tags --dry-run)
echo "$OUTPUT" | grep -q "include::example\\\$" && \
! echo "$OUTPUT" | grep -q "THIS FILE IS GENERATED" && \
! echo "$OUTPUT" | grep -q "\.JSON" && \
! echo "$OUTPUT" | grep -q "^// tag::" && \
echo "PASS: Antora workflow flags work together" || echo "FAIL: Combined flags issue"
echo ""

echo "=== All tests complete ==="
