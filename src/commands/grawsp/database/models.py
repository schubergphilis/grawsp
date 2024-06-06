from __future__ import annotations

from datetime import datetime

from sqlalchemy import ForeignKey, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


class Base(DeclarativeBase):
    pass


class Account(Base):
    __tablename__ = "account"

    authorization_id = mapped_column(ForeignKey("authorization.id"))
    authorization: Mapped[Authorization] = relationship(back_populates="accounts")
    email: Mapped[str] = mapped_column(String(320))
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(64))
    number: Mapped[str] = mapped_column(String(12))
    realm_id: Mapped[int] = mapped_column(ForeignKey("realm.id"))
    realm: Mapped[Realm] = relationship(back_populates="accounts")

    credentials: Mapped[list[Credential]] = relationship(
        back_populates="account",
        cascade="all, delete-orphan",
    )

    sso_roles: Mapped[list[SsoRole]] = relationship(
        back_populates="account",
        cascade="all, delete-orphan",
    )

    def __repr__(self) -> str:
        return f"Account(id={self.id!r}, name={self.name!r}, number={self.number!r})"


class Authorization(Base):
    __tablename__ = "authorization"

    client_access_token_expires_at: Mapped[float]
    client_access_token: Mapped[str] = mapped_column(String(256))
    client_id: Mapped[str] = mapped_column(String(32), unique=True)
    client_name: Mapped[str] = mapped_column(String(32))
    client_secret_expires_at: Mapped[float]
    client_secret: Mapped[str] = mapped_column(String(2048), unique=True)
    device_code: Mapped[str] = mapped_column(String(128), unique=True)
    device_expires_at: Mapped[float]
    id: Mapped[int] = mapped_column(primary_key=True)
    realm_id: Mapped[int] = mapped_column(ForeignKey("realm.id"))
    realm: Mapped[Realm] = relationship(back_populates="authorizations")
    region: Mapped[str] = mapped_column(String(32))

    accounts: Mapped[list[Account]] = relationship(
        back_populates="authorization",
        cascade="all, delete-orphan",
    )

    def __repr__(self) -> str:
        return f"Authorization(id={self.id!r}, client_id={self.client_id!r}, client_name={self.client_name!r}, device_code={self.device_code!r}, realm_id={self.realm_id!r})"

    def is_client_access_token_expired(self) -> bool:
        return (
            not self.client_access_token_expires_at
            or datetime.now()
            >= datetime.fromtimestamp(self.client_access_token_expires_at)
        )

    def is_client_secret_expired(self) -> bool:
        return (
            not self.client_secret_expires_at
            or datetime.now() >= datetime.fromtimestamp(self.client_secret_expires_at)
        )

    def is_device_expired(self) -> bool:
        return not self.device_expires_at or datetime.now() >= datetime.fromtimestamp(
            self.device_expires_at
        )


class Credential(Base):
    __tablename__ = "credential"

    access_key_id: Mapped[str] = mapped_column(String(32))
    account: Mapped[Account] = relationship(back_populates="credentials")
    account_id: Mapped[int] = mapped_column(ForeignKey("account.id"))
    expires_at: Mapped[float]
    id: Mapped[int] = mapped_column(primary_key=True)
    role_name: Mapped[str] = mapped_column(String(256))
    secret_access_key: Mapped[str] = mapped_column(String(64))
    session_token: Mapped[str] = mapped_column(String(1024))

    def __repr__(self) -> str:
        return f"Credential(id={self.id!r}, account_id={self.account_id!r}, role_name={self.role_name!r})"

    def is_expired(self) -> bool:
        return not self.expires_at or datetime.now() >= datetime.fromtimestamp(
            self.expires_at
        )


class Realm(Base):
    __tablename__ = "realm"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(256), unique=True, index=True)
    url: Mapped[str] = mapped_column(String(2048))

    accounts: Mapped[list[Account]] = relationship(
        back_populates="realm",
        cascade="all, delete-orphan",
    )

    authorizations: Mapped[list[Authorization]] = relationship(
        back_populates="realm",
        cascade="all, delete-orphan",
    )

    def __repr__(self) -> str:
        return f"Realm(id={self.id!r}, name={self.name!r}, url={self.url!r})"


class SsoRole(Base):
    __tablename__ = "sso_role"

    account: Mapped[Account] = relationship(back_populates="sso_roles")
    account_id: Mapped[int] = mapped_column(ForeignKey("account.id"))
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(256))

    def __repr__(self) -> str:
        return f"SsoRole(id={self.id!r}, name={self.name!r})"
