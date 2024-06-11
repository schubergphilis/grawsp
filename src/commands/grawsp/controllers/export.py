import configparser
from datetime import datetime
from pathlib import Path

from cement import Controller
from inflection import dasherize
from sqlalchemy.orm import Session

from ....util.terminal.spinner import Spinner
from ..database.models import Credential


class ExportController(Controller):
    class Meta:
        label = "export"
        stacked_on = "base"
        stacked_type = "nested"

        arguments = [
            (
                ["--default-account"],
                {
                    "default": "",
                    "help": "The name of the account that should be set as default.",
                    "dest": "default_account",
                    "type": str,
                },
            ),
            (
                ["--default-role"],
                {
                    "default": "",
                    "help": "The name of the role that should be set as default.",
                    "dest": "default_role",
                    "type": str,
                },
            ),
            (
                ["--path"],
                {
                    "default": "~/.aws/credentials",
                    "help": "The path of the AWS cli credentials file.",
                    "dest": "credentials_file_path",
                    "type": str,
                },
            ),
        ]

    def _default(self) -> None:
        credentials_file_path = Path(
            self.app.pargs.credentials_file_path
            or self.app.config.get("aws", "credentials_path")
        ).expanduser()

        credentials_parser = configparser.ConfigParser()
        database_engine = self.app.database_engine
        default_account = self.app.pargs.default_account
        default_role = self.app.pargs.default_role

        with Spinner("Configuring AWS credentials file") as spinner:
            credentials_file_path.parent.mkdir(exist_ok=True, parents=True)
            credentials_file_path.touch()

            spinner.info(
                f"Using credentials file at {credentials_file_path.as_posix()}"
            )

            with Session(database_engine) as session:
                credentials = (
                    session.query(Credential)
                    .where(Credential.expires_at > datetime.now().timestamp())
                    .all()
                )

                if len(credentials) <= 0:
                    spinner.warning("No valid credentials found")
                    return

                for credential in credentials:
                    profile_name = dasherize(
                        f"{credential.account.name}-{credential.role_name}"
                    ).lower()

                    credentials_parser[profile_name] = {
                        "aws_access_key_id": credential.access_key_id,
                        "aws_secret_access_key": credential.secret_access_key,
                        "aws_session_token": credential.session_token,
                    }

                    spinner.info(
                        f"Adding credentials for {credential.account.name} as {credential.role_name} role"
                    )

                    if (
                        default_account == credential.account.name
                        and default_role == credential.role_name
                    ):
                        credentials_parser["default"] = {
                            "aws_access_key_id": credential.access_key_id,
                            "aws_secret_access_key": credential.secret_access_key,
                            "aws_session_token": credential.session_token,
                        }

                        spinner.info(
                            f"Default profile set to {credential.account.name} account and {credential.role_name} role"
                        )

            with open(credentials_file_path.as_posix(), "w") as fd:
                credentials_parser.write(fd)

            spinner.success("Credentials file was configured")
