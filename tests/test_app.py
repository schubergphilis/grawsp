from src.commands.grawsp.app import GrawspApp


def test_if_we_can_create_app_instance():
    app = GrawspApp()
    assert app is not None
