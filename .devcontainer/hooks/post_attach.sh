#!/usr/bin/env bash

#
# DIRENV
#

direnv allow /workspaces/*

#
# DOCKER
#

sudo chown root:docker /var/run/docker.sock
sudo chmod g+w /var/run/docker.sock

#
# GIT
#

ls -d /workspaces/* | xargs git config --global --add safe.directory

#
# PYTHON
#

poetry config virtualenvs.in-project true
poetry env use "$(cat .python-version)"
poetry install
poetry run pre-commit install

#
# STARSHIP
#

starship preset plain-text-symbols -o ~/.config/starship.toml
starship config container.disabled true
