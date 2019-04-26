package flagfile_test

import (
	"flag"
	"fmt"
	"strings"

	"layeh.com/flagfile"
)

const exampleParseStr = `# Enable video output
enable-video

# Enable audio output
#enable-audio

# List of administrative users
user tim cooper
user dave

# Initial size
size 3

# Message for new users
message hello"\t"world
`

type stringSlice []string

func (s stringSlice) String() string {
	return strings.Join([]string(s), ",")
}

func (s *stringSlice) Set(v string) error {
	*s = append([]string(*s), v)
	return nil
}

func ExampleParse() {
	args, err := flagfile.Parse(strings.NewReader(exampleParseStr))
	if err != nil {
		panic(err)
	}

	f := flag.NewFlagSet("example", flag.ContinueOnError)
	enableVideo := f.Bool("enable-video", false, "Enable video output")
	enableAudio := f.Bool("enable-audio", false, "Enable audio output")
	var users stringSlice
	f.Var(&users, "user", "Administrative users")
	size := f.Int("size", 0, "Initial Size")
	message := f.String("message", "", "New user message")

	if err := f.Parse(args); err != nil {
		panic(err)
	}

	fmt.Println("enable-video", *enableVideo)
	fmt.Println("enable-audio", *enableAudio)
	for _, user := range users {
		fmt.Println("user", user)
	}
	fmt.Println("size", *size)
	fmt.Println("message", *message)
	// Output:
	// enable-video true
	// enable-audio false
	// user tim cooper
	// user dave
	// size 3
	// message hello	world
}
