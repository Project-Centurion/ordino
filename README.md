# Ordino

![CI Status][1]
![License MIT][2]
![Release][3]

This package sorts Go imports in groups and in a default or user-defined order.

## Shoulders

This customized package uses a lot of the logic and some tools from [goimports-reviser](https://github.com/incu6us/goimports-reviser)

## Requirements

* Go 1.17;
* [GolangCI-Lint CMD][4];
* [Staticcheck CMD][5];

## Usage

Install it by running:

```shell
go install github.com/Project-Centurion/ordino@latest
```

Then run:

```shell
ordino -project-name [YourProjectName] -output [TheOutPutYouWant] -order [thePackagesOrderYouWant] [file/path/to/your/gofile.go]
```

More information on named and unnamed args [here](#required-arguments)

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

:warning: :bug: when two import declarations are set, they will not be squashed together. Working on a fix

### Required arguments

* single file sorting

> `file/path/to/your/gofile.go` : the path from the current directory to the file where you want your imports sorted.

example :

```shell
ordino -project-name [YourProjectName] pkg/some_go_gile.go
```

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

### Real life examples

```shell
ordino -order std,project,general -project-name github.com/Project-Centurion/ordino ./...
```

```shell
ordino -order project,general,std -output stdout main.go
```

### Future projects

* Add a 5th option to sort specific patterns
* Add a way to configure sorting through a `.yml` config file

[1]: https://github.com/Project-Centurion/ordino/workflows/Lint%20&%20Build%20-%20GoLang/badge.svg
[2]: https://img.shields.io/github/license/Project-Centurion/ordino
[3]: https://img.shields.io/github/v/release/Project-Centurion/ordino
[4]: https://github.com/golangci/golangci-lint
[5]: https://staticcheck.io
