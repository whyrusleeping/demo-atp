#!/bin/bash
./demo-atp create-profile --handle=$DEMOUSER --password=$DEMOPASSWORD "i am a good poster, i promise" "https://github.com/bluesky-social" "https://reddit.com/r/bluesky"
./demo-atp create-comment --handle=$DEMOUSER --password=$DEMOPASSWORD "$DEMOUSER" "the best comment"
./demo-atp create-comment --handle=$DEMOUSER --password=$DEMOPASSWORD "$DEMOUSER" "sick links bro"
