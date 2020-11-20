#!/bin/bash -e
remove_coloring() {
	$@ 2>&1| sed 's/[[:cntrl:]]\[[0-9]{1,3}m//g'
}
