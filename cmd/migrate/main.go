package main

import (
	"fmt"
	"os"
	"os/exec"
)

// run migrations
func main() {
	args := append([]string{"-path", "migrations", "-database", os.Getenv("DATABASE_URL"), "up"}, os.Args[1:]...)
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}

	cmd := exec.Command("migrate", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if e := cmd.Run(); e != nil {
		os.Exit(1)
	}
}
