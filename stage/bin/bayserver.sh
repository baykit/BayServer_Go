#!/bin/sh
bindir=`dirname $0`
args=$*
daemon=
for arg in $args; do
  if [ "$arg" = "-daemon" ]; then
    daemon=1
  fi
done

if [ "$daemon" = 1 ]; then
   $bindir/bayserver $* < /dev/null  > /dev/null 2>&1 &
else
   $bindir/bayserver $*
fi