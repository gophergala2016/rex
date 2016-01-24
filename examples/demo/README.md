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

TODO
