# CI

There seem to be an explosion of CI/Pipeline tools recently.  I've used a few,
and I'm watching [ick](https://ick.liw.fi/) with particular interest.

While a general-purpose CI system is perhaps overkill for my needs I do
have the desire to automate some tasks which run in isolated and transient
containers - and 99% of these tasks are "build something".

For example:

* Build a website, via the [templer](http://github.com/skx/templer) application.
* Build a debian package.

In both cases the result will then be uploaded somewhere - for example a website might be pushed to a live location via `rsync`, and a generated Debian binary-package might be uploaded to a package-repository via `dput`.



## Architecture of a Job

This is where I suspect I'm going to be a little too Steve-specific, and
might miss things.  But in my experience the task of running a CI job
can be broadly divided into three parts:

* Tasks which are executed (on the host) before we begin.
* Tasks that happen in isolation, in a container or transient environment.
* Tasks which are executed (on the host) after we've finished.

To give a concrete example I might want to build a Debian package of
a repository.  To do that, on the host, I run this:

    git clone ssh://git.example.com/repo/here dest/

> **NOTE** The reason I'm running this on the host as it means I don't need to setup an SSH key within the container.

Once I have that local clone (which probably requires SSH setup), I can
build the package and assuming that the previous checkout is available
this can be done in an anonymous container:

    cd dest/
    apt-get install ..
    debuild -i -us -uc -b

Once the build has completed a generated `*.deb` file will be produced,
and from there it can be uploaded to a package repository. (In the case of
a website-build almost everything is the same, except rather than uploading
a single file we'd upload the complete generated output via `rsync`.)

And of course we might want to run these same things in different environments,
such as Debian Jessie, Debian Stretch, Debian unstable.



## CI Job Configuration

So we've established that we probably want to execute some things inside a
transient docker container, and some things outside (typically the "before"
and "after" stages).

This is how we'd implement this:

* Create a temporary directory on the host.
  * Run the "before" steps here.
* Bind-mount that into the container in a known-location "`/work`".
  * Run the "build" steps inside that container.
* Then upload the results.
  * By running the "after" steps on the host.

We can define a configuration-file with three sections "before", "during",
and "after".  Each will be a series of shell-commands as you would expect.

You can see examples of genuine CI files here:

* [kpie.recipe](kpie.recipe)
   * Builds a Debian binary package of the [kpie](https://github.com/skx/kpie) utility.
* [lumail.recipe](lumail.recipe)
   * Builds a Debian binary package of the [lumail](https://github.com/lumail/lumail) console-based email-client.
* [failing.recipe](failing.recipe)
   * Demonstrates that failures terminate the build(s) cleanly.

Of course I might be crazy!  It might be that you'd want to run ALL the steps
inside a container.  In that case just ignore the `before:` and `after:`
sections in your recipe - as this example shows:

* [no-host.recipe](no-host.recipe)
   * Run all the steps in the container.
   * We have an `after:` section solely to show it worked.


## Thyme

This repository contains my simple `thyme`-script.  Given a recipe it
is executed, with some steps running on the host, and some in an
anonymous container.

By default we run against a Debian stretch image, but you can specify that
on your command-line, or even in your recipe (as the [kpie.recipe](kpie.recipe) does ):

    $ ci --recipe ./failing.recipe --container=debian:jessie

vs:

    $ ci --recipe ./failing.recipe --container=debian:stretch



# Feedback?

Feedback is welcome.  Of course the thing we're missing from Jenkins
or Gitlab runners is the notion of dependencies.  Here a job has three
stages and nothing else.

Of course our three stages might not all be present.

Steve
--
