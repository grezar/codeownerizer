# codeownerizer
GitHub codeowners are required to have the push permission on the repo.
This tool is designed to automatically add code owners who are listed in the
CODEOWNERS file but don't have sufficient permissions to the repository, along
with the appropriate permission.

## GitHub Actions
```
name: Add ungranted codeowners to the repo

on:
  push:
    - main

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - name: Codeownerizer
      uses: grezar/codeownerizer@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
