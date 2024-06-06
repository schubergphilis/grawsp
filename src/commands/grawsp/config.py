from pathlib import Path

from cement import init_defaults

from .constants import APP_NAME
from .defaults import (
    DEFAULT_AWS_REGION,
    DEFAULT_RETRY_AFTER_IN_SECONDS,
    DEFAULT_TIMEOUT_IN_SECONDS,
)

DEFAULT_CONFIG = init_defaults(
    "aws",
    "database",
    "general",
    "user",
)

#
# AWS
#

DEFAULT_CONFIG["aws"]["credentials_path"] = (
    Path("~/.aws/credentials").expanduser().absolute().as_posix()
)
DEFAULT_CONFIG["aws"]["default_realm"] = ""
DEFAULT_CONFIG["aws"]["default_region"] = DEFAULT_AWS_REGION


#
# DATABASE
#

DEFAULT_CONFIG["database"]["path"] = (
    Path(f"~/.local/share/{APP_NAME}/{APP_NAME}.db").expanduser().absolute().as_posix()
)

#
# GENERAL
#

DEFAULT_CONFIG["general"]["firefox_path"] = ""
DEFAULT_CONFIG["general"]["retry_after"] = DEFAULT_RETRY_AFTER_IN_SECONDS
DEFAULT_CONFIG["general"]["timeout"] = DEFAULT_TIMEOUT_IN_SECONDS


#
# USER
#

DEFAULT_CONFIG["user"]["email"] = ""
DEFAULT_CONFIG["user"]["name"] = ""
