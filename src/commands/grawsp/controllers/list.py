import re
from datetime import datetime

from cement import Controller, ex
from humanize import naturaltime
from sqlalchemy.orm import Session

from ..database.models import Account, Authorization, Credential, Realm, SsoRole


class ListController(Controller):
    class Meta:
        label = "list"
        stacked_on = "base"
        stacked_type = "nested"

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
    def accounts(self) -> None:
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
        help="Display information about the AWS authorization",
        arguments=[
            (
                ["--expired"],
                {
                    "action": "store_true",
                    "default": False,
                    "dest": "show_expired",
                    "help": "Include expired authorizations in the output",
                },
            ),
        ],
    )
    def authorization(self) -> None:
        database_engine = self.app.database_engine
        show_expired = self.app.pargs.show_expired
        table_data = []

        with Session(database_engine) as session:
            authorizations_with_realms = (
                session.query(Authorization, Realm)
                .join(Realm, Authorization.realm_id == Realm.id)
                .all()
            )

            if not authorizations_with_realms or len(authorizations_with_realms) <= 0:
                self.app.log.warning("No authorizations found.")
                return

            for authorization, realm in authorizations_with_realms:
                if authorization.is_client_access_token_expired() and not show_expired:
                    continue

                table_data.append(
                    [
                        realm.name,
                        authorization.region,
                        authorization.client_id,
                        naturaltime(
                            datetime.fromtimestamp(
                                authorization.client_access_token_expires_at
                            )
                        ),
                    ]
                )

            self.app.render(
                table_data,
                headers=[
                    "Realm",
                    "Client ID",
                    "Expires In",
                ],
            )

    @ex(
        help="List all the valid credentials available.",
        arguments=[
            (
                ["--expired"],
                {
                    "action": "store_true",
                    "default": False,
                    "dest": "expired",
                    "help": "Include expired credentials in the output",
                },
            ),
        ],
    )
    def creds(self) -> None:
        database_engine = self.app.database_engine
        show_expired = self.app.pargs.expired
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
                        naturaltime(
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
