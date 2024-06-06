import re

from cement import Controller, ex
from inflection import transliterate
from sqlalchemy import delete
from sqlalchemy.orm import Session

from ....services.aws.sso import list_sso_accounts_with_roles
from ....util.terminal.spinner import Spinner
from ..actions.aws import (
    create_credential,
    find_account_by_name,
    find_account_by_number,
    find_authorization,
    find_realm,
    search_accounts,
)
from ..constants import APP_NAME
from ..database.models import Account, Realm, SsoRole
from ..exceptions import RuntimeAppError


class AccountController(Controller):
    class Meta:
        label = "account"
        aliases = ["accounts"]
        stacked_on = "base"
        stacked_type = "nested"

    @ex(
        help="Get credentials to access an account as a role",
        aliases=["auth"],
        arguments=[
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
                ["identifier"],
                {
                    "help": "The ID, name or regular expression identifying the account(s).",
                },
            ),
        ],
    )
    def authorize(self) -> None:
        database_engine = self.app.database_engine
        intermediary_role_name = ""
        realm_name = self.app.pargs.realm or self.app.config.get("aws", "default_realm")
        region = self.app.config.get("aws", "default_region")
        role_name = self.app.pargs.role_name
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

            spinner.info(f"Using {realm_name} realm")

            accounts = []
            identifier = self.app.pargs.identifier

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

            session_name = f"{APP_NAME}-{user_name}-{role_name}"

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

    @ex(
        help="List all the accounts of a realm",
        arguments=[
            (
                ["--pattern"],
                {
                    "default": "^.*$",
                    "help": "A regular expression for the account name.",
                    "dest": "pattern",
                    "type": str,
                },
            ),
        ],
    )
    def list(self) -> None:
        database_engine = self.app.database_engine
        pattern = self.app.pargs.pattern

        with Session(database_engine) as session:
            accounts_with_realms = (
                session.query(Account, Realm)
                .join(Realm, Realm.id == Account.realm_id)
                .all()
            )

            if not accounts_with_realms or len(accounts_with_realms) <= 0:
                self.app.log.warning("No accounts found.")
                return

            table_data = []
            regex = re.compile(pattern)

            for account, realm in accounts_with_realms:
                if regex.match(account.name):
                    sso_roles = session.query(SsoRole).where(
                        SsoRole.account_id == account.id
                    )

                    table_data.append(
                        [
                            account.number,
                            account.name,
                            realm.name,
                            ", ".join([role.name for role in sso_roles]),
                            account.email,
                        ]
                    )

            self.app.render(
                table_data,
                headers=[
                    "ID",
                    "Name",
                    "Realm",
                    "SSO Roles",
                    "E-mail",
                ],
            )

    @ex(
        help="Download accounts information and store it locally",
        aliases=["sync"],
        arguments=[
            (
                ["--region"],
                {
                    "help": "Which AWS region to authenticate to",
                    "default": "",
                    "dest": "region",
                },
            ),
        ],
    )
    def synchronize(self) -> None:
        database_engine = self.app.database_engine

        with Spinner("Synchronizing accounts database") as spinner:
            realm_name = self.app.pargs.realm or self.app.config.get(
                "aws", "default_realm"
            )

            if not realm_name:
                spinner.error("No AWS realm provided")
                raise RuntimeAppError()

            realm = find_realm(
                database_engine=database_engine,
                realm_name=realm_name,
            )

            if not realm:
                spinner.error(f"Could not find realm {realm_name}")
                raise RuntimeAppError()

            region = self.app.pargs.region or self.app.config.get(
                "aws", "default_region"
            )

            spinner.info(f"Using {realm.name} realm in region {region}")

            authorization = find_authorization(
                database_engine=database_engine,
                realm_name=realm_name,
                region=region,
            )

            if not authorization:
                spinner.error(
                    f"You are not authorized to realm {realm_name} in {region} region"
                )
                raise RuntimeAppError()

            spinner.info("Authorized to AWS")

            with Session(database_engine) as session:
                try:
                    accounts = list_sso_accounts_with_roles(
                        access_token=authorization.client_access_token,
                        region=region,
                    )
                except Exception as e:
                    spinner.error("Could not synchronize accounts", submessage=str(e))
                    raise RuntimeAppError() from e
                else:
                    session.execute(
                        delete(Account).where(Account.authorization == authorization)
                    )
                    session.commit()

                    for account_data in accounts:
                        account = Account(
                            authorization_id=authorization.id,
                            email=account_data["email"],
                            name=account_data["account_name"],
                            number=account_data["account_id"],
                            realm_id=realm.id,
                        )

                        session.add(account)
                        session.commit()

                        account = (
                            session.query(Account)
                            .where(Account.number == account_data["account_id"])
                            .first()
                        )

                        session.execute(
                            delete(SsoRole).where(SsoRole.account_id == account.id)
                        )

                        for sso_role in account_data["sso_roles"]:
                            session.add(
                                SsoRole(
                                    account_id=account.id,
                                    name=sso_role,
                                )
                            )

                    session.commit()
                    spinner.success("All done")
