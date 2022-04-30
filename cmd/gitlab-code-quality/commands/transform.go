package commands

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"gitlab-code-quality/model"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/hashstructure/v2"
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
	TransformCmd.Flags().Bool("output", false, "Output")
	TransformCmd.Flags().Bool("debug", false, "Enables debug mode")
	TransformCmd.Flags().String("output-file", "gl-code-quality-report.json", "Output File Name")
	RootCmd.AddCommand(TransformCmd)
}

func transformCmdF(command *cobra.Command, args []string) error {
	sourceReportArg, _ := command.Flags().GetStringSlice("source-report")
	reporterToolArg, _ := command.Flags().GetStringSlice("reporter-tool")
	reportTypeArg, _ := command.Flags().GetStringSlice("report-type")
	outputFileArg, _ := command.Flags().GetString("output-file")
	outputArg, _ := command.Flags().GetBool("output")
	//debugArg, _ := command.Flags().GetBool("debug")

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
		err = xml.Unmarshal(byteValue, &result)

		if err != nil {
			return errors.New("could not parse the provided file, it must be a xml checkstyle compliant")
		}

		// Assemble Gitlab report compatible structure
		for _, file := range result.Files {
			for _, fileError := range file.Errors {
				errorReport := &model.Report{
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
				}

				errorReport.CheckName = errorReport.GetCheckName()
				errorReport.Categories = errorReport.GetCategories()

				// Generate an hash of the reported problem
				hash, err := hashstructure.Hash(errorReport, hashstructure.FormatV2, nil)
				if err != nil {
					return err
				}

				// Convert it to byte array and transform to md5
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(hash))

				h := sha1.New()
				h.Write(b)
				errorReport.Fingerprint = hex.EncodeToString(h.Sum(nil))

				// hasher := md5.New()
				// hasher.Write(b)
				// errorReport.Fingerprint = hex.EncodeToString(hasher.Sum(nil))

				parsedReport = append(parsedReport, errorReport)
			}
		}
	}

	jsonReport, _ := model.ReportListToJSON(parsedReport)
	fmt.Printf("%s\n", jsonReport)

	if outputArg {
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
