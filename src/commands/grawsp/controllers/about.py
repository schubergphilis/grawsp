from cement import Controller
from prompt_toolkit import HTML, print_formatted_text
from prompt_toolkit.styles import Style


class AboutController(Controller):
    class Meta:
        label = "about"
        stacked_on = "base"
        stacked_type = "nested"

    def _default(self) -> None:
        print_formatted_text(
            HTML("Developed by <bullet>\\</bullet> <message>Schuberg Philis</message>"),
            style=Style.from_dict(
                {
                    "bullet": "cyan bold",
                    "message": "white",
                },
            ),
        )
        print_formatted_text(
            HTML("<message>Apache License Version 2.0</message>"),
            style=Style.from_dict(
                {
                    "bullet": "cyan bold",
                    "message": "gray italic",
                },
            ),
        )
