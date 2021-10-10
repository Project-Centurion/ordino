# Ordino

This package sorts imports according to [inducula conventions](https://github.com/Project-Centurion/inducula/blob/master/CONVENTION.md)

## Shoulders

This customized package uses a lot of the logic and some tools from [goimports-reviser](https://github.com/incu6us/goimports-reviser)

## Requirements

* Go 1.17;

## Usage

Install it by running:

```shell
go install github.com/Project-Centurion/ordino@latest
```

Then run :

```shell
ordino -project-name <YourProjectName> -file-path <PathToYourFile> -output <TheOutPutYouWant>
```

* **PathToYourFile** : (*required*) path to the file you want to sort
* **YourProjectName** : (*optional*) if not set will fetch project name from `go.mod`
* **TheOutPutYouWant** : (*optional*) either `file` or `stdout`, by default `file` (will rewrite the file specified in `-file-path`).
