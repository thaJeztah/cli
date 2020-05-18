%title: docker init hackday demo
%author: thaJeztah
%date: 2019-12-17

With years of new features, the *amount of files and configurations* used by
Docker has grown substantially.
<br>

What started with a "docker run", now involves *Dockerfiles*, *.dockerignore*
files, *docker-compose* files, *apps*, *config.json* configuration files,
contexts, and the list is still growing.
<br>

Often, a single Dockerfile turned out to not be enough for a project, and
*multiple Dockerfiles* started to appear, using conventions such as
*Dockerfile.\<something\>* or *\<something\>.Dockerfile* to describe their
purpose (for example, production, development, ci).
<br>

People dislike cluttering their source repositories with all these files at the
root, so often projects *hide Dockerfiles* in subdirectories, or even *separate*
*repositories*.

---

Where it used to be "just look for a Dockerfile in the repository root, and run
*docker build .*", users now needed to dig into *README's*, look for "docker"
targets in *Makefiles*, and so on...
<br>

Building and running a project with Docker became less discoverable!

---

So, what if we made this easier?
<br>

What if we brought back the ease-of-use of the past, and;
<br>

- define a *standard location* for Docker-related files?
<br>
- make Docker *discoverable* again?
<br>
- make it *easy* to bootstrap your project!
<br>
- enable project-owners to *define* how to *build*, *share*, and *run* with Docker
<br>
- but allow developers to *customize* for their local needs?
<br>
- ... without *cluttering* the source repository root

---

My hack is an cursory exploration of this.
<br>

introducing:
<br>

# dot-docker, and the disciples of init

<br>


...Because Docker is easy, *'init*?

---

> **WARNING**: if you are allergic to duct-tape, hacky code, failing demo's,
> or bad presenting, *now* might be a good moment to look away.

<br>

And a message from my legal department:

> Any resemblance to existing projects, current or cancelled, or actual
> products is purely coincidental.

---

##                               8888888b.  8888888888 888b     d888  .d88888b.  
##                               888  "Y88b 888        8888b   d8888 d88P" "Y88b 
##                               888    888 888        88888b.d88888 888     888 
##                               888    888 8888888    888Y88888P888 888     888 
##                               888    888 888        888 Y888P 888 888     888 
##                               888    888 888        888  Y8P  888 888     888 
##                               888  .d88P 888        888   "   888 Y88b. .d88P 
##                               8888888P"  8888888888 888       888  "Y88888P"  
