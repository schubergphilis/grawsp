from cement import Controller
from sqlalchemy import delete
from sqlalchemy.orm import Session

from ....services.aws.sso import list_sso_accounts_with_roles
from ....util.terminal.spinner import Spinner
from ..actions.aws import find_authorization, find_realm
from ..database.models import Account, SsoRole
from ..exceptions import RuntimeAppError


class SyncController(Controller):
    class Meta:
        label = "sync"
        stacked_on = "base"
        stacked_type = "nested"

    def _default(self) -> None:
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

            region = self.app.config.get("aws", "default_region")

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
