package auth

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/msteinert/pam"
)

func Test() bool {
	pw := []byte("PASSWORD")

	t, err := pam.StartFunc("login", "lavavrik", func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			fmt.Print("1: ", msg)
			return string(pw), nil
		case pam.PromptEchoOn:
			fmt.Print("2: ", msg)
			s := bufio.NewScanner(os.Stdin)
			s.Scan()
			return s.Text(), nil
		case pam.ErrorMsg:
			fmt.Fprintf(os.Stderr, "3: %s\n", msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println("4: ", msg)
			return "", nil
		default:
			return "", errors.New("unrecognized message style")
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %s\n", err.Error())
		return false
	}
	err = t.Authenticate(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "authenticate: %s\n", err.Error())
		return false
	}
	fmt.Println("authentication succeeded!")
	return true
}
