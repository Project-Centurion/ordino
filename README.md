# Ordino

This package sorts Go imports in groups and in a default or user-defined order. 

## Shoulders

This customized package uses a lot of the logic and some tools from [goimports-reviser](https://github.com/incu6us/goimports-reviser)

## Requirements

* Go 1.17;
* [GolangCI-Lint CMD][4];
* [Staticcheck CMD][15];

## Usage

Install it by running:

```shell
go install github.com/Project-Centurion/ordino@latest
```

```shell
ordino -project-name [YourProjectName] -output [TheOutPutYouWant] -order [thePackagesOrderYouWant] [file/path/to/your/gofile.go]
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
