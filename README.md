# Let me know

[![baby-gopher](https://raw2.github.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)

`lmk` is a simple command line tool written in Go that draws your attention to
a terminal when another command finishes. Linux only for now but I'd love a PR
that introduces Mac support.


## WTF?!?

Yeah, it may sound silly but how often do you run a command that you know that
takes a lot of time to complete, you go do something else and forget that the
command was running? Even worse, what about when you get side tracked for a bit
longer than you should and when you Alt+Tab to check if the command has finished
it actually errored along the way?

Those situations happen to me more than you might think. Throughout the day I
might run many `bundle install`s, `vagrant up`s, `rake spec`s, etc.. that takes
more than a few seconds to complete. Because looking at a black screen with a
blinking cursor and a whole lot of output is pretty boring, during that period I
usually Alt+Tab to check my emails or twitter and many times I get side tracked
before realizing that I should have been doing something else.


## Ok, so how does it work?

Let's say you want to run the specs for that legacy project you have just been
assigned and the full run takes 5 minutes to complete. With `lmk` you can run
`lmk rake spec` and as soon as `rake spec` finishes running you'll see a `notify-send`
notification poping up on your desktop.

But that's not enough, what if you miss the notification while you are away from
the keyboard? Well, in that case `lmk` will keep letting you know that the
command finished every 30 seconds until you go back to the terminal session that
you left the command running and hit Enter.

```
TODO: Record gif "movie"
```


## Installation

### Standalone

`lmk` can be easily installed as an executable. Download the latest [compiled
binary forms of lmk](https://github.com/fgrehm/lmk/releases) for Linux and Darwin.

### Source

To install `lmk` from source, you need to have a [Go development environment](http://golang.org/doc/install),
version 1.2 or higher, and run `go install github.com/fgrehm/lmk`


## Usage

```sh
Usage: lmk [options...] command

Options:
  -m  Message to display in case of success, defaults to "[command] has completed successfully"
```

### Linux

Since most Linux distributions have `notify-send` pre-installed, everything
should work out of the box. If not, please [let me know](https://github.com/fgrehm/lmk/issues/new).

### OS X

You will need to create a `notify-send` script, available on `$PATH`. The script can forward the message to either
[Growl](http://growl.info/) with [GrowlNotify](http://growl.info/downloads) (version 1.2.2 is free but unreliable)
or to the [Notification Center](http://support.apple.com/kb/HT5362) available on OS X 10.8+
using f.ex. [terminal-notifier](https://github.com/alloy/terminal-notifier).

A (too) primitive script integrating with Growl:

```bash
#!/bin/bash
growlnotify -t "Heads up!" -m "$*"
```


## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
