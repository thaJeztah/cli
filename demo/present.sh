#!/usr/bin/env bash

clear

read -r -p "



With years of new features, the amount of files and configurations used for docker has grown substantially.
"

read -r -p "What started with a 'docker run', now involves Dockerfiles, .dockerignore files, compose files, apps,
config-files, and the list is still growing.
"

read -r -p  'Often, a single Dockerfile is not enough for a project, so multiple Dockerfiles started to appear,
using conventions such as "Dockerfile.<something>" or "<something>.Dockerfile".
'

read -r -p "People dislike cluttering their source repositories with all these files at the root, so often
projects hide Dockerfiles in subdirectories, or even separate repositories.
"

clear

read -r -p "




Where it used to be \"just look for a Dockerfile in the repository root, and run 'docker build .'\",
it's now needed to dig into README's, look for 'docker' targets in the Makefile, and so on...
"

read -r -p "Building and running a project with Docker became less discoverable!
"

read -r -p "So, what if we made this easier?
"

read -r -p "What if we brought back the ease-of-use of the past, and;
"

read -r -p "- define a standard location for Docker-related files?"
read -r -p "- make Docker discoverable again?"
read -r -p "- make it easy to bootstrap your project!"
read -r -p "- allow project-owners to define how to build, share, and run with Docker"
read -r -p "- give developers the ability to 'customize' for their local needs?"

clear

read -r -p "



My hack is an cursory exploration of this:
"

read -r -p 'introducing "dot-docker, and the disciples of init"
'

read -r -p "Because Docker is easy, 'init?
"

clear

read -r -p "



WARNING: if you're allergic to duct-tape, hacky code and failing demo's, or
         bad presenting, now might be a good moment to look away."


read -r -p "

And a message from my legal department:

    Any resemblance to existing projects, current or cancelled, or actual
    products is purely coincidental.
"

read -r

clear
