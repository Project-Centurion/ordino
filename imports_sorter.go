package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/incu6us/goimports-reviser/v2/pkg/std"
)

const (
	stdPkg     = "std"
	aliasedPkg = "alias"
	generalPkg = "general"
	projectPkg = "project"
)

func Execute(projectName, filePath string, order []string) ([]byte, bool, error) {
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	fset := token.NewFileSet()

	pf, err := parser.ParseFile(fset, "", originalContent, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	importsWithMetadata := parseImports(pf)

	stdImports, generalImports, aliasedImports, projectImports := groupImports(
		projectName,
		importsWithMetadata,
		order,
	)

	decls, ok := hasMultipleImportDecls(pf)
	if ok {
		pf.Decls = decls
	}

	fixImports(pf, stdImports, generalImports, aliasedImports, projectImports, importsWithMetadata, order)

	fixedImportsContent, err := generateFile(fset, pf)
	if err != nil {
		return nil, false, err
	}

	formattedContent, err := format.Source(fixedImportsContent)
	if err != nil {
		return nil, false, err
	}

	return formattedContent, !bytes.Equal(originalContent, formattedContent), nil
}

func generateFile(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, f); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func parseImports(f *ast.File) map[string]*commentsMetadata {
	importsWithMetadata := map[string]*commentsMetadata{}

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				for _, spec := range dd.Specs {
					var importSpecStr string
					importSpec := spec.(*ast.ImportSpec)

					if importSpec.Name != nil {
						importSpecStr = strings.Join([]string{importSpec.Name.String(), importSpec.Path.Value}, " ")
					} else {
						importSpecStr = importSpec.Path.Value
					}

					importsWithMetadata[importSpecStr] = &commentsMetadata{
						Doc:     importSpec.Doc,
						Comment: importSpec.Comment,
					}
				}
			}
		}
	}

	return importsWithMetadata
}

func isSingleGoImport(dd *ast.GenDecl) bool {
	if dd.Tok != token.IMPORT {
		return false
	}
	if len(dd.Specs) != 1 {
		return false
	}
	return true
}

func groupImports(
	projectName string,
	importsWithMetadata map[string]*commentsMetadata,
	order []string,
) ([]string, []string, []string, []string) {
	var (
		stdImports       []string
		projectImports   []string
		importsWithAlias []string
		generalImports   []string
	)

	orderContainsAlias := OrderContainsAlias(order)

	for imprt := range importsWithMetadata {
		pkgWithoutAlias := skipPackageAlias(imprt)

		if _, ok := std.StdPackages[pkgWithoutAlias]; ok {
			stdImports = append(stdImports, imprt)
			continue
		}

		if orderContainsAlias && isPackageWithAlias(imprt) {
			importsWithAlias = append(importsWithAlias, imprt)
			continue
		}

		if strings.Contains(pkgWithoutAlias, projectName) {
			projectImports = append(projectImports, imprt)
			continue
		}

		generalImports = append(generalImports, imprt)
	}

	sort.Strings(stdImports)
	sort.Strings(generalImports)
	sort.Strings(importsWithAlias)
	sort.Strings(projectImports)

	return stdImports, generalImports, importsWithAlias, projectImports
}

func skipPackageAlias(pkg string) string {
	values := strings.Split(pkg, " ")
	if len(values) > 1 {
		return strings.Trim(values[1], `"`)
	}

	return strings.Trim(pkg, `"`)
}

func isPackageWithAlias(pkg string) bool {
	values := strings.Split(pkg, " ")

	return len(values) > 1
}

// hasMultipleImportDecls will return combined import declarations to single declaration
//
// Ex.:
// import "fmt"
// import "io"
// -----
// to
// -----
// import (
// 	"fmt"
//	"io"
// )
func hasMultipleImportDecls(f *ast.File) ([]ast.Decl, bool) {
	importSpecs := make([]ast.Spec, 0, len(f.Imports))
	for _, importSpec := range f.Imports {
		importSpecs = append(importSpecs, importSpec)
	}

	var (
		hasMultipleImportDecls   bool
		isFirstImportDeclDefined bool
	)

	decls := make([]ast.Decl, 0, len(f.Decls))
	for _, decl := range f.Decls {
		dd, ok := decl.(*ast.GenDecl)
		if !ok {
			decls = append(decls, decl)
			continue
		}

		if dd.Tok != token.IMPORT || isSingleGoImport(dd) {
			decls = append(decls, dd)
			continue
		}

		if isFirstImportDeclDefined {
			hasMultipleImportDecls = true
			storedGenDecl := decls[len(decls)-1].(*ast.GenDecl)
			if storedGenDecl.Tok == token.IMPORT {
				storedGenDecl.Rparen = dd.End()
			}
			continue
		}

		dd.Specs = importSpecs
		decls = append(decls, dd)
		isFirstImportDeclDefined = true
	}

	return decls, hasMultipleImportDecls
}

func fixImports(
	f *ast.File,
	stdImports, generalImports, aliasedImports, projectImports []string,
	commentsMetadata map[string]*commentsMetadata,
	order []string,
) {
	var importsPositions []*importPosition
	for _, decl := range f.Decls {
		dd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if dd.Tok != token.IMPORT {
			continue
		}

		importsPositions = append(
			importsPositions, &importPosition{
				Start: dd.Pos(),
				End:   dd.End(),
			},
		)

		dd.Specs = rebuildImports(dd.Tok, commentsMetadata, stdImports, generalImports, aliasedImports, projectImports, order)
	}

	clearImportDocs(f, importsPositions)
	removeEmptyImportNode(f)
}

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
		case stdPkg:
			order[i] = stdImports
		case aliasedPkg:
			order[i] = aliasedImports
		case projectPkg:
			order[i] = projectImports
		case generalPkg:
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

type commentsMetadata struct {
	Doc     *ast.CommentGroup
	Comment *ast.CommentGroup
}

type importPosition struct {
	Start token.Pos
	End   token.Pos
}

func (p *importPosition) IsInRange(comment *ast.CommentGroup) bool {
	if p.Start <= comment.Pos() && comment.Pos() <= p.End {
		return true
	}

	return false
}

func importWithComment(imprt string, commentsMetadata map[string]*commentsMetadata) string {
	var comment string
	commentGroup, ok := commentsMetadata[imprt]
	if ok {
		if commentGroup != nil && commentGroup.Comment != nil && len(commentGroup.Comment.List) > 0 {
			comment = fmt.Sprintf("// %s", strings.ReplaceAll(commentGroup.Comment.Text(), "\n", ""))
		}
	}

	return fmt.Sprintf("%s %s", imprt, comment)
}

func OrderContainsAlias(order []string) bool {
	for _, o := range order {
		if o == aliasedPkg {
			return true
		}
	}
	return false
}

func clearImportDocs(f *ast.File, importsPositions []*importPosition) {
	importsComments := make([]*ast.CommentGroup, 0, len(f.Comments))

	for _, comment := range f.Comments {
		for _, importPosition := range importsPositions {
			if importPosition.IsInRange(comment) {
				continue
			}
			importsComments = append(importsComments, comment)
		}
	}

	if len(f.Imports) > 0 {
		f.Comments = importsComments
	}
}

func removeEmptyImportNode(f *ast.File) {
	var (
		decls      []ast.Decl
		hasImports bool
	)

	for _, decl := range f.Decls {
		dd, ok := decl.(*ast.GenDecl)
		if !ok {
			decls = append(decls, decl)

			continue
		}

		if dd.Tok == token.IMPORT && len(dd.Specs) > 0 {
			hasImports = true

			break
		}

		if dd.Tok != token.IMPORT {
			decls = append(decls, decl)
		}
	}

	if !hasImports {
		f.Decls = decls
	}
}
