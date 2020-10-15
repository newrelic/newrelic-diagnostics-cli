#!/bin/sh
go build -o grapher scripts/dependencyGrapher/main.go &&
./grapher | python3 scripts/dependencyGrapher/generateGraph.py &&
rm grapher