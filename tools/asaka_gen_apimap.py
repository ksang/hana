#!/usr/bin/python
import sys
import re

if __name__ == "__main__":
	if len(sys.argv) != 2:
		sys.exit(0)
	fo = open(sys.argv[1], "r")
	lines = fo.readlines()
	print "apinamemap:"
	pattern = re.compile(r'\((.*?)\)')
	for idx, line in enumerate(lines):
		for match in re.findall(pattern, line):
			print "  \"" + str(idx+1) + "\": " + match
			break
