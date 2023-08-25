# Adding to the script catalog

## File structure

Scripts are located in `scriptcatalog/scripts/`. Each script must also have a yml file associated with it, located in `scriptcatalog/`.

## Yaml schema definition

```yml
# Unique name
# Example: parse-dotnet-instrumentation
name: string, required

# Script filename
# Example: parse-dotnet-instrumentation.sh
filename: string, required

# Description of script functionality
# Example: Parses .NET Agent custom instrumentation XML files
description: string, required

# Scripting language used
# Example: bash, powershell
type: string, required

# Supported operating systems
# Example: linux
os: string (enum), required # linux, darwin, windows
```
