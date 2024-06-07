# ATProto Demo App

This repo implements an application on ATProto. It is an 'appview' for a
linktree style app with a comment section on each users profile.

## Architecture

The AppView portion (the Go application here) subscribes to the firehose and
watches for relevant records, which it then indexes in its database.
It then serves that data over its API to the frontend application where it is
rendered.

The users edit their profile or leave comments by writing records to their
repo, which then get propagated through the firehose and collected here.

For this app, I defined two record types `world.bsky.demo.profile` and
`world.bsky.demo.comment`. These are in the `records` directory, and have cbor
marshalers generated for them by the code in the `gen` directory.
