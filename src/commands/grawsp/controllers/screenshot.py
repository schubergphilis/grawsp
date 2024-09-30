import base64
import datetime
import re
from time import sleep

from cement import Controller
from inflection import transliterate
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from sqlalchemy.orm import Session

from ....services.aws.sts import get_console_url
from ....util.terminal.spinner import Spinner
from ..actions.aws import (
    create_credential,
    find_account_by_name,
    find_account_by_number,
    search_accounts,
)
from ..constants import APP_NAME
from ..database.models import SsoRole
from ..defaults import DEFAULT_TIMEOUT_IN_SECONDS
from ..exceptions import RuntimeAppError


class ScreenShotController(Controller):
    class Meta:
        label = "screenshot"
        stacked_on = "base"
        stacked_type = "nested"

        arguments = [
            (
                ["--region"],
                {
                    "default": "",
                    "help": "The name of the region you want the console to be opened with.",
                    "dest": "region",
                    "type": str,
                },
            ),
            (
                ["--role"],
                {
                    "default": "",
                    "help": "The name of the role you want to use.",
                    "dest": "role_name",
                    "type": str,
                },
            ),
            (
                ["--urls"],
                {
                    "default": "",
                    "help": "The urls to take screenshots of.",
                    "dest": "urls",
                    "type": str,
                },
            ),
            (
                ["identifier"],
                {
                    "help": "The ID, name or regular expression identifying the account(s).",
                },
            ),
        ]

    def _default(self) -> None:
        database_engine = self.app.database_engine

        chrome_options = Options()
        chrome_options.add_argument("--headless")
        chrome_options.add_argument(
            "--window-size={}".format(self.app.config.get("screenshot", "resolution"))
        )

        identifier = self.app.pargs.identifier
        realm_name = self.app.pargs.realm or self.app.config.get("aws", "default_realm")
        region = self.app.pargs.region or self.app.config.get("aws", "default_region")
        role_name = self.app.pargs.role_name
        urls = self.app.pargs.role_name or self.app.config.get("screenshot", "urls")
        screenshot_urls = urls.split(",")

        user_name = transliterate(
            re.sub(
                r"\s+",
                "",
                self.app.config.get("user", "name"),
                flags=re.UNICODE,
            ),
        )

        with Spinner("Opening AWS console(s)") as spinner:
            accounts = []

            if identifier.isdigit():
                account = find_account_by_number(
                    database_engine, realm_name, account_number=identifier
                )

                if account:
                    accounts.append(account)
            elif re.match(r"^[a-z0-9\-]+$", identifier):
                account = find_account_by_name(
                    database_engine, realm_name, account_name=identifier
                )

                if account:
                    accounts.append(account)
            else:
                accounts = search_accounts(
                    database_engine, realm_name, pattern=identifier
                )

            if len(accounts) <= 0:
                spinner.warning("Identifier matched no accounts")
                return

            for account in accounts:
                if not role_name and self.app.config.has_option(
                    account.name, "default_role"
                ):
                    role_name = self.app.config.get(account.name, "default_role")

                if not role_name and self.app.config.has_option(
                    realm_name, "default_role"
                ):
                    role_name = self.app.config.get(realm_name, "default_role")

                if not role_name:
                    spinner.error("AWS role could not be determined")
                    raise RuntimeAppError()

                spinner.info(f"Using {role_name} role")

                intermediary_role_name = ""

                with Session(database_engine) as session:
                    sso_roles = [
                        role.name
                        for role in session.query(SsoRole)
                        .where(SsoRole.account_id == account.id)
                        .all()
                    ]

                if role_name not in sso_roles:
                    if self.app.config.has_option(account.name, "default_role"):
                        intermediary_role_name = self.app.config.get(
                            account.name, "default_role"
                        )

                    if self.app.config.has_option(realm_name, "default_role"):
                        intermediary_role_name = self.app.config.get(
                            realm_name, "default_role"
                        )

                    if not intermediary_role_name:
                        spinner.error("Intermediary role could not be determined")
                        raise RuntimeAppError()

                    spinner.info(
                        f"Using {intermediary_role_name} as an intermediary role"
                    )

                session_name = f"{APP_NAME}-{user_name}-{role_name}"

                credential = create_credential(
                    database_engine=database_engine,
                    account_name=account.name,
                    realm_name=realm_name,
                    region=region,
                    role_name=role_name,
                    session_name=session_name,
                    intermediary_role_name=intermediary_role_name,
                )

                try:
                    console_url = get_console_url(
                        access_key_id=credential.access_key_id,
                        secret_access_key=credential.secret_access_key,
                        session_token=credential.session_token,
                        region=region,
                        timeout=DEFAULT_TIMEOUT_IN_SECONDS,
                    )
                except Exception as e:
                    spinner.error("Could not generate console URL", submessage=str(e))
                    raise RuntimeAppError() from e

                # encoded_console_url = urllib.parse.quote(console_url)

                cookie_pref = base64.b64encode(
                    b'{"e":1,"p":1,"f":1,"a":1,"i":"71039ed2-6d13-4476-9af3-a4ac0898e71c","v":"1"}'
                ).decode()

                # don't reuse the driver
                driver = webdriver.Chrome(options=chrome_options)
                # wait = WebDriverWait(
                #     driver, self.app.config.get("screenshot", "timeout")
                # )  #

                try:
                    # Navigate to the console URL
                    driver.get(console_url)

                    # Add cookies
                    driver.add_cookie(
                        {
                            "name": "aws_lang",
                            "value": "en",
                            "domain": ".amazon.com",
                            "sameSite": "Strict",
                            "path": "/",
                        }
                    )
                    driver.add_cookie(
                        {
                            "name": "awsccc",
                            "value": cookie_pref,
                            "domain": ".aws.amazon.com",
                            "sameSite": "Lax",
                            "path": "/",
                        }
                    )

                    # Refresh the page to apply cookies
                    driver.refresh()

                    # # Try to find and click the cookie acceptance button
                    # try:
                    #     button = wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR, "[data-id='awsccc-cb-btn-accept']")))
                    #     button.click()
                    # except TimeoutException:
                    #     print("Cookie acceptance button not found or not clickable. Proceeding...")

                    for url_index, url in enumerate(screenshot_urls):
                        # Navigate to the specific page
                        driver.get(url)

                        # Wait for the page to load
                        # wait.until(EC.presence_of_element_located((By.TAG_NAME, "body")))
                        # the wait onlu seems to work the first page, so we sleep instead
                        sleep(self.app.config.get("screenshot", "sleep"))

                        # screenshot name
                        output_date = datetime.datetime.now().strftime(
                            "%Y-%m-%d_%H:%M:%S"
                        )
                        screenshot_filename = (
                            f"{account.name}-{url_index}-{output_date}.png"
                        )

                        # Take a screenshot
                        driver.save_screenshot(screenshot_filename)
                        spinner.info(
                            f"Screenshot to {account.name} {url_index} {url} to {screenshot_filename}"
                        )

                except Exception as e:
                    spinner.error(f"An error occurred: {e}")

                finally:
                    driver.quit()
                    spinner.info("Closed session")

            spinner.success("All done")
