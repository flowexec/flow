# f:name=test-params f:verb=test
# f:params=secretRef:my-secret:SECRET_VAR|prompt:"Enter name":NAME_VAR|text:default-value:DEFAULT_VAR

Write-Host "Secret: $env:SECRET_VAR"
Write-Host "Name: $env:NAME_VAR"
Write-Host "Default: $env:DEFAULT_VAR"
