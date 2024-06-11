import re

from cement import Controller
from inflection import transliterate
from sqlalchemy.orm import Session

from ....util.terminal.spinner import Spinner
from ..actions.aws import (
    create_authorization,
    create_credential,
    find_account_by_name,
    find_account_by_number,
    search_accounts,
)
from ..constants import APP_NAME
from ..database.models import SsoRole
from ..exceptions import RuntimeAppError


class AuthController(Controller):
    class Meta:
        label = "auth"
        stacked_on = "base"
        stacked_type = "nested"

        arguments = [
            (
                ["--from-role"],
                {
                    "default": "",
                    "help": "The name of the intermediary role to be assumed before.",
                    "dest": "from_role_name",
                    "type": str,
                },
            ),
            (
                ["--retry-after"],
                {
                    "help": "How long to wait until next AWS authorization API check",
                    "default": "",
                    "dest": "retry_after",
                },
            ),
            (
                ["--role"],
                {
                    "default": "",
                    "help": "The name of the role you want to assume.",
                    "dest": "role_name",
                    "type": str,
                },
            ),
            (
                ["--timeout"],
                {
                    "help": "How long to wait until we abort the authorization process",
                    "default": "",
                    "dest": "timeout",
                },
            ),
            (
                ["identifier"],
                {
                    "help": "The ID, name or regular expression identifying the account(s).",
                    "nargs": "?",
                },
            ),
        ]

    def _default(self) -> None:
        database_engine = self.app.database_engine
        from_role_name = self.app.pargs.from_role_name
        intermediary_role_name = ""
        realm_name = self.app.pargs.realm or self.app.config.get("aws", "default_realm")
        region = self.app.config.get("aws", "default_region")
        role_name = self.app.pargs.role_name

        retry_after = int(
            self.app.pargs.retry_after or self.app.config.get("general", "retry_after")
        )

        timeout = int(
            self.app.pargs.timeout or self.app.config.get("general", "timeout")
        )

        user_name = transliterate(
            re.sub(
                r"\s+",
                "",
                self.app.config.get("user", "name"),
                flags=re.UNICODE,
            ),
        )

        with Spinner("Accessing AWS Account") as spinner:
            if not realm_name:
                spinner.error("No AWS realm provided")
                raise RuntimeAppError()

            if not self.app.config.has_section(realm_name):
                spinner.error("No AWS realm configuration found")
                raise RuntimeAppError()

            start_url = self.app.config.get(realm_name, "start_url")

            if not start_url:
                spinner.error("No SSO start url was found")
                raise RuntimeAppError()

            client_name = APP_NAME
            spinner.info(f"Using {realm_name} realm")

            try:
                _ = create_authorization(
                    client_name=client_name,
                    database_engine=database_engine,
                    realm_name=realm_name,
                    region=region,
                    retry_after=retry_after,
                    start_url=start_url,
                    timeout=timeout,
                )
            except Exception as e:
                spinner.error("Could not authorize to AWS", submessage=str(e))
                raise RuntimeAppError("Could not authorize to AWS") from e
            else:
                spinner.success("Authorized to AWS")

            identifier = self.app.pargs.identifier

            if not identifier:
                return

            accounts = []

            if identifier.isdigit():
                account = find_account_by_number(
                    database_engine, realm_name, account_number=identifier
                )

                if account:
                    accounts.append(account)
            elif re.match(r"^[a-z0-9\-]+$", identifier):
                account = find_account_by_name(
                    database_engine, realm_name, account_name=identifier
                )

                if account:
                    accounts.append(account)
            else:
                accounts = search_accounts(
                    database_engine, realm_name, pattern=identifier
                )

            spinner.info(f"Identifier matched {len(accounts)} accounts")

            session_name = f"{client_name}-{user_name}"

            for account in accounts:
                if not role_name and self.app.config.has_option(
                    account.name, "default_role"
                ):
                    role_name = self.app.config.get(account.name, "default_role")

                if not role_name and self.app.config.has_option(
                    realm_name, "default_role"
                ):
                    role_name = self.app.config.get(realm_name, "default_role")

                if not role_name:
                    spinner.error("AWS role could not be determined")
                    raise RuntimeAppError()

                spinner.info(f"Using {role_name} role")

                intermediary_role_name = ""

                with Session(database_engine) as session:
                    sso_roles = [
                        role.name
                        for role in session.query(SsoRole)
                        .where(SsoRole.account_id == account.id)
                        .all()
                    ]

                if role_name not in sso_roles:
                    if self.app.config.has_option(account.name, "default_role"):
                        intermediary_role_name = self.app.config.get(
                            account.name, "default_role"
                        )

                    if self.app.config.has_option(realm_name, "default_role"):
                        intermediary_role_name = self.app.config.get(
                            realm_name, "default_role"
                        )

                    if from_role_name:
                        intermediary_role_name = from_role_name

                    if not intermediary_role_name:
                        spinner.error("Intermediary role could not be determined")
                        raise RuntimeAppError()

                    spinner.info(
                        f"Using {intermediary_role_name} as an intermediary role"
                    )

                try:
                    _ = create_credential(
                        database_engine=database_engine,
                        account_name=account.name,
                        realm_name=realm_name,
                        region=region,
                        role_name=role_name,
                        session_name=session_name,
                        intermediary_role_name=intermediary_role_name,
                    )
                except RuntimeError as e:
                    spinner.error(f"{e}")
                    raise RuntimeAppError() from e

                spinner.info(
                    f"Authorized to {account.name} account as {role_name} role"
                )
