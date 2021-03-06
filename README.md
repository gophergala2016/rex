#Room Experiences (REx) [![Build Status](https://travis-ci.org/gophergala2016/rex.svg?branch=master)](https://travis-ci.org/gophergala2016/rex) [![GoDoc](https://godoc.org/github.com/gophergala2016/rex/room?status.svg)](https://godoc.org/github.com/gophergala2016/rex/room)

[![Demo screenshot](https://raw.githubusercontent.com/gophergala2016/rex/master/screenshots/demo-tv.jpg)](examples/demo)

REx is an experiment attempting to advance the design space of applications and
games running on TV set-top boxes (Android TV in this case).  REx takes
advantage of an underutilized application architecture to provide rich
experience shared by the entire room.

REx pairs an Android TV application with mobile client applications running on
consumer phones (Android and iOS) or custom hardware.  The desired effect is
similar to Nintendo's goals with the Wii U platform.  But the goal for REx is
to provide a low barrier to entry by utilizing existing commodity phones and
set-top boxes.

This project contains working example applications to demonstrate the core
technologies.  The core deliverable is the open source messaging framework that
can be used to build more advanced games when more resources (i.e. time and
money) are available.

For more information about how a REx application works see the design
[document](docs/design.md).

##Features

- Event/Message passing framework for state synchronization.  The framework is
  pure Go so any desktop or mobile app should be able to make use of it for
  low-throughput synchronization and message passing over a local network.

- The framework intended to work across all mobile and TV platforms.  Android
  TV is just the best existing option.  With Apple TV now supporting custom
  apps and as Go support on that device matures look forward to working demos
  built for that platform as well.

- Game servers automatically bind the message bus to a port and make themselves
  discoverable to clients on the local network.  The servers specifically
  target TV set-top boxes (Google Nexus Player, Nvidia Shield, etc). But should
  work on any Android devices (tested on Nexus Player and Galaxy S6).  Untested
  for iOS.

- Game clients automatically discover the server and begin receiving their
  update events.  Works on Android devices (tested on Galaxy S6).  Untested for
  iOS.

- Technical [demo](examples/demo) (with .apk files) showing discovery and state
  synchronization.

##Documentation

The `docs/` directory contains various documents that describe the components
of REx and how they are used.  Important sections include the following:

####**[Examples](docs/examples.md)**

Information about the examples provided with REx.

####**[Development](docs/development.md)**

Useful for people trying to hack on REx and compile binaries and application
archives themselves.

##Motivations

###Why write REx?

Honestly, I want to get together and play board games with my friends without
having to spend all that money on physical pieces (and the effort/space to
store and take care of them).  While many board games exist on mobile devices
few offer a low friction multiplayer experience and often they lose an essence
they had when played physically with friends, sitting at a table.  It is
exactly this feeling I am trying to replicate on a digital platform.  See
[Prior Art](#prior-art) for more background.

###Why write an application for a TV?

See the previous answer.  I believe that a large shared place is key to
replicating the board game experience in a fully digital application.  In this
regard a TV functions almost identically to a table.  Furthermore the common
shared space allows almost all **shared information** to be removed from the
mobile interface.  This can greatly help implementing games involving more
complex mechanics and player decision points, especially those revolving around
**secrets** and **private information** (e.g. Poker, Magic: The Gathering,
etc).

###Why use mobile "companion" applications?

See the previous answer.  Games and applications relying on private information
and complex mechanics including asymmetric player objectives/behaviors benefit
greatly from a "controller" that has a (i) a display and (ii) programmable
logic.  The [Prior Art](#prior-art) can provide more background on this topic.

###Why write REx in Go?

See the previous answer.  The REx architecture typically involves two
applications being developed and sharing libraries.  What's more, there are
multiple mobile platforms to target (and potentially multiple TV platforms).
Any sane developer putting themselves in these circumstances would want
cross-platform development libraries and tools.  The truly sane would go on to
ask not to use C(++) for all their development. Luckily these objectives are
inline with those of the ["gomobile"](https://github.com/golang/mobile) project
(blessed by Google).  Go's cross-platform toolchain unifies iOS (tvOS?) and
Android application development and allows a single team to maintain the entire
game platform.

What's **more**, when an application grows to also require dynamic cloud-based
applications Go easily extends it reach there and allows you develop along the
virtuous path of Go servers connected with Go glue to mobile frontend apps
written in pure Go.  Bare metal feels so good.

Hana Kim's [talk](https://www.youtube.com/watch?v=sQ6-HyPxHKg) from GopherCon
2015 provides an excellent background of gomobile, its motivations, and its
benefits.  The video is highly recommended as reference and background on the
history of Go and mobile platforms.

###Why write REx for Gopher Gala 2016?

See the previous answer.  I wanted to create a novel and compelling experience.
And I wanted to build my first mobile application (and really my first game
using GL graphics).  After looking over the gomobile project and how the
development cycle goes I thought it would be possible to create a cool, shared
interactive experience in a short amount of time, even with limited previous
experience with mobile/game development.

##Prior Art

###Games

As alluded to in the introduction, the space being developed is not entirely
unexplored.  Nintendo's [Wii U](http://www.nintendo.com/wiiu) is the most
notable example of a game platform expanding into a complex "controller"
architecture.  The Wii U has an intriguing story for the market.  But the
economic barrier to entry for the platform is too high for many people.  And
you can't write a Wii U game in pure Go!

The concept of local discovery is not new in gaming.  Games have been doing
this going back 20 years (hyperbolic-ish).  More recently in the mobile space
games like [Spaceteam](http://www.sleepingbeastgames.com/spaceteam/) have
delivered very fun and compelling experiences using only local network
discovery.  These games unquestionably demonstrate the ease of use and low
barrier to entry this approach provides over other global discovery engines.
Users do not need to create an account; they launch right into the fun.  And it
is a special kind of fun.

The resurgence of local co-op in video games like
[Spelunky](http://www.spelunkyworld.com/) and the [Board Game
Renaissance](http://www.theguardian.com/technology/2014/nov/25/board-games-internet-playstation-xbox)
in recent years is evidence that local, in-person gaming utilizing a shared
space can provide some experiences and feelings that are (as of yet)
unattainable through internet based multiplayer experiences which are starkly
impersonal (as starkly impersonal as the rest of the internet).

These are long trends of entertainment and the REx project intends on helping
provide an evolution in the game space with rich in-person multiplayer
experiences built on flexible, cheap, and abundant tools.  REx will take games
like Spaceteam to a new level.

###Collaboration and Productivity

The most widely known collaboration product to make use of local-only discovery
is probably Google's Chromecast (Google Cast technology in general).  Using
local network discovery devices can send streaming requests to Chromecast
devices on the same network.  REx is heavily inspired by this architecture and
could be used to emulate the Chromecast's functionality.

Apple is another notable user of local discovery, authoring their own protocol
called Bonjour. Bonjour is the discovery protocol powering several Apple
services including Air Drop (a file transfer service).  The REx authors take
inspiration from Apple's work developing a clean user experience and workflow
for collaborating with colocated peers.

Both examples here are centralized platforms and have limitations either in
their application or the client operating system.  The REx author believes that
even greater productivity and collaboration tools will appear when application
developers are given an open, cross-platform framework and the tools for
harnessing local discovery-based mechanics in their own systems.

##Authors

Bryan Matsuo <bryan.matsuo@gmail.com>

##Copyright and License

Copyright 2016 Bryan Matsuo

This project available under the MIT open source license.  See the
[LICENSE](LICENSE) file for more information.
