package model

import (
	"testing"
)

const (
	reportFileName  = "eslint-checkstyle.xml"
	reportCheckName = "@typescript-eslint/no-unused-vars"
	checkName       = "no-unused-vars"
)

func newReport() *Report {
	NewReportFromCheckstyle(&CheckStyleError{
		Column:   0,
		Line:     0,
		Message:  "",
		Severity: BugRisk,
	}, ReportTypeIssue, ReportEngineEslint, reportFileName)

	r := Report{
		EngineName: ReportEngineEslint,
		CheckName:  reportCheckName,
		Categories: []string{eslintCategory[checkName]},
	}

	return &r
}

func TestSetCheckName(t *testing.T) {
	r := newReport()
	r.SetCheckName()

	if r.CheckName != checkName {
		t.Fail()
	}
}

func TestSetCategories(t *testing.T) {
	r := newReport()
	r.SetCheckName()
	r.SetCategories()

	for _, category := range r.Categories {
		if category != eslintCategory[r.CheckName] {
			t.Fail()
		}
	}
}
