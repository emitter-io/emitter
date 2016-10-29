$isPr = "No";
if ($env:APPVEYOR_PULL_REQUEST_NUMBER)
{
    $isPr = "Yes = PR #: " + $env:APPVEYOR_PULL_REQUEST_NUMBER
}

Write-Host "  *** AppVeyor configuration information ***" -ForegroundColor Yellow
"      Account Name: " + $env:APPVEYOR_ACCOUNT_NAME
Write-Host "           Version: $env:APPVEYOR_BUILD_VERSION" -ForegroundColor Red
"          Git Repo"
"            -    Name: " + $env:APPVEYOR_REPO_NAME
"            -  Branch: " + $env:APPVEYOR_REPO_BRANCH  
Write-Host "            -    Info: $env:APPVEYOR_REPO_COMMIT / $env:APPVEYOR_REPO_COMMIT_AUTHOR / $env:APPVEYOR_REPO_COMMIT_TIMESTAMP" -ForegroundColor Gray
Write-Host "                       '$env:APPVEYOR_REPO_COMMIT_MESSAGE'" -ForegroundColor Gray
"            - Is a PR: " + $isPr  
"             Platform: " + $env:PLATFORM
"        Configuration: " + $env:CONFIGURATION
"  --------------------------------------------------------------------------"
""