package model

import (
	"encoding/xml"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

const CoberturaDTDDecl = `<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">`

type Coverage struct {
	XMLName         xml.Name   `xml:"coverage"`
	LineRate        float32    `xml:"line-rate,attr"`
	BranchRate      float32    `xml:"branch-rate,attr"`
	Version         string     `xml:"version,attr"`
	Timestamp       int64      `xml:"timestamp,attr"`
	LinesCovered    int64      `xml:"lines-covered,attr"`
	LinesValid      int64      `xml:"lines-valid,attr"`
	BranchesCovered int64      `xml:"branches-covered,attr"`
	BranchesValid   int64      `xml:"branches-valid,attr"`
	Complexity      float32    `xml:"complexity,attr"`
	Sources         []*Source  `xml:"sources>source"`
	Packages        []*Package `xml:"packages>package"`
}

type Source struct {
	Path string `xml:",chardata"`
}

type Package struct {
	Name       string   `xml:"name,attr"`
	LineRate   float32  `xml:"line-rate,attr"`
	BranchRate float32  `xml:"branch-rate,attr"`
	Complexity float32  `xml:"complexity,attr"`
	Classes    []*Class `xml:"classes>class"`
}

type Class struct {
	Name       string    `xml:"name,attr"`
	Filename   string    `xml:"filename,attr"`
	LineRate   float32   `xml:"line-rate,attr"`
	BranchRate float32   `xml:"branch-rate,attr"`
	Complexity float32   `xml:"complexity,attr"`
	Methods    []*Method `xml:"methods>method"`
	Lines      Lines     `xml:"lines>line"`
}

type Method struct {
	Name       string  `xml:"name,attr"`
	Signature  string  `xml:"signature,attr"`
	LineRate   float32 `xml:"line-rate,attr"`
	BranchRate float32 `xml:"branch-rate,attr"`
	Complexity float32 `xml:"complexity,attr"`
	Lines      Lines   `xml:"lines>line"`
}

type Line struct {
	Number int   `xml:"number,attr"`
	Hits   int64 `xml:"hits,attr"`
}

// Lines is a slice of Line pointers, with some convenience methods
type Lines []*Line

// HitRate returns a float32 from 0.0 to 1.0 representing what fraction of lines
// have hits
func (lines Lines) HitRate() (hitRate float32) {
	return float32(lines.NumLinesWithHits()) / float32(len(lines))
}

// NumLines returns the number of lines
func (lines Lines) NumLines() int64 {
	return int64(len(lines))
}

// NumLinesWithHits returns the number of lines with a hit count > 0
func (lines Lines) NumLinesWithHits() (numLinesWithHits int64) {
	for _, line := range lines {
		if line.Hits > 0 {
			numLinesWithHits++
		}
	}
	return numLinesWithHits
}

// AddOrUpdateLine adds a line if it is a different line than the last line recorded.
// If it's the same line as the last line recorded then we update the hits down
// if the new hits is less; otherwise just leave it as-is
func (lines *Lines) AddOrUpdateLine(lineNumber int, hits int64) {
	if len(*lines) > 0 {
		lastLine := (*lines)[len(*lines)-1]
		if lineNumber == lastLine.Number {
			if hits < lastLine.Hits {
				lastLine.Hits = hits
			}
			return
		}
	}
	*lines = append(*lines, &Line{Number: lineNumber, Hits: hits})
}

// HitRate returns a float32 from 0.0 to 1.0 representing what fraction of lines
// have hits
func (method Method) HitRate() float32 {
	return method.Lines.HitRate()
}

// NumLines returns the number of lines
func (method Method) NumLines() int64 {
	return method.Lines.NumLines()
}

// NumLinesWithHits returns the number of lines with a hit count > 0
func (method Method) NumLinesWithHits() int64 {
	return method.Lines.NumLinesWithHits()
}

// HitRate returns a float32 from 0.0 to 1.0 representing what fraction of lines
// have hits
func (class Class) HitRate() float32 {
	return float32(class.NumLinesWithHits()) / float32(class.NumLines())
}

// NumLines returns the number of lines
func (class Class) NumLines() (numLines int64) {
	for _, method := range class.Methods {
		numLines += method.NumLines()
	}
	return numLines
}

// NumLinesWithHits returns the number of lines with a hit count > 0
func (class Class) NumLinesWithHits() (numLinesWithHits int64) {
	for _, method := range class.Methods {
		numLinesWithHits += method.NumLinesWithHits()
	}
	return numLinesWithHits
}

// HitRate returns a float32 from 0.0 to 1.0 representing what fraction of lines
// have hits
func (pkg Package) HitRate() float32 {
	return float32(pkg.NumLinesWithHits()) / float32(pkg.NumLines())
}

// NumLines returns the number of lines
func (pkg Package) NumLines() (numLines int64) {
	for _, class := range pkg.Classes {
		numLines += class.NumLines()
	}
	return numLines
}

// NumLinesWithHits returns the number of lines with a hit count > 0
func (pkg Package) NumLinesWithHits() (numLinesWithHits int64) {
	for _, class := range pkg.Classes {
		numLinesWithHits += class.NumLinesWithHits()
	}
	return numLinesWithHits
}

// HitRate returns a float32 from 0.0 to 1.0 representing what fraction of lines
// have hits
func (cov Coverage) HitRate() float32 {
	return float32(cov.NumLinesWithHits()) / float32(cov.NumLines())
}

// NumLines returns the number of lines
func (cov Coverage) NumLines() (numLines int64) {
	for _, pkg := range cov.Packages {
		numLines += pkg.NumLines()
	}
	return numLines
}

// NumLinesWithHits returns the number of lines with a hit count > 0
func (cov Coverage) NumLinesWithHits() (numLinesWithHits int64) {
	for _, pkg := range cov.Packages {
		numLinesWithHits += pkg.NumLinesWithHits()
	}
	return numLinesWithHits
}

func AppendIfUnique(sources []*Source, dir string) []*Source {
	for _, source := range sources {
		if source.Path == dir {
			return sources
		}
	}
	return append(sources, &Source{dir})
}

func GetPackages(profiles []*Profile) ([]*packages.Package, error) {
	if len(profiles) == 0 {
		return []*packages.Package{}, nil
	}

	var pkgNames []string
	for _, profile := range profiles {
		pkgNames = append(pkgNames, getPackageName(profile.FileName))
	}
	return packages.Load(&packages.Config{Mode: packages.NeedFiles | packages.NeedModule}, pkgNames...)
}

func getPackageName(filename string) string {
	pkgName, _ := filepath.Split(filename)
	// TODO(boumenot): Windows vs. Linux
	return strings.TrimRight(strings.TrimRight(pkgName, "\\"), "/")
}

func findAbsFilePath(pkg *packages.Package, profileName string) string {
	filename := filepath.Base(profileName)
	for _, fullpath := range pkg.GoFiles {
		if filepath.Base(fullpath) == filename {
			return fullpath
		}
	}
	return ""
}

func (cov *Coverage) ParseProfiles(profiles []*Profile, pkgMap map[string]*packages.Package, ignore *Ignore) error {
	cov.Packages = []*Package{}
	for _, profile := range profiles {
		pkgName := getPackageName(profile.FileName)
		pkgPkg := pkgMap[pkgName]
		if err := cov.parseProfile(profile, pkgPkg, ignore); err != nil {
			return err
		}
	}
	cov.LinesValid = cov.NumLines()
	cov.LinesCovered = cov.NumLinesWithHits()
	cov.LineRate = cov.HitRate()
	return nil
}

func (cov *Coverage) parseProfile(profile *Profile, pkgPkg *packages.Package, ignore *Ignore) error {
	if pkgPkg == nil || pkgPkg.Module == nil {
		return fmt.Errorf("package required when using go modules")
	}
	fileName := profile.FileName[len(pkgPkg.Module.Path)+1:]
	absFilePath := findAbsFilePath(pkgPkg, profile.FileName)
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, absFilePath, nil, 0)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(absFilePath)
	if err != nil {
		return err
	}

	if ignore.Match(fileName, data) {
		return nil
	}

	pkgPath, _ := filepath.Split(fileName)
	pkgPath = strings.TrimRight(strings.TrimRight(pkgPath, "/"), "\\")
	pkgPath = filepath.Join(pkgPkg.Module.Path, pkgPath)
	// TODO(boumenot): package paths are not file paths, there is a consistent separator
	pkgPath = strings.Replace(pkgPath, "\\", "/", -1)

	var pkg *Package
	for _, p := range cov.Packages {
		if p.Name == pkgPath {
			pkg = p
		}
	}
	if pkg == nil {
		pkg = &Package{Name: pkgPkg.ID, Classes: []*Class{}}
		cov.Packages = append(cov.Packages, pkg)
	}
	visitor := &fileVisitor{
		fset:     fset,
		fileName: fileName,
		fileData: data,
		classes:  make(map[string]*Class),
		pkg:      pkg,
		profile:  profile,
	}
	ast.Walk(visitor, parsed)
	pkg.LineRate = pkg.HitRate()
	return nil
}

type fileVisitor struct {
	fset     *token.FileSet
	fileName string
	fileData []byte
	pkg      *Package
	classes  map[string]*Class
	profile  *Profile
	byFiles  bool
}

func (v *fileVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		class := v.class(n)
		method := v.method(n)
		method.LineRate = method.Lines.HitRate()
		class.Methods = append(class.Methods, method)
		for _, line := range method.Lines {
			class.Lines = append(class.Lines, line)
		}
		class.LineRate = class.Lines.HitRate()
	}
	return v
}

func (v *fileVisitor) method(n *ast.FuncDecl) *Method {
	method := &Method{Name: n.Name.Name}
	method.Lines = []*Line{}

	start := v.fset.Position(n.Pos())
	end := v.fset.Position(n.End())
	startLine := start.Line
	startCol := start.Column
	endLine := end.Line
	endCol := end.Column
	// The blocks are sorted, so we can stop counting as soon as we reach the end of the relevant block.
	for _, b := range v.profile.Blocks {
		if b.StartLine > endLine || (b.StartLine == endLine && b.StartCol >= endCol) {
			// Past the end of the function.
			break
		}
		if b.EndLine < startLine || (b.EndLine == startLine && b.EndCol <= startCol) {
			// Before the beginning of the function
			continue
		}
		for i := b.StartLine; i <= b.EndLine; i++ {
			method.Lines.AddOrUpdateLine(i, int64(b.Count))
		}
	}
	return method
}

func (v *fileVisitor) class(n *ast.FuncDecl) *Class {
	var className string
	if v.byFiles {
		//className = filepath.Base(v.fileName)
		//
		// NOTE(boumenot): ReportGenerator creates links that collide if names are not distinct.
		// This could be an issue in how I am generating the report, but I have not been able
		// to figure it out.  The work around is to generate a fully qualified name based on
		// the file path.
		//
		// src/lib/util/foo.go -> src.lib.util.foo.go
		className = strings.Replace(v.fileName, "/", ".", -1)
		className = strings.Replace(className, "\\", ".", -1)
	} else {
		className = v.recvName(n)
	}
	class := v.classes[className]
	if class == nil {
		class = &Class{Name: className, Filename: v.fileName, Methods: []*Method{}, Lines: []*Line{}}
		v.classes[className] = class
		v.pkg.Classes = append(v.pkg.Classes, class)
	}
	return class
}

func (v *fileVisitor) recvName(n *ast.FuncDecl) string {
	if n.Recv == nil {
		return "-"
	}
	recv := n.Recv.List[0].Type
	start := v.fset.Position(recv.Pos())
	end := v.fset.Position(recv.End())
	name := string(v.fileData[start.Offset:end.Offset])
	return strings.TrimSpace(strings.TrimLeft(name, "*"))
}
