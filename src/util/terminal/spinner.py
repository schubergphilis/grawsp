from __future__ import annotations

from prompt_toolkit import HTML, print_formatted_text
from prompt_toolkit.styles import Style
from yaspin import yaspin


class Spinner:
    def __init__(self, message: str) -> None:
        self._spinner = yaspin(
            text=message,
            color="cyan",
        )

    def __enter__(self) -> Spinner:
        self._spinner.start()
        return self

    def __exit__(self, type, value, traceback) -> None:
        self._spinner.stop()

    @property
    def message(self) -> str:
        return self._spinner.text

    @message.setter
    def message(self, value: str) -> None:
        self._spinner.text = value

    def info(self, message: str, submessage: str = "") -> None:
        with self._spinner.hidden():
            print_formatted_text(
                HTML(
                    f"<bullet>\\</bullet> <message>{message}</message>"
                    + (f" <submessage>{submessage}</submessage>" if submessage else ""),
                ),
                style=Style.from_dict(
                    {
                        "bullet": "cyan bold",
                        "message": "white",
                        "submessage": "gray italic",
                    },
                ),
            )

    def warning(self, message: str, submessage: str = "") -> None:
        with self._spinner.hidden():
            print_formatted_text(
                HTML(
                    f"<bullet>\\</bullet> <message>{message}</message>"
                    + (f" <submessage>{submessage}</submessage>" if submessage else ""),
                ),
                style=Style.from_dict(
                    {
                        "bullet": "yellow bold",
                        "message": "yellow",
                        "submessage": "gray italic",
                    },
                ),
            )

    def error(self, message: str, submessage: str = "") -> None:
        with self._spinner.hidden():
            print_formatted_text(
                HTML(
                    f"<bullet>\\</bullet> <message>{message}</message>"
                    + (f" <submessage>{submessage}</submessage>" if submessage else ""),
                ),
                style=Style.from_dict(
                    {
                        "bullet": "red bold",
                        "message": "red",
                        "submessage": "gray italic",
                    },
                ),
            )

    def success(self, message: str = "") -> None:
        self._spinner.color = "green"

        if message:
            self._spinner.text = message

        self._spinner.ok("✔")

    def fail(self, message: str = "") -> None:
        self._spinner.color = "red"

        if message:
            self._spinner.text = message

        self._spinner.fail("⚠")
