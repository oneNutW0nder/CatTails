package main

import (
	"fmt"
	"os/exec"
)

func main() {
	out, _ := exec.Command("/bin/bash", "-c", "ls -lah").Output()
	fmt.Println(string(out))
}
