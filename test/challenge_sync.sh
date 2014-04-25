#!/bin/bash

B=/tmp/wgo-test
LIM=10

if [ $# -eq 1 ] ; then
	LIM=$1
fi

for i in `seq 1 $LIM`; do touch $B/c/$i ; done
for i in `seq 1 $LIM`; do touch $B/u/1 ; done
for i in `seq 1 $LIM`; do touch $B/d/$i && rm $B/d/$i ; done
for i in `seq 1 $LIM`; do touch $B/a/$i && touch $B/a/$i && rm $B/a/$i; done
for i in `seq 1 $LIM`; do touch $B/r/$i && mv $B/r/$i $B/r/$i.mv; done
for i in `seq 1 $LIM`; do touch $B/cudr/$i && touch $B/cudr/$i && rm $B/cudr/$i; done
