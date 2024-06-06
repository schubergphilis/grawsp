import traceback

from cement import App

from .config import DEFAULT_CONFIG
from .constants import APP_NAME
from .controllers.account import AccountController
from .controllers.base import BaseController
from .controllers.console import ConsoleController
from .controllers.credential import CredentialController
from .exceptions import AppError
from .hooks import database_hook

#
# APP
#


class GrawspApp(App):
    class Meta:
        config_defaults = DEFAULT_CONFIG
        label = APP_NAME
        log_handler = "colorlog"
        output_handler = "tabulate"

        config_files = [
            f"~/.config/{APP_NAME}/{APP_NAME}.conf",
        ]

        extensions = [
            "alarm",
            "colorlog",
            "jinja2",
            "json",
            "tabulate",
            "yaml",
        ]

        handlers = [
            BaseController,
            AccountController,
            ConsoleController,
            CredentialController,
        ]

        hooks = [
            ("post_setup", database_hook),
        ]


#
# START
#


def run() -> None:
    with GrawspApp() as app:
        try:
            app.run()
        except AppError as e:
            if app.debug:
                print("")
                traceback.print_exception(e)

            exit(1)
        except Exception as e:
            print("")
            traceback.print_exception(e)
            exit(1)
        else:
            app.exit_code = 0


if __name__ == "__main__":
    run()
