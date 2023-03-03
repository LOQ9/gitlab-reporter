# gitlab-reporter (Previously gitlab-code-quality)
Gitlab Reporter is a tool that aims to provide the necessary toolset, like a Swiss army knife for generating the required output used by the Gitlab Widgets.

## Usage & Examples

### Coverage

The coverage report generation is based on the implementation available at https://github.com/boumenot/gocover-cobertura  

### Code Quality

Currently it merges multiple files from several code linters and outputs them combined using the code climate file format.
This allows the use of preferred linting tools and combining them with the Gitlab Code Quality Widget 
(https://docs.gitlab.com/ee/user/project/merge_requests/code_quality.html#code-quality-widget). 

This tool only supports files in the `checkstyle` format.
For javascript projects using eslint the flag `--format=checkstyle` is required:  

Example:  
```
npx eslint --format=checkstyle --ext .ts src/ -c .eslintrc.js
```

For golang projects using golang-ci the flag `--out-format checkstyle` is required:  

Example:  
```
golangci-lint --out-format checkstyle run ./...
```

Generating a single code quality report  
```
go run cmd/gitlab-reporter/main.go codequality --source-report sample/eslint-checkstyle.xml --reporter-tool eslint
```

Generating a combined code quality report  
```
go run cmd/gitlab-reporter/main.go codequality --source-report sample/eslint-checkstyle.xml --source-report sample/golang-checkstyle.xml --reporter-tool eslint --reporter-tool golangci-lint
```
