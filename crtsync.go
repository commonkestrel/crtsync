package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/jibble330/cli"
	/*"golang.org/x/crypto/ssh"
	  "github.com/bramvdbogaerde/go-scp/auth"
	  scp "github.com/bramvdbogaerde/go-scp"*/)

type command struct {
    Name string
    Path string
    Key string
    Loop bool
    LoopDelay int //In milliseconds
}

type helpcommand struct {
    Name, Help string
}

var dir string

func main() {
    cli.RegisterCommand("sync", sync)
    cli.RegisterCommand("list", list)
    cli.RegisterCommand("add", add)
    cli.RegisterCommand("remove", remove)
    cli.RegisterCommand("init", cominit)

    if cli.Bool("help", false) {
        help()
    }

    cli.Exec()

    _, filepath, _, _ := runtime.Caller(0)
    dir = path.Dir(filepath)
    fmt.Println(dir)
}

func help() {
    commands := []helpcommand{}
    if len(cli.ARGS) > 0 {
        com := cli.ARGS[0]
        for _, c := range commands {
            if c.Name == com {
                fmt.Println(c.Help)
                return
            }
        }
        fmt.Println("Please ")
    }
}

func cominit(args, flags []string) {
    
}

func list(args, flags []string) {
    filestr, err := os.ReadFile(path.Join(dir, "lock/lock.json"))
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("lock file does not exist. please run the init command to initialize the directory.")
            os.Exit(1)
        }
        panic(err)
    }

    bytes := []byte(filestr)

    var commands []command
    err = json.Unmarshal(bytes, &commands)
    if err != nil {
        panic(err)
    }

    maxlen := [4]int{len("Name"), len("Key"), len("Path"), len("Loop")}
    for _, c := range commands {
        if len(c.Name) > maxlen[0] {
            maxlen[0] = len(c.Name)
        }
        if len(c.Path) > maxlen[1] {
            maxlen[1] = len(c.Path)
        }
        if len(c.Key) > maxlen[2] {
            maxlen[2] = len(c.Key)
        }
        if len(fmt.Sprint(c.Loop)) > maxlen[3] {
            maxlen[3] = len(fmt.Sprint(c.Loop))
        }
    }

    fmt.Printf("%*s | %*s | %*s | %*s\n\n", maxlen[0]-len("Name"), "Name", maxlen[1]-len("Path"), "Path", maxlen[2]-len("Key"), "Key", maxlen[3]-len("Loop"), "Loop")
    for _, c := range commands {
        fmt.Printf("%*s | %*s | %*s | %*s\n", maxlen[0]-len(c.Name), c.Name, maxlen[1]-len(c.Path), c.Path, maxlen[2]-len(c.Key), c.Key, maxlen[3]-len(fmt.Sprint(c.Loop)), fmt.Sprint(c.Loop))
    }
}

func add(args, flags []string) {
    if len(args) < 3 {
        fmt.Println("incorrect usage. requires at least the name, key, and path of the command")
    }

    name, animpath, key := args[0], args[1], args[2]
    loop := cli.Bool("loop", false)
    loopdelay := cli.Int("delay", 0)

    cmd := command{name, animpath, key, loop, loopdelay}

    filestr, err := os.ReadFile(path.Join(dir, "lock/lock.json"))
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("lock file does not exist. please run the init command to initialize the directory.")
            os.Exit(1)
        }
        panic(err)
    }

    bytes := []byte(filestr)

    var commands []command
    err = json.Unmarshal(bytes, &commands)
    if err != nil {
        panic(err)
    }
    
    for _, c := range commands {
        if cmd.Name == c.Name || cmd.Key == c.Key {
            fmt.Println("key and name must be unique")
            os.Exit(1)
        }
    }

    commands = append(commands, cmd)
    fmt.Println(cmd, commands)

    

    marshaled, err := json.Marshal(commands)
    if err != nil {
        panic(err)
    }
    
    err = os.WriteFile(path.Join(dir, "lock/lock.json"), marshaled, 0644)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("lock file does not exist. please run the init command to initialize the directory.")
            os.Exit(1)
        }
        panic(err)
    }
}

func remove(args, flags []string) {

}

func sync(args, flags []string) {
	//client := scp.NewClient("example.com:22", &clientConfig)
}