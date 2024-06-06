from __future__ import annotations

import json
import urllib
from typing import Any

import boto3
import requests

from .iam import find_role_by_name


def assume_role(
    access_key_id: str,
    duration: int,
    region: str,
    role_name: str,
    secret_access_key: str,
    session_name: str,
    session_token: str,
) -> dict[str, Any]:
    session = boto3.Session(
        aws_access_key_id=access_key_id,
        aws_secret_access_key=secret_access_key,
        aws_session_token=session_token,
    )

    sts = session.client("sts", region_name=region)
    role = find_role_by_name(
        access_key_id, region, role_name, secret_access_key, session_token
    )

    if not role:
        raise RuntimeError(f"Role '{role_name}' not found")

    response = sts.assume_role(
        RoleArn=role["role_arn"],
        RoleSessionName=session_name,
        DurationSeconds=duration,
    )

    try:
        return {
            "access_key_id": response["Credentials"]["AccessKeyId"],
            "expires_at": (response["Credentials"]["Expiration"]).timestamp(),
            "secret_access_key": response["Credentials"]["SecretAccessKey"],
            "session_token": response["Credentials"]["SessionToken"],
        }
    except KeyError as e:
        raise RuntimeError(f"Could not assume role '{role_name}', reason: {e}") from e


def get_console_url(
    access_key_id: str,
    secret_access_key: str,
    session_token: str,
    region: str = "",
    timeout: int = 10,
) -> str:
    session_data = {
        "sessionId": access_key_id,
        "sessionKey": secret_access_key,
        "sessionToken": session_token,
    }

    federated_signin_endpoint = "https://signin.aws.amazon.com/federation"

    response = requests.get(
        federated_signin_endpoint,
        timeout=timeout,
        params={
            "Action": "getSigninToken",
            "Session": json.dumps(session_data),
        },
    )

    signin_token = json.loads(response.text)
    destination_url = "https://console.aws.amazon.com/"

    if region:
        destination_url = f"{destination_url}?region={region}#"

    query_string = urllib.parse.urlencode(
        {
            "Action": "login",
            "Issuer": "amazon.com",
            "Destination": destination_url,
            "SigninToken": signin_token["SigninToken"],
        }
    )

    federated_url = f"{federated_signin_endpoint}?{query_string}"

    return federated_url
