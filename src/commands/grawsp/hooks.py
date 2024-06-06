from pathlib import Path

from cement import App
from sqlalchemy import create_engine

from .database.models import Base


def database_hook(app: App) -> None:
    path = Path(app.config.get("database", "path"))

    path.parent.mkdir(parents=True, exist_ok=True)
    path.touch(exist_ok=True)

    uri = f"sqlite:///{path.as_posix()}"
    engine = create_engine(uri)

    Base.metadata.create_all(engine)
    app.extend("database_engine", engine)
