# flagfile [![GoDoc](https://godoc.org/layeh.com/flagfile?status.svg)](https://godoc.org/layeh.com/flagfile)

Package flagfile converts a flag file into flag-compatible arguments.

A flag file is alternative way to provide command-line flags for a program.
For example, the below arguments:

	./example --enable-video --user=tim --user=dave --size 3 --message="hello\tworld"

Would be stored in the flag file as the following:

	# Enable video output
	enable-video

	# Enable audio output
	#enable-audio

	# List of administrative users
	user tim
	user dave

	# Initial size
	size 3

	# Message for new users
	message "hello\tworld"

## License

[MPL 2.0](https://www.mozilla.org/en-US/MPL/2.0/)

## Author

Tim Cooper (<tim.cooper@layeh.com>)
