from cement import Controller


class BaseController(Controller):
    class Meta:
        label = "base"

        arguments = [
            (
                ["--realm"],
                {
                    "help": "Which AWS realm to use",
                    "default": "",
                    "dest": "realm",
                },
            )
        ]
