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
    - name: Generate token
      id: generate_token
      uses: tibdex/github-app-token@v1
      with:
        app_id: ${{ secrets.APP_ID }}
        private_key: ${{ secrets.PRIVATE_KEY }}

    - name: Install codeownerizer
      uses: grezar/codeownerizer@v1

    - name: Grant
      run: codeownerizer
      env:
        GITHUB_TOKEN: ${{ steps.generate_token.outputs.token }}
```
