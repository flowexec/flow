REM f:name=test-params f:verb=test
REM f:params=secretRef:my-secret:SECRET_VAR|prompt:"Enter name":NAME_VAR|text:default-value:DEFAULT_VAR
@echo off

echo Secret: %SECRET_VAR%
echo Name: %NAME_VAR%
echo Default: %DEFAULT_VAR%
