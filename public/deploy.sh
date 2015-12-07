#! /bin/bash

/usr/bin/rsync -avze 'ssh -p 22' --exclude '.git' --exclude '*.sh' -v --delete . thrawn@thrawn01.org:starwars-countdown
