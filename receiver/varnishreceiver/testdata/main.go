package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	fmt.Printf("%#v\n", "hello")
	out, err := exec.Command("date").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", out)

	out, err = exec.Command("sudo", "varnishstat", "-j").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("stats are %s\n", out)

	// cmd := exec.Command("sudo", "cat")
	// cmd.Stdout = os.Stdout
	// stdin, _ := cmd.StdinPipe()
	// file, _ := os.Open("hidden.txt")
	// io.Copy(stdin, file)
	// stdin.Close()
	// _ = cmd.Run()

}
