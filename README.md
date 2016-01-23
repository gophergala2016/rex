#Room Experiences (REx) [![Build Status](https://travis-ci.org/gophergala2016/rex.svg?branch=master)](https://travis-ci.org/gophergala2016/rex) [![GoDoc](https://godoc.org/github.com/gophergala2016/rex/room?status.svg)](https://godoc.org/github.com/gophergala2016/rex/room)

REx is an experiment attempting to advance the design space of applications and
games running on TV set-top boxes (Android TV in this case).  REx takes
advantage of an underutilized application architucture to provide rich
experience shared by the entire room.

REx pairs an Android TV application with mobile client applications running on
consumer phones (Android and iOS) or custom hardware.  The desired effect is
similar to Nintendo's goals with the Wii U platform.  But the goal for REx is
to provide a low barrier to entry by utilizing existing commodity phones and
set-top boxes.

This project delivers several examples to demonstrate the core technologies as
well an provide an engaging (playable) multiplayer game experience to drive
home the concept.  But, the core deliverable is the open source messaging
framework that can be used to build more advanced games when more resources
(i.e. time and money) are available.

For more information about how a REx application works see the design
[document](docs/design.md).

##Documentation

The `docs/` directory contains various documents that describe the components
of REx and how they are used.  Important sections include the following:

####**[Examples](docs/examples.md)**

Information about the examples provided with REx including a working
implementation of the game Egyptian Ratscrew.

####**[Development](docs/development.md)**

Useful for people trying to hack on REx and compile binaries and application
archives themselvs.

##Motivations

###Why write REx?

Honestly, I want to get together and play board games with my friends without
having to spend all that money on physical pieces (and the effort/space to
store and take care of them).  While many board games exist on mobile devices
many don't offer a low friction multiplayer experience and often they lose some
essense had playing the physical games with friends, sitting at a table.  It is
exactly this feeling I am trying to replicate on a digital platform.  See [Prior Art](#prior-art) for more background.

###Why write an application for a TV?

See the previous answer.  I believe that a large shared place is key to
replicating the board game experience in a fully digital application.  In this
regard a TV functions almost identically to a table.  Furthermore the common
shared space allows almost all **shared information** to be removed from the
mobile interface.  This opens the possibility of implementing games involving
more complex mechanics and player decision points revolving around **secrets**
and **private information** (e.g. poker, Magic: The Gathering, etc).

###Why use mobile "companion" applications?

See the previous answer.  Games and applications relying on private information
and complex mechanics including asymmetric player objectives/behaviors benefit
greatly from a "controller" that has a (i) a display and (ii) programmable
logic.  The [Prior Art](#prior-art) can provide more background on this topic.

###Why write REx in Go?

See the previous answer.  The REx architecture typically involves two
applications being developed and sharing libraries.  What's more, there are
multiple mobile platforms to target (and potentially multiple TV plaforms).
Any sane developer putting themselves in these circumstances would want
cross-platform development libraries and tools.  Furthermore, the truly sane
would ask not to use C(++) for all their development. Luckily these objects are
inline with those of the ["gomobile"](https://github.com/golang/mobile) project
(blessed by Google).  Go's cross-platform toolchain unifies iOS (tvOS?) and
Android application development and allows a single team to maintain the entire
game platform.

What's **more**, when an application grows to also require dynamic cloud-based
applications Go easily extends it reach there and allows you develop the
virtuous path of Go servers connected with Go glue to Go mobile clients.  How
does that bare metal feel?

Hana Kim's [talk](https://www.youtube.com/watch?v=sQ6-HyPxHKg) from GopherCon
2015 provides an execellent background of gomobile, its motivations, and its
benefits.  The video is highly recommended as reference and background on the
history of Go and mobile platforms.

###Why write REx for Gopher Gala 2016?

See the previous answer.  I wanted to create a novel and compelling experience.
And I wanted to build my first real mobile application (and really my first
game with real graphics -- sorry they aren't very good).  After looking over
the gomobile project and how the development cycle goes I thought it would be
possible to create a really special interactive experience in a short amount of
time, even with limited previous mobile and game development experience.

##Prior Art

###Games

As alluded to in the introduction, the space being developed is not unexplored.
Nintendo's [Wii U](http://www.nintendo.com/wiiu) is the most notable example of
a game platform expanding into a complex "controller" architecture.  The Wii U
has an intriguing story for the market.  But the barrier to entry for the
platform is high (higher than would be ideal).  And you can't write a Wii U
game in pure Go!

The concept of local discovery is not new in gaming.  Games have been doing
this going back 20 years (hyperbolic-ish).  More recently in the mobile space
games like [Spaceteam](http://www.sleepingbeastgames.com/spaceteam/) have
delivered very fun and compelling experiences using only local network
discovery.  These games unquestionably demonstrate the ease of use and low
barrier to entry this approach provides over other global discovery engines.
Users do not need to create an account; they launch right into the fun.  And it
is a special kind of fun.

The resurgence of local co-op in videogames like
[Spelunky](http://www.spelunkyworld.com/) and the [Board Game
Renaissance](http://www.theguardian.com/technology/2014/nov/25/board-games-internet-playstation-xbox)
in recent years is evidence that local, in-person gaming utilizing a shared
space can provide some experiences and feelings that are (as of yet)
unatainable through internet based multiplayer experiences which are starkly
impersonal (which remain as starkly impersonal as the rest of the internet).

These are long trends of entertainment and the REx project intends on helping
provide an evolution in the game space with rich in-person multiplayer
experiences built on flexible, cheap, and abundant tools.  REx will take games
like Spaceteam to a new level.

###Collaboration and Productivity

TODO

##Authors

Bryan Matsuo <bryan.matsuo@gmail.com>

##Copyright and License

Copyright 2016 Bryan Matsuo

This project available under the MIT open source license.  See the
[LICENSE](LICENSE) file for more information.
