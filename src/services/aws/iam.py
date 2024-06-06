from typing import Any

import boto3
from botocore.exceptions import ClientError


def find_role_by_name(
    access_key_id: str,
    region: str,
    role_name: str,
    secret_access_key: str,
    session_token: str,
) -> dict[str, Any]:
    session = boto3.Session(
        aws_access_key_id=access_key_id,
        aws_secret_access_key=secret_access_key,
        aws_session_token=session_token,
    )

    iam = session.client("iam", region_name=region)

    try:
        response = iam.get_role(RoleName=role_name)

        return {
            "role_arn": response["Role"]["Arn"],
            "role_id": response["Role"]["RoleId"],
            "role_name": response["Role"]["RoleName"],
            "path": response["Role"]["Path"],
        }
    except ClientError as e:
        if e.response["Error"]["Code"] == "NoSuchEntity":
            return None
    except KeyError as e:
        raise RuntimeError(f"Could not find role '{role_name}', reason: {e}") from e
