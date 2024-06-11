[![Release Status](https://github.com/schubergphilis/grawsp/actions/workflows/pipeline.yml/badge.svg)](https://github.com/schubergphilis/grawsp/actions/workflows/pipeline.yml)

# grawsp

A command line application to assist engineers manage credentials in an AWS landing
zone.

- SSO-OIDC client
- Assume SSO enabled roles
- Use intermediary roles to assume others, when the role is not SSO enabled
- Manage credentials on multiple landing zones (realms)
- Export access credentials to your local AWS cli configuration file
- View which credentials are valid or expired
- Open AWS consoles from the command line (*)
- Get credentials for multiple accounts as a specific role
- Locally cache credentials

(*) Currently only Firefox is supported

## Requirements

- Linux or macOS (*)
- Python 3.10+

(*) Windows support only through WSL

## Installing

You can install it like any other Python package hosted in PyPi:

```bash
pip install grawsp
```

... or using `pipx`:

```bash
pipx install grawsp
```

Make sure you have the `~/.local/bin` directory in your `$PATH` and that should be
enough for you to be able to use `grawsp`.

## Getting Started

### Configuration

The path to the configuration file is `~/.config/grawsp/grawsp.conf` and here is what
the contents of the file could be:

```text
[user]
email = your-email@your-company.com
name = Your Name

[aws]
default_realm = my-landingzone-1
default_region = eu-central-1

[my-landingzone-1]
default_role = MyReadOnlyRole
start_url = https://d-1111111111.awsapps.com/start/

[my-landingzone-2]
default_role = MyAdminRole
start_url = https://d-2222222222.awsapps.com/start/

[general]
firefox_path = /Applications/Firefox.app/Contents/MacOS/firefox
```

### Quickstart

First you need to register your device and authenticate yourself:

```bash
grawsp auth # will open your default browser to follow the SSO-OIDC process
```

Then you need to synchronise the list of AWS accounts available to you:

```bash
grawsp sync
grawsp list accounts
```

Now you can also get credentials for a role in an account:

```bash
grawsp auth 012345678910
grawsp auth my-account-dev
grawsp auth "my.*-dev"
grawsp auth --role ReadOnly "my.*-dev"
grawsp auth --role Admin --from-role Operator "my.*-dev"
grawsp list creds
```

If you need to open the web console(*):

```bash
grawsp open-console "my.*-dev"
grawsp open-console --role AdminRole --region ap-south-2 my-account-dev
```

If you want to export your credentials to use in the [AWS Command Line Interface](https://aws.amazon.com/cli/):

```bash
grawsp export --default-account my-account-dev --default-role ReadOnly
```

(*) This will use Firefox and not your default browser

### We need to talk about Firefox

Firefox is the only browser which allows us to isolate multiple tabs for the same
website. If you also install [this extension](https://addons.mozilla.org/en-US/firefox/addon/open-url-in-container/),
then `grawsp` will be able to open the AWS web console for multiple accounts in the same
browser window.

Unfortunately we could not replicate the same feature in other browsers. We are still
researching what would be the best experience for our users.

## Contributing

This projects makes use of the [devcontainer](https://containers.dev/) standard, so
if you want to contribute just open the project in a editor or IDE which supports
development containers, like [Visual Studio Code](https://code.visualstudio.com/docs/devcontainers/containers)
and your environment will be properly setup.

If you don't want to use an external development container, you will need the following
dependencies to be installed and configured, refer to each dependency documentation to
understand how to install and configure them.

- Python 3.10+
- Poetry
- make
- direnv

Feel free then to fork the project and create a pull request to it once the change is
completed. The project will run the pipeline automatically on all pull requests.

The project uses `make` and the tool to drive all project related tasks:

| Job     | Description                                               |
| ------- | --------------------------------------------------------- |
| all     | Runs lint, scan, build and test jobs                      |
| build   | Build a package and store it in `dist/` dir               |
| clean   | Clean build and temporary files                           |
| env     | Reloads `.envrc`                                          |
| lint    | Runs `ruff` against the source code                       |
| release | Publish the package to PyPi                               |
| scan    | Uses `bandit` to scan the code for common security issues |
| test    | Run the application tests                                 |

## License

```text
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this
file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the specific language governing
permissions and limitations under the License.
```
