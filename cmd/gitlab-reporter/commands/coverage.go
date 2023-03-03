package commands

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/LOQ9/gitlab-reporter/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/tools/go/packages"
)

// CoverageCmd ...
var CoverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "Coverage report",
	RunE:  coverageCmdF,
}

type CoverageCommand struct {
	// sourceReport   string
	byFiles        bool
	ignoreGenFiles bool
	ignoreDirs     string
	ignoreFiles    string
}

func NewCoverageCommand(flags *pflag.FlagSet) *CoverageCommand {
	coverageCommand := CoverageCommand{}
	// coverageCommand.sourceReport, _ = flags.GetString("source-report")
	coverageCommand.byFiles, _ = flags.GetBool("by-files")
	coverageCommand.ignoreGenFiles, _ = flags.GetBool("ignore-gen-files")
	coverageCommand.ignoreDirs, _ = flags.GetString("ignore-dirs")
	coverageCommand.ignoreFiles, _ = flags.GetString("ignore-files")

	return &coverageCommand
}

func init() {
	CoverageCmd.Flags().Bool("by-files", false, "code coverage by file, not class")
	CoverageCmd.Flags().Bool("ignore-gen-files", false, "ignore generated files")
	CoverageCmd.Flags().String("ignore-dirs", "", "ignore dirs matching this regexp")
	CoverageCmd.Flags().String("ignore-files", "", "ignore files matching this regexp")
	RootCmd.AddCommand(CoverageCmd)
}

func coverageCmdF(command *cobra.Command, args []string) error {
	coverageCommand := NewCoverageCommand(command.Flags())

	var err error
	var ignore model.Ignore
	if coverageCommand.ignoreDirs != "" {
		ignore.Dirs, err = regexp.Compile(coverageCommand.ignoreDirs)
		if err != nil {
			return errors.Wrap(err, "Bad -ignore-dirs regexp")
		}
	}

	if coverageCommand.ignoreFiles != "" {
		ignore.Files, err = regexp.Compile(coverageCommand.ignoreFiles)
		if err != nil {
			return errors.Wrap(err, "Bad -ignore-files regexp")
		}
	}

	if err := convert(os.Stdin, os.Stdout, &ignore); err != nil {
		return errors.Wrap(err, "code coverage conversion failed")
	}

	return nil
}

func convert(in io.Reader, out io.Writer, ignore *model.Ignore) error {
	profiles, err := model.ParseProfiles(in, ignore)
	if err != nil {
		return err
	}

	pkgs, err := model.GetPackages(profiles)
	if err != nil {
		return err
	}

	sources := make([]*model.Source, 0)
	pkgMap := make(map[string]*packages.Package)
	for _, pkg := range pkgs {
		if pkg.Module == nil {
			continue
		}

		sources = model.AppendIfUnique(sources, pkg.Module.Dir)
		pkgMap[pkg.ID] = pkg
	}

	coverage := model.Coverage{Sources: sources, Packages: nil, Timestamp: time.Now().UnixNano() / int64(time.Millisecond)}
	if err := coverage.ParseProfiles(profiles, pkgMap, ignore); err != nil {
		return err
	}

	_, _ = fmt.Fprint(out, xml.Header)
	_, _ = fmt.Fprintln(out, model.CoberturaDTDDecl)

	encoder := xml.NewEncoder(out)
	encoder.Indent("", "  ")
	if err := encoder.Encode(coverage); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out)
	return nil
}
