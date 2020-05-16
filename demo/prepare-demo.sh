#!/usr/bin/env sh

# install demo project templates
mkdir -p ~/.docker
if [ -f ~/.docker/project-templates ]; then
	ln -s ../project-templates ~/.docker/project-templates
fi

build_cli() (
	set -ex
	DISABLE_WARN_OUTSIDE_CONTAINER=1 make -C ../ binary
	cp -L ../build/docker ./docker
)

build_cli
