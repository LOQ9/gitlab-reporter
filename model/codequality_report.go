package model

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/mitchellh/hashstructure/v2"
)

type Report struct {
	EngineName        string         `json:"engine_name"`
	Fingerprint       string         `json:"fingerprint,omitempty"`
	Categories        []string       `json:"categories,omitempty"`
	CheckName         string         `json:"check_name"`
	Content           ReportContent  `json:"content,omitempty"`
	Description       string         `json:"description"`
	Location          ReportLocation `json:"location,omitempty"`
	OtherLocations    []interface{}  `json:"other_locations,omitempty"`
	RemediationPoints int            `json:"remediation_points,omitempty"`
	Severity          string         `json:"severity,omitempty"`
	Type              string         `json:"type"`
}

const (
	BugRisk       string = "Bug Risk"
	Clarity              = "Clarity"
	Compatibility        = "Compatibility"
	Complexity           = "Complexity"
	Security             = "Security"
	Style                = "Style"

	SeverityInfo     = "info"
	SeverityMinor    = "minor"
	SeverityMajor    = "major"
	SeverityCritical = "critical"
	SeverityBlocker  = "blocker"

	ReportTypeIssue = "issue"

	ReportEngineEslint = "eslint"
)

type ReportContent struct {
	Body string `json:"body"`
}

type ReportLocation struct {
	Path string `json:"path"`
	//Lines     ReportLocationLines     `json:"lines,omitempty"`
	Positions ReportLocationPositions `json:"positions,omitempty"`
}

type ReportLocationLines struct {
	Begin int `json:"begin"`
	End   int `json:"end"`
}

type ReportLocationPositions struct {
	Begin ReportLocationPositionsData `json:"begin"`
	End   ReportLocationPositionsData `json:"end"`
}

type ReportLocationPositionsData struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func NewReportFromCheckstyle(checkstyleReport *CheckStyleError, reportType string, reportEngine string, fileName string) *Report {
	newReport := &Report{
		EngineName: reportEngine,
		Type:       reportType,
		CheckName:  checkstyleReport.Source,
		Location: ReportLocation{
			Path: fileName,
			Positions: ReportLocationPositions{
				Begin: ReportLocationPositionsData{
					Line:   checkstyleReport.Line,
					Column: checkstyleReport.Column,
				},
				End: ReportLocationPositionsData{
					Line:   checkstyleReport.Line,
					Column: checkstyleReport.Column,
				},
			},
		},
		Description: checkstyleReport.Message,
	}

	newReport.SetDefaults()
	newReport.SetSeverity(checkstyleReport.Severity)
	newReport.SetCheckName()
	newReport.SetCategories()
	newReport.ComputeFingerprint()

	return newReport
}

func (r *Report) ToJSON() ([]byte, error) {
	e, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func ReportListToJSON(r []*Report) ([]byte, error) {
	e, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *Report) SetSeverity(severity string) {
	reportSeverity := strings.ToLower(severity)
	r.Severity = reportSeverity

	switch reportSeverity {
	case "info":
		r.Severity = SeverityInfo
	case "warning":
		r.Severity = SeverityMinor
	case "error":
		r.Severity = SeverityMajor
	}
}

func (r *Report) SetCheckName() {
	switch r.EngineName {
	case ReportEngineEslint:
		checkNameSplit := strings.Split(r.CheckName, "/")
		r.CheckName = checkNameSplit[len(checkNameSplit)-1]
	}
}

func (r *Report) SetCategories() {

	r.Categories = []string{Style}

	switch r.EngineName {
	case ReportEngineEslint:
		if eslintCategory[r.CheckName] != "" {
			r.Categories = []string{eslintCategory[r.CheckName]}
		}
	}
}

func (r *Report) ComputeFingerprint() {

	issueReport := Report{
		CheckName:   r.CheckName,
		Location:    ReportLocation{Path: r.Location.Path},
		Description: r.Description,
	}

	// Generate an hash of the reported problem
	hash, err := hashstructure.Hash(issueReport, hashstructure.FormatV2, nil)
	if err != nil {
		return
	}

	// Convert it to byte array and transform to md5
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(hash))

	hasher := md5.New()
	hasher.Write(b)

	switch r.EngineName {
	case ReportEngineEslint:
		if r.CheckName == Complexity {
			r.Fingerprint = hex.EncodeToString(hasher.Sum(nil))
		}

	}
}

func (r *Report) SetDefaults() {

	if r.Location.Positions.Begin.Line == 0 {
		r.Location.Positions.Begin.Line = 1
	}

	if r.Location.Positions.Begin.Column == 0 {
		r.Location.Positions.Begin.Column = 1
	}

	if r.Location.Positions.End.Line == 0 {
		r.Location.Positions.End.Line = 1
	}

	if r.Location.Positions.End.Column == 0 {
		r.Location.Positions.End.Column = 1
	}
}
