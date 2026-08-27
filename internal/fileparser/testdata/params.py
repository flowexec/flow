# f:name=test-params f:verb=test
# f:params=secretRef:my-secret:SECRET_VAR|prompt:"Enter name":NAME_VAR|text:default-value:DEFAULT_VAR

import os

print("Secret:", os.environ["SECRET_VAR"])
print("Name:", os.environ["NAME_VAR"])
print("Default:", os.environ["DEFAULT_VAR"])
