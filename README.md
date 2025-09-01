# AzurePR TUI
## 🤔 What is this?

<img width="1566" height="898" alt="image" src="https://github.com/user-attachments/assets/f934800f-0c22-480c-8916-d00839897726" />

Microsoft broke the PR overview in Azure DevOps boards (thanks, Satya 🙃), so I hacked together a terminal UI that shows you open PRs and some quick info.
Runs in your Windows terminal, just a single .exe.

## 🛠 Why does this exist?

Weekend boredom + AI-assisted chaos = this tool.
Use at your own risk. (Pretty sure it’s safe—no secrets leaking as far as I can tell.)

## 🔑 Requirements

You’ll need a [Personal Access Token (PAT)])(https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows) with:

- Code (Read)

## 🚀 Installation
Option 1: Download the release

1) Grab the latest .exe from [Releases](https://github.com/lazynormz/DevOps_PR/releases)

2) Drop it somewhere permanent (e.g. C:\Tools\AzurePR).

Option 2: Build it yourself

```
go build -o bin/AzurePR.exe
```

### Add it to PATH (recommended)

So you can just type AzurePR anywhere in your terminal:

1) Put AzurePR.exe in a folder like C:\Tools\AzurePR.

2) Add that folder to your System PATH:

    - Windows Search → “Edit system environment variables”

     - Environment Variables → Path → Edit → Add C:\Tools\AzurePR

Now you can run:
```
AzurePR
```

### (Optional) Symlink

If you don’t want to mess with PATH every time:
```
mklink "C:\Windows\System32\AzurePR.exe" "C:\Tools\AzurePR\AzurePR.exe"
```
⚠️ Needs admin rights.

## 🏗 Building

Requires Go 1.25.0

Build with:
```
go build -o bin/AzurePR.exe
```

# ⚠️ Notes

Only tested on Windows for now.

Expect some rough edges—it’s more “weekend hack” than “production-ready.”
