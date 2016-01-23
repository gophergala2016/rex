#Room Experiences (REx) [![Build Status](https://travis-ci.org/gophergala2016/rex.svg?branch=master)](https://travis-ci.org/gophergala2016/rex)

REx is an experiment attempting to advance the design space of applications and
games running on TV set-top box (Android TV in this case).  REx takes advantage
of an underutilized application architucture to provide rich experience shared
by the entire room.

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
Users do not need to create an account, the launch right into the fun.  And it
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
