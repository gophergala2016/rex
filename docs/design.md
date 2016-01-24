#REx Design

The REx architecture includes one server application and many clients connected
to the same local network (LAN).  Typically the server will be a TV application
but it may be a mobile or desktop application.  Similarly, a client will often
be a phone but may in fact be any network connected device.  The only
restriction is that all devices are on the same LAN.

The restriction that devices exist on the same network is reasonable for now.
Because the goal is to achieve experiences shared by persons physically in the
same room we can assume they are connected to the same network.

##Communication

A REx server provides a bus by which client applications send (unicast)
messages to the server application.  The server broadcasts an event log which
clients use to update their state.  There is no way in the initial design for
the server to unicast events to an individual client.  All clients receive the
same messages and must filter out irrelevant updates.  This restriction could
be lifted in the future but it simplifies the framework for now.

The following sections describe how clients establish a connection with the
server (discovery) and how messages/events are delivered between the two.  The
bus uses HTTP as a transport protocol (though this is abstracted by the REx
framework).  The endpoints used for communication are described here but the
content of requests and responses is not described.  Refer to the communication
protocol [docs](protocol.md) for more information about HTTP entities involved
in REx communication.

###Discovery

Zeroconf (mDNS) is used for discovery over the LAN.  When the server
application starts a bus is created by REx. After the bus is bound to a random
port and REx begins listening an mDNS service record is configured specifying
the port on which REx clients should connect.

###Client Sessions

When a client is connecting to a server initially it creates a session that it
will use as an identifier in further communication.  A session may (and
typically will) span multiple network connections.

Session IDs are unique strings.  Session IDs are currently generated client
side as a (random) UUID, though this is not part of the protocol specification.
The convention for the server is to associate the first string message sent
using a new session id as the common name for the session (identifier other
clients would see as a "user").

###Message Transport

To send a message to the server the client issues an HTTP request to the server
bus.

    POST /rex/v0/messages HTTP/1.1

The message will be relayed and dispatched to server application logic and may
cause events to be broadcast to all clients (including the message originator).

###Event Transport

All connected clients receive a stream of the server event log.  This stream is
consumed from an HTTP endpoint as a chunked response.

    GET /rex/v0/events?start=0 HTTP/1.1

Whener the client needs to reconnect to the event stream they make another
request like above but update the **start** parameter to be the index of the
first event they would like to receive.

Events and indexed and will persist in the log for the application lifetime so
clients clients may ensure (within reasonable limits) that they will consume
all events in the order they were generated on the server.

While the framework does not support it at this time, the server will
eventually be able to perform both compaction and persistence for the event
log to produce more robust applications.
