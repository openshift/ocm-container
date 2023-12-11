#!/usr/bin/env bash

set -o pipefail

usage() {
  cat >&2 <<EOF
  usage: $0 [-i path/to/file | stringWith@@PLATFORM@@toReplace] --ARMARCH --X86ARCH [ OPTIONS ]

  requires either -i and a path to a file to replace, or a positional string to replace
  requires one of each of the x86 arch options below and arm arch options below.

  ARMARCH Options:
  --arm64                 target arm64 binary for arm arch
  --aarch64               target aarch64 binary for arm arch
  --custom-arm64 string   target arm64 string
  X86ARCH Options:
  --amd64                 target amd64 binary for x86 arch
  --x86_64                target x86_64 binary for x86 arch
  --custom-amd64 string   target x86_64 string
  Additonal Options:
  -v | --verbose    Enables more verbose output
  -i | --file       Replaces all instances of "@@PLATFORM@@" in a file, rewriting the file
EOF
}

# Remove this if check if there is not a need for parameters to be passed in.
if [ $# -lt 3 ]; then
  usage
  exit 1
fi

REPLACESTR=
REPLACEFILE=
VERBOSE=false

while [ "$1" != "" ]; do
  case $1 in
    --aarch64 )         ARMARCH=aarch64
                        ;;
    --arm64 )           ARMARCH=arm64
                        ;;
    --amd64 )           X86ARCH=amd64
                        ;;
    --x86_64 )          X86ARCH=x86_64
                        ;;
    --custom-arm64)     shift
                        ARMARCH=$1
                        ;;
    --custom-amd64)     shift
                        X86ARCH=$1
                        ;;
    -i | --file )       shift
                        REPLACEFILE=$1
                        ;;
    -v | --verbose )    VERBOSE=true
                        ;;
    --* )               echo "Unexpected parameter $1" >&2
                        usage
                        exit 1
                        ;;
    * )                 if [[ -z $REPLACESTR ]];
                        then
                            REPLACESTR=$1
                        else
                            echo "Too many positional arguments." >&2
                            usage
                            exit 1
                        fi
                        ;;
  esac
  shift
done

if [[ -z $REPLACESTR ]] && [[ -z $REPLACEFILE ]]
then
    echo "Missing required string or file" >&2
    usage
    exit 1
fi

if [[ -n $REPLACESTR ]] && [[ $(grep "@@PLATFORM@@" <<< $REPLACESTR | wc -l) -eq 0 ]]
then
    echo "Replacement string needs '@@PLATFORM@@' to replace." >&2
    usage
    exit 1
fi

if [[ -z $ARMARCH ]]
then
    echo "Missing ARMARCH parameter." >&2
    usage
    exit 1
fi

if [[ -z $X86ARCH ]]
then
    echo "Missing X86ARCH parameter." >&2
    usage
    exit 1
fi

arch=$(uname -m)

replace_file() {
    sed -Ei "s/@@PLATFORM@@/$2/" $1
}

replace_str() {
    sed -e "s/@@PLATFORM@@/$2/" <<< $1
}

if [[ $arch == "arm64" ]] || [[ $arch == "aarch64" ]];
then
    if [[ -n $REPLACEFILE ]]; then replace_file $REPLACEFILE $ARMARCH; fi
    if [[ -n $REPLACESTR ]]; then replace_str $REPLACESTR $ARMARCH; fi
    exit 0
fi

if [[ $arch == "amd64" ]] || [[ $arch == "x86_64" ]];
then
    if [[ -n $REPLACEFILE ]]; then replace_file $REPLACEFILE $X86ARCH; fi
    if [[ -n $REPLACESTR ]]; then replace_str $REPLACESTR $X86ARCH; fi
    exit 0
fi

echo "Unexpected architecture from 'uname -m': $arch" >&2
exit 2
