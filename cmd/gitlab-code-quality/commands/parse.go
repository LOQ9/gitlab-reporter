package commands

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"gitlab-code-quality/model"
	"io/ioutil"
	"os"
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
	TransformCmd.Flags().StringSlice("report-type", []string{"issue"}, "Report Type")
	RootCmd.AddCommand(TransformCmd)
}

func transformCmdF(command *cobra.Command, args []string) error {
	sourceReportArg, _ := command.Flags().GetStringSlice("source-report")
	reporterToolArg, _ := command.Flags().GetStringSlice("reporter-tool")
	reportTypeArg, _ := command.Flags().GetStringSlice("report-type")

	reporterTool := make([]string, len(sourceReportArg))
	copy(reporterTool, reporterToolArg)

	reportType := make([]string, len(sourceReportArg))
	copy(reportType, reportTypeArg)

	parsedReport := make([]*model.Report, 0)

	for idx, sourceReport := range sourceReportArg {
		reportFromFile, err := os.ReadFile(sourceReport)
		if err != nil {
			return errors.New("specified source report was not found")
		}

		// Read our opened xmlFile as a byte array.
		byteValue, _ := ioutil.ReadAll(bytes.NewReader(reportFromFile))

		var result model.CheckStyleResult
		xml.Unmarshal(byteValue, &result)

		// Assemble Gitlab report compatible structure
		for _, file := range result.Files {
			for _, fileError := range file.Errors {
				parsedReport = append(parsedReport, &model.Report{
					EngineName: reporterTool[idx],
					Type:       reportType[idx],
					CheckName:  fileError.Source,
					Location: model.ReportLocation{
						Path: file.Name,
						Lines: model.ReportLocationLines{
							Begin: fileError.Line,
						},
					},
					Severity:    strings.ToLower(fileError.Severity),
					Description: fileError.Message,
				})
			}
		}
	}

	jsonReport, _ := model.ReportListToJSON(parsedReport)
	fmt.Printf("%s\n", jsonReport)

	return nil
}
