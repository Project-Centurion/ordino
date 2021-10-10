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

func Execute(projectName, filePath string) ([]byte, bool, error) {
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
	)

	decls, ok := hasMultipleImportDecls(pf)
	if ok {
		pf.Decls = decls
	}

	fixImports(pf, stdImports, generalImports, aliasedImports, projectImports, importsWithMetadata)

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
			if isSingleCgoImport(dd) {
				continue
			}
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

func isSingleCgoImport(dd *ast.GenDecl) bool {
	if dd.Tok != token.IMPORT {
		return false
	}
	if len(dd.Specs) != 1 {
		return false
	}
	return dd.Specs[0].(*ast.ImportSpec).Path.Value == `"C"`
}

func groupImports(
	projectName string,
	importsWithMetadata map[string]*commentsMetadata,
) ([]string, []string, []string, []string) {
	var (
		stdImports       []string
		projectImports   []string
		importsWithAlias []string
		generalImports   []string
	)

	for imprt := range importsWithMetadata {
		pkgWithoutAlias := skipPackageAlias(imprt)

		if _, ok := std.StdPackages[pkgWithoutAlias]; ok {
			stdImports = append(stdImports, imprt)
			continue
		}

		if isPackageWithAlias(imprt) {
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
	if len(values) > 1 {
		return true
	}
	return false
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

		if dd.Tok != token.IMPORT || isSingleCgoImport(dd) {
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
) {
	var importsPositions []*importPosition
	for _, decl := range f.Decls {
		dd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if dd.Tok != token.IMPORT || isSingleCgoImport(dd) {
			continue
		}

		importsPositions = append(
			importsPositions, &importPosition{
				Start: dd.Pos(),
				End:   dd.End(),
			},
		)

		dd.Specs = rebuildImports(dd.Tok, commentsMetadata, stdImports, generalImports, aliasedImports, projectImports)
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
) []ast.Spec {
	var specs []ast.Spec

	linesCounter := len(stdImports)
	for _, stdImport := range stdImports {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(stdImport, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && (len(generalImports) > 0 || len(aliasedImports) > 0 || len(projectImports) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	linesCounter = len(aliasedImports)
	for _, projectLocalPkg := range aliasedImports {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(projectLocalPkg, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && (len(generalImports) > 0 || len(aliasedImports) > 0 || len(projectImports) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	for _, projectImport := range projectImports {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(projectImport, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		if linesCounter == 0 && (len(generalImports) > 0 || len(aliasedImports) > 0 || len(projectImports) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
	}

	linesCounter = len(generalImports)
	for _, generalImport := range generalImports {
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: importWithComment(generalImport, commentsMetadata), Kind: tok},
		}
		specs = append(specs, spec)

		linesCounter--

		if linesCounter == 0 && (len(generalImports) > 0 || len(aliasedImports) > 0 || len(projectImports) > 0) {
			spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

			specs = append(specs, spec)
		}
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
