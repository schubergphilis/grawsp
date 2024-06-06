from datetime import datetime

from cement import Controller, ex
from humanize import naturaltime
from sqlalchemy.orm import Session

from ....util.terminal.spinner import Spinner
from ..actions.aws import create_authorization
from ..constants import APP_NAME
from ..database.models import Authorization, Realm
from ..exceptions import RuntimeAppError


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
            ),
        ]

    @ex(
        aliases=["auth"],
        help="Authorize your local client to AWS using SSO",
        arguments=[
            (
                ["--region"],
                {
                    "help": "Which AWS region to authenticate to",
                    "default": "",
                    "dest": "region",
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
                ["--timeout"],
                {
                    "help": "How long to wait until we abort the authorization process",
                    "default": "",
                    "dest": "timeout",
                },
            ),
        ],
    )
    def authorize(self) -> None:
        with Spinner("Authorizing to AWS") as spinner:
            realm_name = self.app.pargs.realm or self.app.config.get(
                "aws", "default_realm"
            )

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
            database_engine = self.app.database_engine
            region = self.app.pargs.region or self.app.config.get(
                "aws", "default_region"
            )
            retry_after = int(
                self.app.pargs.retry_after
                or self.app.config.get("general", "retry_after")
            )
            timeout = int(
                self.app.pargs.timeout or self.app.config.get("general", "timeout")
            )

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
    def status(self) -> None:
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
                    "Region",
                    "Client ID",
                    "Expires In",
                ],
            )
