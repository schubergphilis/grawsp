from __future__ import annotations

import re
import webbrowser
from datetime import datetime, timedelta
from time import sleep

from botocore.exceptions import ClientError
from sqlalchemy import Engine
from sqlalchemy.orm import Session

from ....services.aws.sso import (
    assume_sso_role,
    authorize_device,
    create_access_token,
    register_client,
)
from ....services.aws.sts import assume_role
from ..constants import APP_NAME
from ..database.models import Account, Authorization, Credential, Realm, SsoRole
from ..defaults import (
    DEFAULT_RETRY_AFTER_IN_SECONDS,
    DEFAULT_SESSION_DURATION_IN_SECONDS,
    DEFAULT_TIMEOUT_IN_SECONDS,
)
from ..exceptions import (
    NotFoundAppError,
    RuntimeAppError,
    TimeoutReachedAppError,
)


def create_credential(
    database_engine: Engine,
    account_name: str,
    realm_name: str,
    region: str,
    role_name: str,
    session_name: str = "",
    intermediary_role_name: str = "",
) -> Credential:
    with Session(database_engine, expire_on_commit=False) as session:
        credential = find_credential(
            account_name, database_engine, realm_name, role_name
        )

        if credential:
            if not credential.is_expired():
                return credential
            else:
                session.delete(credential)
                session.commit()

        authorization = find_authorization(database_engine, realm_name, region)

        if not authorization:
            raise NotFoundAppError("Authorization not found")

        account = find_account_by_name(database_engine, realm_name, account_name)

        if not account:
            raise NotFoundAppError(f"Account {account_name} was not found")

        is_sso = (
            session.query(SsoRole)
            .where(SsoRole.name == role_name, SsoRole.account_id == account.id)
            .first()
            is not None
        )

        if is_sso:
            creds = assume_sso_role(
                access_token=authorization.client_access_token,
                account_id=account.number,
                region=region,
                role_name=role_name,
            )
        else:
            if not intermediary_role_name:
                raise RuntimeAppError("An intermediary role was not provided")

            intermediary_creds = create_credential(
                database_engine,
                account_name,
                realm_name,
                region,
                role_name=intermediary_role_name,
            )

            creds = assume_role(
                access_key_id=intermediary_creds.access_key_id,
                duration=DEFAULT_SESSION_DURATION_IN_SECONDS,
                region=region,
                role_name=role_name,
                secret_access_key=intermediary_creds.secret_access_key,
                session_name=session_name,
                session_token=intermediary_creds.session_token,
            )

        credential = Credential(
            access_key_id=creds["access_key_id"],
            account_id=account.id,
            expires_at=creds["expires_at"],
            role_name=role_name,
            secret_access_key=creds["secret_access_key"],
            session_token=creds["session_token"],
        )

        session.add(credential)
        session.commit()

    return credential


def create_authorization(
    database_engine: Engine,
    realm_name: str,
    region: str,
    start_url: str,
    client_name: str = APP_NAME,
    retry_after: int = DEFAULT_RETRY_AFTER_IN_SECONDS,
    timeout: int = DEFAULT_TIMEOUT_IN_SECONDS,
) -> Authorization:
    with Session(database_engine) as session:
        realm = create_realm(
            database_engine=database_engine,
            realm_name=realm_name,
            start_url=start_url,
        )

        authorization = find_authorization(
            database_engine=database_engine,
            realm_name=realm_name,
            region=region,
        )

        if not authorization:
            authorization = Authorization(
                client_name=client_name,
                realm=realm,
                region=region,
            )

        if authorization.is_client_secret_expired():
            client_registration_data = register_client(
                name=client_name,
                region=region,
            )

            authorization.client_secret = client_registration_data["client_secret"]
            authorization.client_id = client_registration_data["client_id"]
            authorization.client_secret_expires_at = client_registration_data[
                "client_secret_expires_at"
            ]

        if (
            authorization.is_device_expired()
            and authorization.is_client_access_token_expired()
        ):
            device_authorization_data = authorize_device(
                client_id=authorization.client_id,
                client_secret=authorization.client_secret,
                region=authorization.region,
                start_url=start_url,
            )

            authorization.device_code = device_authorization_data["device_code"]
            authorization.device_expires_at = device_authorization_data[
                "device_expires_at"
            ]
            verfication_url = device_authorization_data["verfication_url"]

            webbrowser.open_new_tab(verfication_url)

        if authorization.is_client_access_token_expired():
            start_time = datetime.now()

            while True:
                try:
                    access_token_data = create_access_token(
                        client_id=authorization.client_id,
                        client_secret=authorization.client_secret,
                        device_code=authorization.device_code,
                        region=authorization.region,
                    )
                except ClientError as e:
                    if e.response["Error"]["Code"] != "AuthorizationPendingException":
                        raise e

                    elapsed_time = datetime.now() - start_time

                    if elapsed_time >= timedelta(seconds=timeout):
                        raise TimeoutReachedAppError(
                            "Authorization was not approved by the user"
                        ) from e

                    sleep(retry_after)
                else:
                    break

            authorization.client_access_token = access_token_data["client_access_token"]
            authorization.client_access_token_expires_at = access_token_data[
                "client_access_token_expires_at"
            ]

        session.add(authorization)
        session.commit()

        return authorization


def create_realm(
    database_engine: Engine,
    realm_name: str,
    start_url: str,
) -> Realm:
    with Session(database_engine) as session:
        realm = find_realm(database_engine, realm_name)

        if not realm:
            realm = Realm(name=realm_name, url=start_url)
            session.add(realm)
            session.commit()
            realm = session.query(Realm).filter_by(name=realm_name).first()
        else:
            realm.url = start_url
            session.commit()

        return realm


def find_account_by_name(
    database_engine: Engine,
    realm_name: str,
    account_name: str,
) -> Account | None:
    realm = find_realm(database_engine, realm_name)

    if not realm:
        return None

    with Session(database_engine) as session:
        account = (
            session.query(Account)
            .where(Account.name == account_name, Account.realm_id == realm.id)
            .first()
        )
        return account


def find_account_by_number(
    database_engine: Engine,
    realm_name: str,
    account_number: str,
) -> Account | None:
    realm = find_realm(database_engine, realm_name)

    if not realm:
        return None

    with Session(database_engine) as session:
        account = (
            session.query(Account)
            .where(Account.realm == realm, Account.number == account_number)
            .first()
        )
        return account


def find_authorization(
    database_engine: Engine,
    realm_name: str,
    region: str,
) -> Authorization | None:
    with Session(database_engine) as session:
        realm = find_realm(database_engine, realm_name)

        if not realm:
            return None

        authorization = (
            session.query(Authorization)
            .where(Authorization.realm == realm, Authorization.region == region)
            .first()
        )

        return authorization


def find_credential(
    account_name: str,
    database_engine: Engine,
    realm_name: str,
    role_name: str,
) -> Credential | None:
    account = find_account_by_name(database_engine, realm_name, account_name)

    if not account:
        raise NotFoundAppError(f"Account {account_name} was not found")

    with Session(database_engine) as session:
        credential = (
            session.query(Credential)
            .where(Credential.account == account, Credential.role_name == role_name)
            .first()
        )

        return credential


def find_realm(
    database_engine: Engine,
    realm_name: str,
) -> Realm | None:
    with Session(database_engine) as session:
        realm = session.query(Realm).where(Realm.name == realm_name).first()

        return realm


def search_accounts(
    database_engine: Engine,
    realm_name: str,
    pattern: str,
) -> list[Account]:
    realm = find_realm(database_engine, realm_name)

    if not realm:
        return []

    accounts = []

    with Session(database_engine) as session:
        all_accounts = session.query(Account).where(Account.realm == realm).all()
        accounts = [
            account
            for account in all_accounts
            if re.match(pattern, account.number) or re.match(pattern, account.name)
        ]

    return accounts
