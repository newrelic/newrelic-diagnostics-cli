#!/bin/sh

if  [[ $(basename `pwd`) != "nr-diagnostics" ]]
  then
    echo "Won't run outside a 'nr-diagnostics' directory'"
    exit
fi

echo "Looking for output"
if rm ./nrdiag-output.json 2> /dev/null
  then echo "...removed nrdiag-output.json"
fi

if rm ./nrdiag-output.zip 2> /dev/null
  then echo "...removed nrdiag-output.zip"
fi

if rm -r ./nrdiag-filelist.txt 2> /dev/null
  then echo "...removed nrdiag-filelist.txt"
fi

if rm -r ./nrdiag-output/ 2> /dev/null
  then echo "...removed nrdiag-output directory"
fi

echo "Looking for binaries"
if rm ./nr-diagnostics 2> /dev/null
  then echo "...removed nr-diagnostics binary"
fi

if rm -r ./bin/ 2> /dev/null
  then echo "...removed bin directory"
fi
