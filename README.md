# FuzzyPatch

FuzzyPatch is a Go library that enables fuzzy matching and patching of text files.
It's particularly useful for applying patches when the exact line positions may have changed.
It's meant for use with AI coding agents which often make small mistakes when generating diffs.

## Installation

```bash
go get github.com/icholy/fuzzypatch
```

## Usage

FuzzyPatch accepts diff in the following format:

```
<<<<<<< SEARCH line:<n>
[search text...]
=======
[replace text...]
>>>>>>> REPLACE
```

Where:
- `<n>` is the line number hint where the search should start
- `[search text...]` is the text to find (can span multiple lines)
- `[replace text...]` is the text to replace it with (can span multiple lines)

### Example

```go
package main

import (
    "fmt"
    "github.com/icholy/fuzzypatch"
)

var difftext = `
<<<<<<< SEARCH line:2
    console.log("Hello, world!");
=======
    console.log("Hello, Universe!");
>>>>>>> REPLACE
`

var source = `
function sayHello() {
    console.log("Hello, world!");
    // Some comment
    doSomethingElse();
}
`

func main() {
    // Parse a diff block
    diffs, _ := fuzzypatch.Parse(difftext)

    // Map diffs to edits
    var edits []fuzzypatch.Edit
    for _, diff := range diffs {
        // Search with 0.9 similarity threshold (90% similar)
        if edit, ok := fuzzypatch.Search(source, diff, 0.9); ok {
            edits = append(edits, edit)
        }
    }

    // Apply all edits in one pass
    result, _ := fuzzypatch.Apply(source, edits)
    fmt.Println(result)
}
```