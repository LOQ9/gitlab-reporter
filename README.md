# gitlab-code-quality
Gitlab Code Quality is a tool that merges multiple files from several code linters and outputs them combined using the code climate file format.  
This is allows to use the prefered linting tools and combining them with the Gitlab Code Quality Widget (https://docs.gitlab.com/ee/user/project/merge_requests/code_quality.html#code-quality-widget). 

## Usage & Examples

Currently this tool only supports files in the `checkstyle` format.
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

Generating a single report  
```
go run cmd/gitlab-code-quality/main.go transform --source-report sample/eslint-checkstyle.xml --reporter-tool eslint
```

Generating a combined report  
```
go run cmd/gitlab-code-quality/main.go transform --source-report sample/eslint-checkstyle.xml --source-report sample/golang-checkstyle.xml --reporter-tool eslint --reporter-tool golangci-lint
```
