# Ordino

This package sorts Go imports in groups and in a default or user-defined order. 

## Shoulders

This customized package uses a lot of the logic and some tools from [goimports-reviser](https://github.com/incu6us/goimports-reviser)

## Requirements

* Go 1.17;
* [GolangCI-Lint CMD](https://github.com/golangci/golangci-lint);
* [Staticcheck CMD](https://staticcheck.io);

## Usage

Install it by running:

```shell
go install github.com/Project-Centurion/ordino@latest
```

```shell
ordino -project-name [YourProjectName] -output [TheOutPutYouWant] -order [thePackagesOrderYouWant] [file/path/to/your/gofile.go]
```

### Warnings

:warning: On actual stable version of ordino, any comment on top of an import (doc) will be removed by `ordino`,
any comment at the end of an import (comment) will stay.

On next versions of `ordino`, this should be fixed. Do not hesitate to raise a PR to propose a fix.

Example of comment which will be remove:
```go
import (
	"bytes"
	"fmt"
	"go/ast" //this a comment that will stay
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	// This is a comment that will be removed
	"github.com/incu6us/goimports-reviser/v2/pkg/std"
)
```

### Required arguments

* single file sorting

> `file/path/to/your/gofile.go` : the path from the current directory to the file where you want your imports sorted.

* recursive run

> `./...` : sorts imports from all the `.go` files under the current directory

example :

```shell
ordino -project-name [YourProjectName] ./...
```

### Optional arguments

* `theOutPutYouWant` : either `file` or `stdout`, by default `file` (will rewrite the files sorted).
* `thePackagesOrderYouWant` : constructed like this `std,alias,project,general` by default, meaning the order you want between
the packages, separated by commas, no spaces. Aliased packages being separated are optional.
* `YourProjectName` : the imports you want sorted as `project` imports. Please provide the path to those imports such as `github.com/MyGreatProject`
or `github.com/MyGreatProject/mySuperGreatGoRepository`. If not set, the project name will be fetched from `go.mod`


### Real life examples: 

```shell
ordino -order std,project,general -project-name github.com/Project-Centurion/ordino ./...
```

```shell
ordino -order project,general,std -output stdout main.go
```
