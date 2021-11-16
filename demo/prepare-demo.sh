#!/usr/bin/env sh

set -ex

# install demo project templates
mkdir -p ~/.docker
if [ -f ~/.docker/project-templates ]; then
	ln -s ../project-templates ~/.docker/project-templates
fi

build_cli() (
	DISABLE_WARN_OUTSIDE_CONTAINER=1 BUILDTIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")" make -C ../ binary
	cp -L ../build/docker ./docker
)

# docker run -it --rm -v (pwd):/present mdp bash -c 'sleep 1; mdp --nofade --invert --notrans slides.md'
build_mdp() (
	# FIXME: can't use BuildKit here until https://github.com/moby/moby/issues/38254 is fixed
	DOCKER_BUILDKIT=0 docker build -t mdp -f- https://github.com/visit1985/mdp.git <<-'EOF'
	FROM alpine AS build
	RUN apk add --no-cache build-base ncurses-dev
	RUN mkdir -p /src/mdp/
	WORKDIR /src/mdp/
	COPY . .
	RUN make

	FROM alpine AS final
	RUN apk add --no-cache bash
	CMD /bin/bash
	COPY --from=build /usr/lib/libncursesw.so.6 /usr/lib/
	COPY --from=build /lib/ld-musl-x86_64.so.1 /lib/
	COPY --from=build /src/mdp/mdp /usr/local/bin/
	WORKDIR /present
	ENV TERM=xterm-256color
	# Need a short sleep to wait for docker to trigger a resize to set term width/height
	CMD sleep 1.0 && mdp --nofade --invert --notrans *.md
	EOF
)

build_presentation() (
	docker build -t presentation -f- . <<-'EOF'
	FROM mdp
	COPY slides.md .
	CMD ["bash", "-c", "sleep 1; mdp --nofade --invert --notrans slides.md"]
	EOF
)

build_cli
build_mdp
build_presentation

