package main

import (
	"fmt"
	"os"

	"github.com/angelsolaorbaiceta/binmat/signature"
)

// TODO: Read from yaml files in the ./config/binmat/sigs directory
var signatures = []*signature.Signature{
	signature.Make("ELF", "Executable and Linkable Format", []byte{0x7f, 0x45, 0x4c, 0x46}),
	signature.Make("ls", "ls command", []byte{0x74, 0xfc, 0xff, 0xff, 0xc6, 0x05, 0x19, 0x45}),
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file path>\n", os.Args[0])
		os.Exit(1)
	}

	// /bin/ls
	// 00008100: 74fc ffff c605 1945 0000 01c6 050e 4500

	for _, sig := range signatures {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
			os.Exit(1)
		}

		matches, err := sig.CheckMatch(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking signature: %s\n", err)
			os.Exit(1)
		}

		matches.Write(os.Stdout, os.Args[1])
	}
}
