# Scripts

Scripts are an additional datasource for information that isn't collected by a task.

## Script output

Scripts should output to stdout. All output to stdout is collected and saved in a file based on the name of the script, ie: `name-of-script.out`. This is saved in the directory specified by `-output-path`, defaulting to the current directory.

Scripts can also output files. Files should be created in the current working directory. If `-output-path` is used, the current working directory will be set to the path specified. All output files are included in the results zip in the `ScriptOutput/` directory. 

## Script results

The results of running a script can be found in the `nrdiag-output.json` file with the following schema:

```json
"Script": {
    "Name": "example",
    "Description": "Example Description",
    "Output": "example output",
    "OutputFiles": [
        "/path/to/example.out",
        "/path/to/another-file.out"
    ],
    "OutputTruncated": false
}
```

The `Output` field contains the stdout output. If it is over 20000 characters, it is truncated and the `OutputTruncated` field is set to `true`. Even if trucated, the full output is still available in the `ScriptOutput/` directory in the zip file.

A list of files the script created can be found in the `Outputfiles` field. 

## Adding to the script catalog

Contributes to the script catalog are always welcome. Before contributing please read the
[code of conduct](./../CODE_OF_CONDUCT.md). To contribute,
[fork](https://help.github.com/articles/fork-a-repo/) this repository, commit your changes, and [send a Pull Request](https://help.github.com/articles/using-pull-requests/).

### File structure

Scripts are located in `scriptcatalog/scripts/`. Each script must also have a yml file associated with it, located in `scriptcatalog/`.

### Yaml schema definition

```yml
# Unique name
# Example:
# name: parse-dotnet-instrumentation
name: string, required

# Script filename
# Example:
# filename: parse-dotnet-instrumentation.sh
filename: string, required

# Description of script functionality
# Example:
# description: Parses .NET Agent custom instrumentation XML files
description: string, required

# Scripting language used
# Example:
# type: bash
type: string, required

# Supported operating systems
# Example:
# os: linux, darwin
os: string (enum), required # linux, darwin, windows

# Files output by the script
# Example:
# outputFiles:
#   - additional_logs_*.zip
outputFiles: list (string), optional # list of files the script creates. Wildcard * supported
```
