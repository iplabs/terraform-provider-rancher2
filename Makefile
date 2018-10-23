VERSION=`cat ./VERSION.txt`

default: build

clean:
	@rm -rf ./vendor/
	@rm -rf ./.bin/
	@rm -rf ./.scannerwork/
	@rm -f ./*.out

prepare:
	@mkdir -p ./.bin
	@echo Downloading Go depdendencies...
	@go mod download

build: prepare
	@echo Building v${VERSION}...
	@go install .

test: prepare
	@echo Performing unit tests...
	@go test ./... --cover -json --coverprofile=./coverage.out > ./test-report.out

sonar: test
	@sonar-scanner -Dsonar.projectVersion=${VERSION}

dist: prepare
	@mkdir -p "./.bin/linux_amd64"
	@mkdir -p "./.bin/darwin_amd64"
	@mkdir -p "./.bin/windows_amd64"
	@GOOS=linux GOARCH=amd64 go build -o ./.bin/linux_amd64/terraform-provider-rancher2_v${VERSION}_x4 .
	@GOOS=darwin GOARCH=amd64 go build -o ./.bin/darwin_amd64/terraform-provider-rancher2_v${VERSION}_x4 .
	@GOOS=windows GOARCH=amd64 go build -o ./.bin/windows_amd64/terraform-provider-rancher2_v${VERSION}_x4.exe .
