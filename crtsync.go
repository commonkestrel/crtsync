package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	"example/crtsync/padding"

	"github.com/commonkestrel/cli"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type command struct {
    Name string
    Filename string
    Key string
    Loop bool
    LoopDelay int //In milliseconds
}

var (
    dir string
    store string
    index []command

    keys = []string{"poweron", "poweroff", "volumeup", "stop", "previous", "playpause", "next", "down", "volumedown", "up", "equal", "start", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
)

func init() {
    _, filepath, _, _ := runtime.Caller(0)
    dir = path.Dir(filepath)
    store = path.Join(dir, "store")

    indexfile, err := os.ReadFile(path.Join(store, "index.json"))
    if err != nil {
        if os.IsNotExist(err) {
            return
        }
        panic(err)
    }

    err = json.Unmarshal(indexfile, &index)
    if err != nil {
        fmt.Println("failed to unmarshal index.json. ", err)
    }
}

func main() {
    cli.RegisterCommand("sync", sync)
    cli.RegisterCommand("list", list)
    cli.RegisterCommand("add", add)
    cli.RegisterCommand("remove", remove)
    cli.RegisterCommand("rm", remove)
    cli.RegisterCommand("init", storeinit)
    cli.Default(help)

    if cli.Bool("help", false) {
        help(nil, nil)
        return
    }

    cli.Exec()
    if len(cli.ARGS) > 0 {
        command := cli.ARGS[0]
        if command != "sync" && command != "list" && command != "add" && command != "remove" && command != "rm" && command != "init" {
            fmt.Print("command not recognized. use \"crtsync --help\" to see available commands")
        }
    }
}

func help(args, flags []string) {
    fmt.Println(`
crtsync - tool used for syncing and rendering animations and pictures to a raspberry pi matrix

usage: crtsync <command> [<args>]

commands:
    init - Initializes the store folder with a given ssh key
        usage: crtsync init <path to ssh key>
    add - Adds a command to the index
        usage: crtsync add <name> <filename> <button> [--loop] [--delay=<delay in ms>]
    remove[rm] - Remove a command from the index
        usage: crtsync remove <name>
    list - List all registered commands
        usage: crtsync list
    sync - Sync store repository to the raspberry pi
        usage: crtsync sync`)
}

func writeindex() {
    bytes, err := json.Marshal(index)
    if err != nil {
        fmt.Println("failed to marshal json. ", err)
    }
    
    os.WriteFile(path.Join(store, "index.json"), bytes, 0644)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("index file does not exist. please run init")
        } else {
            fmt.Println("failed to write to index file. ", err)
        }
    }
}

func storeinit(args, flags []string) {
    if len(args) == 0 {
        fmt.Println("must provide path to key file")
        os.Exit(1)
    }

    keypath := args[0]
    _, err := os.Stat(keypath)
    if os.IsNotExist(err) {
        fmt.Println("key file does not exist")
    }

    err = os.Mkdir(store, os.ModeDir)
    if err != nil && !os.IsExist(err) {
        panic(err)
    }

    file, err := os.Create(path.Join(store, "index.json"))
    if err != nil && !os.IsExist(err) {
        panic(err)
    }
    defer file.Close()

    _, err = file.WriteString("[]")
    if err != nil {
        panic(err)
    }

    keyfile, err := os.Create(path.Join(store, "id_rsa"))
    if err != nil {
        panic(err)
    }
    defer keyfile.Close()

    key, err := os.ReadFile(keypath)
    if err != nil {
        panic(err)
    }
    _, err = keyfile.Write(key)
    if err != nil {
        panic(err)
    }
}

func list(args, flags []string) {
    maxlen := [5]int{len("Name"), len("File"), len("Key"), len("Loop"), len("Delay")}
    for _, c := range index {
        if len(c.Name) > maxlen[0] {
            maxlen[0] = len(c.Name)
        }
        if len(c.Filename) > maxlen[1] {
            maxlen[1] = len(c.Filename)
        }
        if len(c.Key) > maxlen[2] {
            maxlen[2] = len(c.Key)
        }
        if len(fmt.Sprint(c.Loop)) > maxlen[3] {
            maxlen[3] = len(fmt.Sprint(c.Loop))
        }
        if len(fmt.Sprint(c.LoopDelay))+2 > maxlen[4] {
            maxlen[4] = len(fmt.Sprint(c.LoopDelay))+2
        }
    }

    fmt.Println(padding.Fill(" ", maxlen[0]+maxlen[1]+maxlen[2]+maxlen[3]+maxlen[4]+15, '_', padding.RIGHT), " ")
    fmt.Printf("| %v | %v | %v | %v | %v |\n", padding.Pad("Name", maxlen[0], padding.EDGES), padding.Pad("File", maxlen[1], padding.EDGES), padding.Pad("Key", maxlen[2], padding.EDGES), padding.Pad("Loop", maxlen[3], padding.EDGES), padding.Pad("Delay", maxlen[4], padding.EDGES))
    fmt.Printf("|-%v-|-%v-|-%v-|-%v-|-%v-|\n", padding.Fill("", maxlen[0], '-', padding.RIGHT), padding.Fill("", maxlen[1], '-', padding.RIGHT), padding.Fill("", maxlen[2], '-', padding.RIGHT), padding.Fill("", maxlen[3], '-', padding.RIGHT), padding.Fill("", maxlen[4], '-', padding.RIGHT))
    for _, c := range index {
        fmt.Printf("| %v | %v | %v | %v | %v |\n", padding.Pad(c.Name, maxlen[0], padding.RIGHT), padding.Pad(c.Filename, maxlen[1], padding.RIGHT), padding.Pad(c.Key, maxlen[2], padding.RIGHT), padding.Pad(c.Loop, maxlen[3], padding.RIGHT), padding.Pad(fmt.Sprint(c.LoopDelay)+"ms", maxlen[4], padding.RIGHT))
    }
    fmt.Print(padding.Fill(" ", maxlen[0]+maxlen[1]+maxlen[2]+maxlen[3]+maxlen[4]+15, '‾', padding.RIGHT), " ")
}

func add(args, flags []string) {
    if len(args) < 3 {
        fmt.Println("incorrect usage. requires at least the name, file name, and key of the command")
        os.Exit(1)
    }

    name, filename, key := args[0], args[1], strings.ToLower(args[2])
    loop := cli.Bool("loop", false)
    loopdelay := cli.Int("delay", 0)

    var valid bool
    for _, k := range keys {
        if k == key {
            valid = true
            break
        }
    }
    if !valid {
        fmt.Println("invalid key. valid keys are ", keys)
        os.Exit(1)
    }

    cmd := command{name, filename, key, loop, loopdelay}

    for _, c := range index {
        if cmd.Name == c.Name || cmd.Key == c.Key {
            fmt.Println("key and name must be unique")
            os.Exit(1)
        }
    }
    index = append(index, cmd)

    writeindex()
    fmt.Printf("added \"%v\" to the index. run \"crtsync list\" to see the current index", cmd.Name)
}

func remove(args, flags []string) {
    if len(args) < 1 {
        fmt.Println("name is a required argument.")
        os.Exit(1)
    }

    name := args[0]

    i := -1
    for pos, c := range index {
        if c.Name == name {
            i = pos
            break
        }
    }
    if i == -1 {
        fmt.Println("no command with the given name exists. use crtsync list to see registered index.")
        os.Exit(1)
    }

    index = append(index[:i], index[i+1:]...)

    writeindex()
    fmt.Printf("removed \"%v\" from the index. run \"crtsync list\" to see the current index", name)
}

func exists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

func getKey() (key ssh.Signer, err error) {
    //usr, _ := user.Current()
    file := path.Join(store, "id_rsa")
    buf, err := os.ReadFile(file)
    if err != nil {
        return
    }
    key, err = ssh.ParsePrivateKey(buf)
    if err != nil {
        return
    }
    return
}

func sync(args, flags []string) {
    for _, c := range index {
        if !exists(path.Join(store, c.Filename)) {
            fmt.Printf("err: %v not found for %v\n", c.Filename, c.Name)
            os.Exit(1)
        }
    }

    fmt.Println("\ncopying index.json -> /home/pi/Documents/Scripts/CRT/store/index.json")
    err := copy(path.Join(store, "index.json"), "/home/pi/Documents/Scripts/CRT/store/index.json")
    if err != nil {
        panic(err)
    }
    
    for _, c := range index {
        fmt.Printf("copying %v -> %v\n", c.Filename, "/home/pi/Documents/Scripts/CRT/store/"+c.Filename)
        err := copy(path.Join(store, c.Filename), "/home/pi/Documents/Scripts/CRT/store/"+c.Filename)
        if err != nil {
            panic(err)
        }
    }
    fmt.Println()
}

func copy(source, destination string) error {
    key, err := getKey()
    if err != nil {
        return err
    }

    config := &ssh.ClientConfig{
        User: "pi",
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(key),
        },
        HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
            return nil
        },
    }

    client, err := ssh.Dial("tcp", "raspberrypi:22", config)
    if err != nil {
        fmt.Println("Failed to dial pi@raspberrypi:22: ", err.Error())
        os.Exit(1)
    }

    session, err := client.NewSession()
    if err != nil {
        fmt.Println("Failed to create session: ", err.Error())
        os.Exit(1)
    }
    defer session.Close()

    err = scp.CopyPath(source, destination, session)
    if err != nil {
        fmt.Printf("Failed to copy %v: %v\n", path.Base(source), err.Error())
    }
    return nil
}
