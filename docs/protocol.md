#REx Protocol

This document describes how the REx clients and servers (objects managed with
the library) communicate with each other.  It is assumed that the client has
already discovered the server using the procedure outlined in the design
[document](design.md).

##Server API

###POST /rex/v0/messages

####Request

Content-Type: application/json

Parameters:

- **session** (string): The session (client application) sending the message.

- **data** (string): The application data being delivered in the message.

- **time** (string): The client's current time (abstract time -- not datetime).
  The specification of this value has not been finalized.

####Response

Status: 200 (or error)

Content-Type: N/A

###GET /rex/v0/events

Parameters:

- **start** (int): The first event index to include in the response.

####Response

Status: 200 (or error)

Content-Type: application/json

Parameters:

- **index** (int): Absolute position of the event in the log.

- **time** (string): The server's current time (abstract time)

- **data** (string): Application data included with the event.

The response is a stream of event objects.  In Go, they should be decoded using
a `json.Decoder` object.
