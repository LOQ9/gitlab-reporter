package commands

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/LOQ9/gitlab-reporter/model"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CodeQualityCmd ...
var CodeQualityCmd = &cobra.Command{
	Use:   "codequality",
	Short: "Code quality report",
	RunE:  codeQualityCmdF,
}

type CodeQualityCommand struct {
	sourceReport   []string
	reporterEngine []string
	reportType     []string
	outputFile     string
	outputArg      bool
	detectReport   bool
}

func NewCodeQualityCommand(flags *pflag.FlagSet) *CodeQualityCommand {
	codeQualityCommand := CodeQualityCommand{}

	codeQualityCommand.sourceReport, _ = flags.GetStringSlice("source-report")
	codeQualityCommand.reporterEngine, _ = flags.GetStringSlice("reporter-tool")
	codeQualityCommand.reportType, _ = flags.GetStringSlice("report-type")
	codeQualityCommand.outputFile, _ = flags.GetString("output-file")
	codeQualityCommand.outputArg, _ = flags.GetBool("output")
	codeQualityCommand.detectReport, _ = flags.GetBool("detect-report")

	return &codeQualityCommand
}

func (t *CodeQualityCommand) FindReport(reportLocation string) []string {
	matches, _ := filepath.Glob("*-checkstyle.xml")
	return matches
}

func (t *CodeQualityCommand) AddReport(reportFile string, reportType string, reportEngine string) *CodeQualityCommand {
	t.reporterEngine = append(t.reporterEngine, reportEngine)
	t.reportType = append(t.reportType, reportType)
	t.sourceReport = append(t.sourceReport, reportFile)

	return t
}

func (t *CodeQualityCommand) CreateFile(fileData []byte) error {
	f, errCreate := os.Create(t.outputFile)

	if errCreate != nil {
		return errCreate
	}

	defer f.Close()

	_, errWrite := f.Write(fileData)

	if errWrite != nil {
		return errWrite
	}

	return nil
}

func init() {
	CodeQualityCmd.Flags().StringSlice("source-report", []string{""}, "Source Report")
	CodeQualityCmd.Flags().StringSlice("reporter-tool", []string{""}, "Reporter Tool")
	CodeQualityCmd.Flags().StringSlice("report-type", []string{model.ReportTypeIssue}, "Report Type")
	CodeQualityCmd.Flags().Bool("output", true, "Output")
	CodeQualityCmd.Flags().Bool("debug", false, "Enables debug mode")
	CodeQualityCmd.Flags().Bool("detect-report", true, "Automatically detect report files")
	CodeQualityCmd.Flags().String("output-file", "", "Output File Name")
	RootCmd.AddCommand(CodeQualityCmd)
}

func codeQualityCmdF(command *cobra.Command, args []string) error {
	transformCommand := NewCodeQualityCommand(command.Flags())

	parsedReport := make([]*model.Report, 0)

	// Detect Report Files Automatically
	if transformCommand.detectReport {
		for _, reportFile := range transformCommand.FindReport("*-checkstyle.xml") {
			fmt.Printf("Detected report: file (%s)\n", reportFile)
			// Spliting will only work when the regex is appropriate
			splitFileName := strings.Split(reportFile, "-")
			transformCommand = transformCommand.AddReport(reportFile, model.ReportTypeIssue, splitFileName[0])
		}
	}

	for idx, report := range transformCommand.sourceReport {
		fmt.Printf("Using report: file (%s) type (%s) engine (%s)\n", report, transformCommand.reportType[idx], transformCommand.reporterEngine[idx])

		reportFromFile, err := os.ReadFile(report)
		if err != nil {
			return errors.New("specified source report was not found")
		}

		// Read our opened xmlFile as a byte array.
		byteValue, _ := ioutil.ReadAll(bytes.NewReader(reportFromFile))

		var result model.CheckStyleResult
		err = xml.Unmarshal(byteValue, &result)

		if err != nil {
			return errors.New("could not parse the provided file, it must be xml checkstyle compliant")
		}

		// Assemble Gitlab report compatible structure
		for _, file := range result.Files {
			for _, fileCheckStyleError := range file.Errors {
				parsedReport = append(parsedReport, model.NewReportFromCheckstyle(fileCheckStyleError, transformCommand.reportType[idx], transformCommand.reporterEngine[idx], file.Name))
			}
		}
	}

	jsonReport, _ := model.ReportListToJSON(parsedReport)

	if transformCommand.outputArg {
		fmt.Printf("%s\n", jsonReport)
	}

	if transformCommand.outputFile != "" {
		if err := transformCommand.CreateFile(jsonReport); err != nil {
			return err
		}

		fmt.Printf("Report created at: %s\n", transformCommand.outputFile)
	}

	return nil
}
