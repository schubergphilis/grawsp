FROM mundoalem/devcontainer:latest

SHELL [ "/bin/bash", "-o", "pipefail", "-c" ]

USER dev

ENV HOME="/home/dev"
ENV PATH="${HOME}/.local/bin:${PATH}"

COPY .python-version .python-version

RUN pyenv install

ENTRYPOINT [ "/bin/bash" ]
