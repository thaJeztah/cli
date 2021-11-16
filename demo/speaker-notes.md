# dot-docker, and the disciples of init

Docker Hackweek hack, december 2019

To prepare the demo:

```bash
cd demo
./prepare-demo.sh
```

This builds the `presentation` image (including the slides), and the `docker` cli
from the code of this branch.

To present the slides:

```bash
docker run -it --rm presentation
```

(exit presentation with `q`)

Update your path to use the local `docker` cli without having to type `./docker`,
and cleanup the `./.docker` directory in the demo (if present)

If you're using Bash/Zsh

```bash
export PATH=$(pwd):$PATH
rm -rf ./.docker
```

This hack adds a `docker init` sub-command to the cli

```bash
docker init --help
Usage:	docker init [OPTIONS]

Initialize a docker project

Options:
  -t, --template string   Template to use for initializing docker
```

The `docker init` sub-command initializes a Docker project by creating a local
`.docker` directory from a template. Templates can be stored in the
`~/.docker/project-templates` directory (additional "system-wide" paths could be
added, or templates could be distributed through a registry).

```bash
docker init
Use the arrow keys to navigate: ↓ ↑ → ←
? Select Docker template:
  ▸ basic
    default
    hackit
```

For example, let's pick a "basic" example

```console
tree -a

.
├── .docker
│   ├── .gitignore
│   ├── config.json
│   └── local
│       └── .keep
```


The basic example has:

- a config.json to define the project's default cli configuration
- a default .gitignore (more about this later)
- a local directory


Where the config.json at the root of the .docker directory can define settings
that the project author defines, a developer may have their own preferences.

What if, for this project I want experimental CLI features to be enabled?

```bash
docker version | grep Experimental
```

The `local` directory allows local overrides, for example, to override some CLI
settings. By default, the `local` directory is excluded from git, so I don't have
to worry about my personal preferences being committed to source control.

```bash
echo '{"experimental":"enabled"}' > ./.docker/local/config.json
```

This is just to illustrate / explore options that we could consider, and if we extend
the cli configuration with additional options (default namespace/filtering perhaps? only show
images, stacks, containers for a specific project)

So, what could be in a `.docker` directory? Let's pick a slightly extended example:

```bash
rm -rf ./.docker

docker init
```

we pick the "hackit" template

```console
tree -a
.
├── .docker
│   ├── .gitignore
│   ├── buildstack
│   │   ├── Dockerfile.db
│   │   ├── Dockerfile.web
│   │   └── docker-compose.yml
│   ├── config.json
│   ├── dev
│   │   ├── Dockerfile
│   │   └── description.txt
│   ├── local
│   │   └── config.json
│   ├── prod
│   │   ├── Dockerfile
│   │   └── description.txt
│   └── webapp
│       └── docker-compose.yml
```

Again, this is just exploring options: we could make the CLI more interactive:

- discovering Dockerfiles, compose-files is complicated
- building and running containers often requires various parameters, and typing
  all of those on the command-line is not maintainable. Compose files and `docker app`
  solves some of those, but again, discovering may be a thing
- perhaps we can make this usable for other things (build, buildx, stacks)?


Let's try a build! (notice there's no `.` after build!)

```console
docker build
Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
Select config?
  ▸ buildstack (db) (Compose File Service)
    buildstack (web) (Compose File Service)
    dev (Dockerfile)
↓   prod (Dockerfile)
```

Here, we can select what to build:

- different variations or parts of the project
- with a whole directory-structure to play with, we can add more metadata / files
  where needed.
- For example, a description for each of the options

Build the "dev" or "prod" variant

Or perhaps, build a specific service that's defined in a compose file?

Deploying a stack (or "compose project", or "app") could become easier:

```console
docker stack deploy
Use the arrow keys to navigate: ↓ ↑ → ←
Select stack?
  ▸ buildstack
    webapp
```

After selecting the stack to deploy, deploy goes as usual:

```console
docker stack deploy
✔ buildstack
Using stack "buildstack"
Ignoring unsupported options: build

Creating network buildstack_default
Creating service buildstack_web
Creating service buildstack_db
```

This version of the CLI also allows `docker run` without passing an image. If
no image is passed, the image is selected interactively. Currently the interactive
selection is not "project" aware, but this could be added. When adding that
feature, selection could both allow running a single service from the stack
(interactively), or find all images related to the project:

```console
docker run -it --rm
Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
? Select image:
    docker-dev:split-resource-types
    docs/docstage:latest
  ▸ presentation:latest
↓   mdp:latest
```

Search is also implemented, which can be useful if many images are present in
the local image cache (and could even be extended to searching Docker Hub);

```console
docker run -it --rm
Search: ubuntu█
? Select image:
  ▸ ubuntu:20.04
    ubuntu:bionic
    ubuntu:focal
```

I only worked on this for a day, so that's all there is for now; hope this can
inspire some ideas though!
