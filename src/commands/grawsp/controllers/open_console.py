import re
import urllib
import webbrowser

from cement import Controller
from inflection import transliterate
from sqlalchemy.orm import Session

from ....services.aws.sts import get_console_url
from ....util.terminal.spinner import Spinner
from ..actions.aws import (
    create_credential,
    find_account_by_name,
    find_account_by_number,
    search_accounts,
)
from ..constants import APP_NAME
from ..database.models import SsoRole
from ..defaults import DEFAULT_TIMEOUT_IN_SECONDS
from ..exceptions import RuntimeAppError


class OpenConsoleController(Controller):
    class Meta:
        label = "open-console"
        stacked_on = "base"
        stacked_type = "nested"

        arguments = [
            (
                ["--region"],
                {
                    "default": "",
                    "help": "The name of the region you want the console to be opened with.",
                    "dest": "region",
                    "type": str,
                },
            ),
            (
                ["--role"],
                {
                    "default": "",
                    "help": "The name of the role you want to use.",
                    "dest": "role_name",
                    "type": str,
                },
            ),
            (
                ["identifier"],
                {
                    "help": "The ID, name or regular expression identifying the account(s).",
                },
            ),
        ]

    def _default(self) -> None:
        browser_name = "firefox-custom"
        database_engine = self.app.database_engine
        firefox_path = self.app.config.get("general", "firefox_path")
        identifier = self.app.pargs.identifier
        realm_name = self.app.pargs.realm or self.app.config.get("aws", "default_realm")
        region = self.app.pargs.region or self.app.config.get("aws", "default_region")
        role_name = self.app.pargs.role_name
        user_name = transliterate(
            re.sub(
                r"\s+",
                "",
                self.app.config.get("user", "name"),
                flags=re.UNICODE,
            ),
        )

        with Spinner("Opening AWS console(s)") as spinner:
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

            if len(accounts) <= 0:
                spinner.warning("Identifier matched no accounts")
                return

            webbrowser.register(
                browser_name, None, webbrowser.GenericBrowser([firefox_path, "%s"])
            )

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

                    if not intermediary_role_name:
                        spinner.error("Intermediary role could not be determined")
                        raise RuntimeAppError()

                    spinner.info(
                        f"Using {intermediary_role_name} as an intermediary role"
                    )

                session_name = f"{APP_NAME}-{user_name}-{role_name}"

                credential = create_credential(
                    database_engine=database_engine,
                    account_name=account.name,
                    realm_name=realm_name,
                    region=region,
                    role_name=role_name,
                    session_name=session_name,
                    intermediary_role_name=intermediary_role_name,
                )

                try:
                    console_url = get_console_url(
                        access_key_id=credential.access_key_id,
                        secret_access_key=credential.secret_access_key,
                        session_token=credential.session_token,
                        region=region,
                        timeout=DEFAULT_TIMEOUT_IN_SECONDS,
                    )
                except Exception as e:
                    spinner.error("Could not generate console URL", submessage=str(e))
                    raise RuntimeAppError() from e

                encoded_console_url = urllib.parse.quote(console_url)
                tab_color = "blue"

                webbrowser.get(browser_name).open_new_tab(
                    f"ext+container:name={account.name}&color={tab_color}&url={encoded_console_url}"
                )

                spinner.info(f"AWS console opened to {account.name}")

            spinner.success("All done")
