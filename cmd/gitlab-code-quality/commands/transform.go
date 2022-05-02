package commands

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"gitlab-code-quality/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// TransformCmd ...
var TransformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform report",
	RunE:  transformCmdF,
}

func init() {
	TransformCmd.Flags().StringSlice("source-report", []string{""}, "Source Report")
	TransformCmd.Flags().StringSlice("reporter-tool", []string{""}, "Reporter Tool")
	TransformCmd.Flags().StringSlice("report-type", []string{model.ReportTypeIssue}, "Report Type")
	TransformCmd.Flags().Bool("output", false, "Output")
	TransformCmd.Flags().Bool("debug", false, "Enables debug mode")
	TransformCmd.Flags().Bool("detect-report", false, "Automatically detect report files")
	TransformCmd.Flags().String("output-file", "", "Output File Name")
	RootCmd.AddCommand(TransformCmd)
}

func transformCmdF(command *cobra.Command, args []string) error {
	sourceReportArg, _ := command.Flags().GetStringSlice("source-report")
	reporterToolArg, _ := command.Flags().GetStringSlice("reporter-tool")
	reportTypeArg, _ := command.Flags().GetStringSlice("report-type")
	outputFileArg, _ := command.Flags().GetString("output-file")
	outputArg, _ := command.Flags().GetBool("output")
	detectReportArg, _ := command.Flags().GetBool("detect-report")
	//debugArg, _ := command.Flags().GetBool("debug")

	sourceReport := make([]string, len(sourceReportArg))
	copy(sourceReport, sourceReportArg)

	reporterTool := make([]string, len(sourceReport))
	copy(reporterTool, reporterToolArg)

	reportType := make([]string, len(sourceReport))
	copy(reportType, reportTypeArg)

	parsedReport := make([]*model.Report, 0)

	// Detect Report Files Automatically
	if detectReportArg {
		matches, err := filepath.Glob("*-*-checkstyle.xml")

		if err == nil && len(matches) > 0 {
			for _, fileName := range matches {
				splitFileName := strings.Split(fileName, "-")
				reporterTool = append(reporterTool, splitFileName[1])
				reportType = append(reportType, model.ReportTypeIssue)
				sourceReport = append(sourceReport, fileName)
			}
		}
	}

	for idx, report := range sourceReport {
		reportFromFile, err := os.ReadFile(report)
		if err != nil {
			return errors.New("specified source report was not found")
		}

		// Read our opened xmlFile as a byte array.
		byteValue, _ := ioutil.ReadAll(bytes.NewReader(reportFromFile))

		var result model.CheckStyleResult
		err = xml.Unmarshal(byteValue, &result)

		if err != nil {
			return errors.New("could not parse the provided file, it must be a xml checkstyle compliant")
		}

		// Assemble Gitlab report compatible structure
		for _, file := range result.Files {
			for _, fileError := range file.Errors {
				errorReport := model.NewReportFromCheckstyle(fileError, reportType[idx], reporterTool[idx], file.Name)
				parsedReport = append(parsedReport, errorReport)
			}
		}
	}

	jsonReport, _ := model.ReportListToJSON(parsedReport)

	if outputArg {
		fmt.Printf("%s\n", jsonReport)
	}

	if outputFileArg != "" {
		f, err := os.Create(outputFileArg)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err2 := f.Write(jsonReport)

		if err2 != nil {
			return err2
		}
	}

	return nil
}
