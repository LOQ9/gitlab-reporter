package model

import (
	"encoding/json"
	"strings"
)

type Report struct {
	EngineName        string         `json:"engine_name"`
	Fingerprint       string         `json:"fingerprint"`
	Categories        []string       `json:"categories,omitempty"`
	CheckName         string         `json:"check_name"`
	Content           ReportContent  `json:"content,omitempty"`
	Description       string         `json:"description"`
	Location          ReportLocation `json:"location,omitempty"`
	OtherLocations    []interface{}  `json:"other_locations,omitempty"`
	RemediationPoints int            `json:"remediation_points"`
	Severity          string         `json:"severity"`
	Type              string         `json:"type"`
}

const (
	BugRisk       string = "Bug Risk"
	Clarity              = "Clarity"
	Compatibility        = "Compatibility"
	Complexity           = "Complexity"
	Security             = "Security"
	Style                = "Style"
)

type ReportContent struct {
	Body string `json:"body"`
}

type ReportLocation struct {
	Path  string              `json:"path"`
	Lines ReportLocationLines `json:"lines"`
}

type ReportLocationLines struct {
	Begin int `json:"begin"`
	End   int `json:"end"`
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

func (r *Report) GetCheckName() string {
	switch r.EngineName {
	case "eslint":
		checkNameSplit := strings.Split(r.CheckName, "/")
		r.CheckName = checkNameSplit[len(checkNameSplit)-1]
	}

	return r.CheckName
}

func (r *Report) GetCategories() []string {

	r.Categories = []string{Style}

	switch r.EngineName {
	case "eslint":
		if eslintCategory[r.CheckName] != "" {
			r.Categories = []string{eslintCategory[r.CheckName]}
		}
	}

	return r.Categories
}
