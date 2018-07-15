# CI

There seem to be an explosion of CI/Pipeline tools recently.  I've used a few,
and I'm watching [ick](https://ick.liw.fi/) with particular interest.

While a general-purpose CI system is perhaps overkill for my needs I do
have the desire to automate some tasks which run in isolated and transient
containers - and 99% of these tasks are "build something".

For example:

* Build a website, via [templer](http://github.com/skx/templer), [hugo](https://gohugo.io/), or similar tool.
* Build a debian package.

In both cases the result will then be uploaded somewhere - for example a website might be pushed to a live location via `rsync`, and a generated Debian binary-package might be uploaded to a package-repository via `dput`.



## Architecture of a Job

In my experience the task of running a CI job can be broadly divided into three parts:

* Tasks which are executed (on the host) before we begin.
  * For example cloning a remote `git` repository.
  * Running this on the host simplifies things because you don't need to setup an SSH key in the container environment.
* Tasks that happen in isolation, in a container or transient environment.
* Tasks which are executed (on the host) after we've finished.
  * For example uploading the generated result(s) to a remote host.
  * Again running this step on the host simplifies things because you don't need to setup credentials in your container for carrying out the upload.


## CI Job Configuration

We've established that we probably want to execute the main build inside a
transient docker container, and some things outside (typically the "before"
and "after" stages).

This is how `thyme` implements this:

* Create a temporary directory on the host.
  * Run the "before" steps against this temporary directory-tree.
  * This is probably where you'd clone your remote `git`, `hg`, etc, repository.
* Bind-mount that into the container in a known-location "`/work`".
  * Run the "build" steps inside that container against the repository.
* Then upload the results.
  * By running the "after" steps on the host.

In short a job is configured by writing a simple recipe which has three
distinct sections within it, each containing shell-commands:

* "before"
  * Runs on the host.
* "during"
  * Runs in the container.
* "after"
  * Runs on the host.

Each of these sections is optional, though of course you'll need to add at least
one.  This repository includes some genuine CI recipes here:

* [kpie.recipe](kpie.recipe)
   * Builds a Debian binary package of the [kpie](https://github.com/skx/kpie) utility.
* [lumail.recipe](lumail.recipe)
   * Build a [lumail](https://github.com/lumail/lumail) binary.
* [failing.recipe](failing.recipe)
   * Demonstrates that failures terminate the build(s) cleanly.

It might be that you'd want to run ALL the steps inside a container.  In that
case just ignore the `before:` and `after:` sections in your recipe - as this
example shows:

* [no-host.recipe](no-host.recipe)
   * Run all the steps in the container.
   * We have an `after:` section solely to show it worked.



## Usage

To test this out:

* Write a recipe, based upon one of the examples.
* Invoke `thyme --recipe=/path/to/recipe.file` to run your job.
* Finally to automate things 100%:
   * Add a git-hook to make this happen every time you run `git push`!

By default `thyme` will run using Debian stretch image, but you can specify
a different one upon the command-line, or even in your recipe (as
the [kpie.recipe](kpie.recipe) does ):

    $ thyme --recipe ./failing.recipe --container=debian:jessie

vs:

    $ thyme --recipe ./failing.recipe --container=debian:stretch




# HTTP Build-Server

This repository contains a simple HTTP-server which can be used to
list and trigger builds:

    $ go run server.go

Once launched you'll see a list of all recipes (<*.recipe>) and clicking
on one will trigger the job - the output will be streamed to your client.

The screen will scroll down to follow new output, until the process has
terminated.



# Feedback?

Feedback is welcome.  Of course the thing we're missing from Jenkins
or Gitlab runners is the notion of dependencies.  Here a job has three
stages and nothing else.

Of course our three stages might not all be present.

Steve
--
