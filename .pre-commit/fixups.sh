#!/bin/bash

fixups=$(git rev-list --regexp-ignore-case --grep='^fixup!' --oneline '@{u}..')
if [[ -n $fixups ]]; then
    echo "Please rebase the following commits before pushing:"
    echo "$fixups"
    exit 1
fi
