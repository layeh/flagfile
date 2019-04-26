// Package flagfile converts a flag file into flag-compatible arguments.
//
// A flag file is alternative way to provide command-line flags for a program.
// For example, the below arguments:
//
//  ./example --enable-video --user=tim --user=dave --size 3 --message="hello\tworld"
//
// Would be stored in the flag file as the following:
//
//  # Enable video output
//  enable-video
//
//  # Enable audio output
//  #enable-audio
//
//  # List of administrative users
//  user tim
//  user dave
//
//  # Initial size
//  size 3
//
//  # Message for new users
//  message "hello\tworld"
package flagfile // import "layeh.com/flagfile"
