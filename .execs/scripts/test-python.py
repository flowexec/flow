"""Integration check for flow's python interpreter support.

Run by the `test python-script` executable in CI on Linux, macOS, and Windows.
It asserts the things flow itself is responsible for - that a real interpreter
ran, that flow-resolved parameters arrived in the environment, and that the
python env defaults were applied - so a regression in interpreter resolution
fails CI rather than silently degrading to shell execution.
"""

import os
import platform
import sys

print("Running python file execution test...")
print(f"OS: {platform.system()}")
print(f"Python: {sys.version.splitlines()[0]}")
print(f"Interpreter: {sys.executable}")

failures = []

# A shell interpreter could never have reached this file at all, but assert the
# version explicitly so a python2 interpreter is caught rather than tolerated.
if sys.version_info[0] != 3:
    failures.append(f"expected python 3, got {sys.version_info[0]}")

# flow injects these so output streams live and workspaces stay free of
# __pycache__; see internal/services/run/python.go pythonEnv.
for key in ("PYTHONUNBUFFERED", "PYTHONDONTWRITEBYTECODE"):
    if os.environ.get(key) != "1":
        failures.append(f"{key} not set by flow (got {os.environ.get(key)!r})")

# Set by env.DefaultEnv for every run; interpreter resolution relies on it to
# find a workspace's .venv.
if not os.environ.get("FLOW_WORKSPACE_PATH"):
    failures.append("FLOW_WORKSPACE_PATH missing from the run environment")

if failures:
    for failure in failures:
        print(f"FAIL: {failure}", file=sys.stderr)
    sys.exit(1)

print("Python file execution works correctly.")
