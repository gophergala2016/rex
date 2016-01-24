#REx Examples

The project provides examples to demonstrate the REx architecture and provide a
compelling game that begins to show the possibilities of shared applications
utilizing companion apps to provide a "private" user interface in addition to
the information on the shared TV screen.

##Pre-compiled packages

Precompiled android packages are provided for each
[release](https://github.com/gophergala2016/rex/releases).  Users can download
one of the "android" tarballs and load the included .apk files onto their
Android devices with installing the `gomobile` and Android development
toolchains.

##Installation (via gomobile)

###Setup (Android)

To target a platform follow all the typical setup for standard development
using the native SDKs.

**TODO: Links**

Before examples can be installed the `gomobile` environment must be initialized
for the target platform ("ios" or "android").  The golang/mobile project is
experimental so please take any advice here with a grain of salt.  Some
instructions for android are provided but more up to date ones should be
available from the [project](https://github.com/golang/mobile).

If this has not already been done, run the following commands:

    go get golang.org/x/mobile/cmd/gomobile
    gomobile init

You may need to install the `adb` program as well, which is an OS-dependent
task.  OS X users can use Homebrew.

    brew install android-platform-tools

###Running Examples Locally

The `go build` command is capable of compiling native binaries to run examples
during development.

    cd examples/demo/cmd/rexserver
    go build ./cmd/rexserver
    ./rexserver

The same process is used for building and running client apps.

###Installing Examples

I, @bmatsuo, have had trouble using `gomobile install`, specifically with my
Android TV (Nexus Player).  So I will just describe the build process using
`gomobile build`, which produces application archives.  Various
[guides](http://www.talkandroid.com/guides/android-tv-guides/how-to-sideload-apps-apk-files-on-the-nexus-player-or-adt-1/)
can be found for sideloading .apk files onto a TV.  Sideloading apps onto a
phone is old news and copious information exists online for that.

First create the application archive (.apk or .app file) using `gomobile`.

    cd examples/demo/cmd/rexserver
    gomobile build -target=android ./cmd/rexserver

Copy the produced `rexserver.apk` file to an android device then install and
open it.

##Index

###[Demo](../examples/demo)

The demo provides a full example of the REx room framework's capabilities.
Clients auto-discover and connect to server applications and synchronize their
state when touches cause the client to update.

![Demo screengrab](https://raw.githubusercontent.com/gophergala2016/rex/master/screenshots/demo.png)

###[Ratscrew](../examples/ratscrew)

Unfortunately the full game example of Egyptian Ratscrew had to be postponed
beyond the Gopher Gala 2016 as technical issues deploying autodiscovery to
Android slowed overall progress.
