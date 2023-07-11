#!/bin/bash

go test -coverprofile=testcoverage
go tool cover -html=testcoverage -o testcoverage.html
go tool cover -func=testcoverage
echo Done generating test coverage reports!