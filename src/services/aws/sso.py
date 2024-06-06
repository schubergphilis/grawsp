from __future__ import annotations

from datetime import datetime, timedelta
from typing import Any

import boto3

#
# FUNCTIONS
#


def authorize_device(
    client_id: str,
    client_secret: str,
    region: str,
    start_url: str,
) -> dict[str, Any]:
    session = boto3.Session()
    sso_oidc = session.client("sso-oidc", region_name=region)

    response = sso_oidc.start_device_authorization(
        clientId=client_id,
        clientSecret=client_secret,
        startUrl=start_url,
    )

    try:
        return {
            "device_code": response["deviceCode"],
            "device_expires_at": (
                datetime.now() + timedelta(seconds=response["expiresIn"])
            ).timestamp(),
            "verfication_url": response["verificationUriComplete"],
        }
    except KeyError as e:
        raise RuntimeError(f"Could not authorize device, reason: {e}") from e


def create_access_token(
    client_id: str,
    client_secret: str,
    device_code: str,
    region: str,
) -> dict[str, Any]:
    session = boto3.Session()
    sso_oidc = session.client("sso-oidc", region_name=region)

    response = sso_oidc.create_token(
        clientId=client_id,
        clientSecret=client_secret,
        grantType="urn:ietf:params:oauth:grant-type:device_code",
        deviceCode=device_code,
    )

    try:
        return {
            "client_access_token": response["accessToken"],
            "client_access_token_expires_at": (
                datetime.now() + timedelta(seconds=response["expiresIn"])
            ).timestamp(),
        }
    except KeyError as e:
        raise RuntimeError(f"Could not create token, reason: {e}") from e


def list_sso_accounts(
    access_token: str,
    region: str,
) -> list[dict[str, Any]]:
    session = boto3.Session()
    sso = session.client("sso", region_name=region)
    accounts = []
    next_token = None

    while True:
        options = {
            "accessToken": access_token,
        }

        if next_token:
            options["nextToken"] = next_token

        response = sso.list_accounts(**options)

        if "accountList" in response:
            for account in response["accountList"]:
                try:
                    accounts.append(
                        {
                            "email": account["emailAddress"],
                            "account_id": account["accountId"],
                            "account_name": account["accountName"],
                        }
                    )
                except KeyError as e:
                    raise RuntimeError(
                        f"Could not retrieve accounts, reason: {e}"
                    ) from e

        if "nextToken" in response:
            next_token = response["nextToken"]
        else:
            next_token = None
            break

    return accounts


def list_sso_accounts_with_roles(
    access_token: str,
    region: str,
) -> list[dict[str, Any]]:
    accounts = []

    for account in list_sso_accounts(access_token, region):
        sso_roles = list_sso_roles(access_token, account["account_id"], region)
        account["sso_roles"] = sso_roles

        accounts.append(account)

    return accounts


def list_sso_roles(
    access_token: str,
    account_id: str,
    region: str,
) -> list[str]:
    session = boto3.Session()
    sso = session.client("sso", region_name=region)

    next_token = None
    roles = []

    while True:
        options = {
            "accessToken": access_token,
            "accountId": account_id,
        }

        if next_token:
            options["nextToken"] = next_token

        response = sso.list_account_roles(**options)

        if "roleList" not in response:
            break

        for role in response["roleList"]:
            if "roleName" in role:
                roles.append(role["roleName"])

        next_token = (
            response.get("nextToken", None) if "nextToken" in response else None
        )

        if not next_token:
            break

    return roles


def register_client(name: str, region: str) -> dict[str, Any]:
    session = boto3.Session()
    sso_oidc = session.client("sso-oidc", region_name=region)

    response = sso_oidc.register_client(
        clientName=name,
        clientType="public",
    )

    try:
        return {
            "client_id": response["clientId"],
            "client_secret": response["clientSecret"],
            "client_secret_expires_at": response["clientSecretExpiresAt"],
        }
    except KeyError as e:
        raise RuntimeError(f"Could not register client '{name}', reason: {e}") from e


def assume_sso_role(
    access_token: str,
    account_id: str,
    region: str,
    role_name: str,
) -> dict[str, Any]:
    session = boto3.Session()
    sso = session.client("sso", region_name=region)

    response = sso.get_role_credentials(
        roleName=role_name,
        accountId=account_id,
        accessToken=access_token,
    )

    try:
        expiration_in_seconds = response["roleCredentials"]["expiration"] / 1000

        return {
            "access_key_id": response["roleCredentials"]["accessKeyId"],
            "expires_at": datetime.fromtimestamp(expiration_in_seconds).timestamp(),
            "secret_access_key": response["roleCredentials"]["secretAccessKey"],
            "session_token": response["roleCredentials"]["sessionToken"],
        }
    except KeyError as e:
        raise RuntimeError(
            f"Could not assume role '{role_name}' in account '{account_id}', reason: {e}"
        ) from e
