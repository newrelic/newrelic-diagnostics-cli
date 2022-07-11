#!/bin/sh

if  [[ $(basename `pwd`) != "newrelic-diagnostics-cli" ]]
  then
    echo "Won't run outside a 'newrelic-diagnostics-cli' directory'"
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
if rm ./newrelic-diagnostics-cli 2> /dev/null
  then echo "...removed newrelic-diagnostics-cli binary"
fi

if rm -r ./bin/ 2> /dev/null
  then echo "...removed bin directory"
fi

for file in ./nrdiag*.zip; do
  if rm -r "$file" 2> /dev/null; then
    echo "...removed $file"
  fi
done

for file in ./nrdiag*/; do
  if rm -r "$file" 2> /dev/null; then
    echo "...removed $file"
  fi
done