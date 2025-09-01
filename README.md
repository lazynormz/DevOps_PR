## What is this?
Because Microsoft recently broke their Azure DevOps boards in regards to their PR overview, I created this little TUI that gives a quick overview of open PRs and other quick-info.

## Why make this?
I was bored during the weekend and decided to use ai to create this slop of a codebase. Use at own risk (should be safe, no leaks as far as I can see)

## Requirements
Required permissions for the PAT:
- Code (Read)

## Running
If you want to run this application, either grab the latest release, or build from source

## Building
Built using Go version 1.25.0
```sh
go build -o bin/AzurePR.exe
```
