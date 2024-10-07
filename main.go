package main

import (
	"fmt"
	"os"

	"github.com/angelsolaorbaiceta/binmat/signature"
)

// TODO: Read from yaml files in the ./config/binmat/sigs directory
var signatures = signature.Signatures{
	signature.Make("ELF", "Executable and Linkable Format", []byte{0x7f, 0x45, 0x4c, 0x46}),
	signature.Make("ls", "ls command", []byte{0x74, 0xfc, 0xff, 0xff, 0xc6, 0x05, 0x19, 0x45}),
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file path>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	fmt.Fprintf(os.Stderr, "Checking matches in file: %s...\n", filePath)

	// /bin/ls
	// 00008100: 74fc ffff c605 1945 0000 01c6 050e 4500

	matches, err := signatures.Check(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking signatures: %s\n", err)
		os.Exit(1)
	}

	for _, match := range matches {
		match.Write(os.Stdout, filePath)
	}
}
