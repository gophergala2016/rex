#REx Examples

The project provides examples to demonstrate the REx architecture and provide a
compelling game that begins to show the possibilities of shared applications
utilizing companion apps to provide a "private" user interface in addition to
the information on the shared TV screen.

##Setup (Android)

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

##Running Examples Locally

The `go build` command is capable of compiling native binaries to run examples
during development.

    cd examples/ratscrew
    go get -d ./...
    go build ./cmd/rexserver
    go build ./cmd/rexclient

The produces `rexserver` and `rexclient` binaries can be executed on your local
machine and should work almost identically to how it will work installed on a
device.

##Installing Examples

I, @bmatsuo, have had trouble using `gomobile install`, specifically with my
Android TV (Nexus Player).  So I will just describe the build process using
`gomobile build`, which produces application archives.  Various
[guides](http://www.talkandroid.com/guides/android-tv-guides/how-to-sideload-apps-apk-files-on-the-nexus-player-or-adt-1/)
can be found for sideloading .apk files onto a TV.  Sideloading apps onto a
phone is old news and copious information exists online for that.

First create the application archive (.apk or .app file) using `gomobile`.

    cd examples/ratscrew
    gomobile build -target=android ./cmd/rexserver
    gomobile build -target=android ./cmd/rexclient

Copy the server .apk file to the TV using whatever method works for you (see
link above).  Finally copy the client .apk file to your mobile device.  You
should now be able to start and use both applications.
