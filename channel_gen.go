// +build ignore

package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()

	gofmt := os.Getenv("GOROOT") + "/bin/gofmt"
	log.Print("use 'gofmt' at ", gofmt)

	var cmdErr bytes.Buffer
	var cmd1Out bytes.Buffer
	var cmd2Out bytes.Buffer
	var cmd3Out bytes.Buffer

	log.Print("gofmt -r 'Channel -> " + flag.Arg(0) + "'")
	cmd1 := exec.Command(gofmt, "-r", "Channel -> "+flag.Arg(0), "channel.go")
	cmd1.Stdout = &cmd1Out
	cmd1.Stderr = &cmdErr
	if err := cmd1.Run(); err != nil {
		log.Fatal(cmdErr.String())
	}

	log.Print("gofmt -r 'NewChannel -> New" + flag.Arg(0) + "'")
	cmd2 := exec.Command(gofmt, "-r", "NewChannel -> New"+flag.Arg(0))
	cmd2.Stdin = &cmd1Out
	cmd2.Stdout = &cmd2Out
	cmd2.Stderr = &cmdErr
	if err := cmd2.Run(); err != nil {
		log.Fatal(cmdErr.String())
	}

	log.Print("gofmt -r 'KEY -> " + flag.Arg(1) + "'")
	cmd3 := exec.Command(gofmt, "-r", "KEY -> "+flag.Arg(1))
	cmd3.Stdin = &cmd2Out
	cmd3.Stdout = &cmd3Out
	cmd3.Stderr = &cmdErr
	if err := cmd3.Run(); err != nil {
		log.Fatal(cmdErr.String())
	}

	cmd3Out.ReadBytes('\n') // ignore build
	cmd3Out.ReadBytes('\n') // empty line
	cmd3Out.ReadBytes('\n') // Int32Channel
	cmd3Out.ReadBytes('\n') // Uint32Channel
	cmd3Out.ReadBytes('\n') // Int64Channel
	cmd3Out.ReadBytes('\n') // Uint64Channel
	cmd3Out.ReadBytes('\n') // StringChannel

	log.Print("save to target file '", flag.Arg(2), "'")
	file, err := os.Create(flag.Arg(2))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString(
		"// DO NOT EDIT\n// GENERATE BY 'go run channel_gen.go " +
			flag.Arg(0) + " " + flag.Arg(1) + " " + flag.Arg(2) + "'\n"); err != nil {
		log.Fatal(err)
	}

	var code = cmd3Out.Bytes()

	if len(flag.Args()) > 3 {
		log.Print("rename package as '" + flag.Arg(3) + "'")
		code = bytes.Replace(code, []byte("package link"), []byte("package "+flag.Arg(3)), 1)
	}

	if _, err := file.Write(code); err != nil {
		log.Fatal(err)
	}
}
