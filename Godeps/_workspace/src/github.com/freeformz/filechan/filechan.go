package filechan

import (
	"bufio"
	"os"
)

// Returns an unbuffered channel, that when read, returns
// individual lines from a file.
// Only returns an error when opening the file. Any other
// errors cause the channel to close.
func FileLineChannel(fpath string) (<-chan string, error) {

	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	c := make(chan string)

	go func() {
		defer close(c)
		defer file.Close()

		buf := bufio.NewReader(file)
		var line string
		var err error

		readLine := func() {
			line, err = buf.ReadString('\n')
		}

		for readLine(); err == nil; readLine() {
			c <- line
		}
	}()

	return c, nil
}
