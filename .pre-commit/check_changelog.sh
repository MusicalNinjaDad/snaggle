#! /bin/bash

sed -n 's/const Version = "\([0-9]\+\.[0-9]\+\.[0-9]\+\)"/"## \\\[v\1\\\] - "/p' version.go | xargs -i grep {} CHANGELOG
