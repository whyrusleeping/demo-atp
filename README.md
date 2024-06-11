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

## Running it

### Build

First, build the AppView server:

```
go build
```

Then build the frontend server.

```
cd client
pnpm i
```

### Running

Run the Go AppView

```
./demo-atp
```

Then run the frontend

```
cd client
pnpm dev
```

### Now put some test data in

Publish some test data with your account by setting your auth info and running the script

```
export DEMOUSER=your.username.social
export DEMOPASSWORD=thebestapppassword
./setup-data.sh
```
