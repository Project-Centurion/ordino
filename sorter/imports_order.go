package sorter

import (
	"go/ast"
	"go/token"
)

func rebuildImports(
	tok token.Token,
	commentsMetadata map[string]*commentsMetadata,
	stdImports []string,
	generalImports []string,
	aliasedImports []string,
	projectImports []string,
	flagOrders []string,
) []ast.Spec {
	var specs []ast.Spec

	order := map[int][]string{}
	for i, flagOrder := range flagOrders {
		switch flagOrder {
		case StdPkg:
			order[i] = stdImports
		case AliasedPkg:
			order[i] = aliasedImports
		case ProjectPkg:
			order[i] = projectImports
		case GeneralPkg:
			order[i] = generalImports
		}
	}

	linesCounter := len(order[0])
	for _, firstImportGroup := range order[0] {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(firstImportGroup, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && (len(order[1]) > 0 || len(order[2]) > 0 || len(order[3]) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	linesCounter = len(order[1])
	for _, secondImportGroup := range order[1] {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(secondImportGroup, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && (len(order[2]) > 0 || len(order[3]) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	linesCounter = len(order[2])
	for _, thirdImportGroup := range order[2] {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(thirdImportGroup, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && len(order[3]) > 0 {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	for _, fourthImportGroup := range order[3] {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(fourthImportGroup, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)
	}

	return specs
}
