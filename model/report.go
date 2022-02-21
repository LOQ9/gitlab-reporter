package model

import "encoding/json"

type Report struct {
	EngineName        string         `json:"engine_name"`
	Fingerprint       string         `json:"fingerprint"`
	Categories        []string       `json:"categories"`
	CheckName         string         `json:"check_name"`
	Content           ReportContent  `json:"content,omitempty"`
	Description       string         `json:"description"`
	Location          ReportLocation `json:"location,omitempty"`
	OtherLocations    []interface{}  `json:"other_locations,omitempty"`
	RemediationPoints int            `json:"remediation_points"`
	Severity          string         `json:"severity"`
	Type              string         `json:"type"`
}

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
