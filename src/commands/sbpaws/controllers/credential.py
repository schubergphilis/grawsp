import configparser
from datetime import datetime
from pathlib import Path

import humanize
from cement import Controller, ex
from inflection import dasherize
from sqlalchemy.orm import Session

from ....util.terminal.spinner import Spinner
from ..database.models import Credential


class CredentialController(Controller):
    class Meta:
        label = "credential"
        aliases = ["credentials", "creds"]
        stacked_on = "base"
        stacked_type = "nested"

    @ex(
        help="Configure AWS cli credentials file.",
        arguments=[
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
        ],
    )
    def configure(self) -> None:
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

    @ex(
        help="List all the valid credentials available.",
        arguments=[
            (
                ["--expired"],
                {
                    "action": "store_true",
                    "default": False,
                    "dest": "show_expired",
                    "help": "Include expired credentials in the output",
                },
            ),
        ],
    )
    def list(self) -> None:
        database_engine = self.app.database_engine
        show_expired = self.app.pargs.show_expired
        table_data = []

        with Session(database_engine) as session:
            credentials = session.query(Credential).all()

            for credential in credentials:
                if credential.is_expired() and not show_expired:
                    continue

                table_data.append(
                    [
                        credential.account.name,
                        credential.role_name,
                        credential.access_key_id,
                        humanize.naturaltime(
                            datetime.fromtimestamp(credential.expires_at),
                        ),
                    ],
                )

            if len(table_data) <= 0:
                self.app.log.warning("No credentials found.")
                return

            self.app.render(
                table_data,
                headers=[
                    "Account Name",
                    "Role",
                    "Access Key ID",
                    "Expires In",
                ],
            )
