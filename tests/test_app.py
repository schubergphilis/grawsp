from src.commands.sbpaws.app import SbpAwsApp


def test_if_we_can_create_app_instance():
    app = SbpAwsApp()
    assert app is not None
