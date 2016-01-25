#REx Demo

This example includes a server and client that demonstrate the communication
between the server and (multiple) clients.

For information about installing and running the example see the
[documentation](../../docs/examples.md).

##About

The demo is based on the example provided in
[golang.org/x/mobile/examples/basic](https://github.com/golang/mobile/tree/master/example/basic).
Touch events move the sprite (green triangle) around the screen.  In this
example though the touch events are passed to the server and all connected
clients will see the triangle dance on the screen as you tap around.

##Usage

First launch the server on your desktop or mobile device.  After the server is
running (it has a red background) launch the client application (it has a blue
background).

"Touch" the screen on the client to update the position of the green triangle
(you might not see a triangle at first).  You should see the server (red)
update it's triangle's position to the touch position.

##Screenshots

The following screenshots and photos show the server (red) and several
connected clients (light blue).  The state (last message time, number of
messages, and triangle position) are synchronized between all processes.

This first image shows the applications running on the same desktop.  The
location of the triangle on each screen is synchronized when touch events
occur.  The stats at the top of each screen are there to ensure the states of
each device is up to date.

![Demo on desktop](https://raw.githubusercontent.com/gophergala2016/rex/master/screenshots/demo.png)

This image shows the server running on an Android TV and a computer running two
connected clients.

![Demo on TV](https://raw.githubusercontent.com/gophergala2016/rex/master/screenshots/demo-tv.jpg)
