package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path"
    "runtime"
    "net"

    "github.com/jibble330/cli"
    "example/crtsync/padding"
    "golang.org/x/crypto/ssh"
    "github.com/tmc/scp"
)

type command struct {
    Name string
    Filename string
    Key string
    Loop bool
    LoopDelay int //In milliseconds
}

type helpcommand struct {
    Name, Help string
}

var (
    dir string
    store string
    index []command

    keys = []string{"power", "volumeup", "stop", "previous", "playpause", "next", "down", "volumedown", "up", "equal", "start", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
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
    cli.RegisterCommand("init", storeinit)

    if cli.Bool("help", false) {
        help()
    }

    cli.Exec()
}

func help() {
    index := []helpcommand{}
    if len(cli.ARGS) > 0 {
        com := cli.ARGS[0]
        for _, c := range index {
            if c.Name == com {
                fmt.Println(c.Help)
                return
            }
        }
        fmt.Println("Please ")
    }
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
    maxlen := [4]int{len("Name"), len("File"), len("Key"), len("Loop")}
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
    }

    fmt.Printf(" %v | %v | %v | %v \n", padding.Pad("Name", maxlen[0], padding.EDGES), padding.Pad("File", maxlen[1], padding.EDGES), padding.Pad("Key", maxlen[2], padding.EDGES), padding.Pad("Loop", maxlen[3], padding.EDGES))
    fmt.Printf("-%v-|-%v-|-%v-|-%v-\n", padding.Fill("", maxlen[0], '-', padding.RIGHT), padding.Fill("", maxlen[1], '-', padding.RIGHT), padding.Fill("", maxlen[2], '-', padding.RIGHT), padding.Fill("", maxlen[3], '-', padding.RIGHT))
    for _, c := range index {
        fmt.Printf(" %v | %v | %v | %v \n", padding.Pad(c.Name, maxlen[0], padding.RIGHT), padding.Pad(c.Filename, maxlen[1], padding.RIGHT), padding.Pad(c.Key, maxlen[2], padding.RIGHT), padding.Pad(c.Loop, maxlen[3], padding.RIGHT))
    }
}

func add(args, flags []string) {
    if len(args) < 3 {
        fmt.Println("incorrect usage. requires at least the name, key, and path of the command")
        os.Exit(1)
    }

    name, filename, key := args[0], args[1], args[2]
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
            fmt.Printf("missing file %v for %v.\n", c.Filename, c.Name)
            os.Exit(1)
        }
    }

    err := copy(path.Join(store, "index.json"), "/home/pi/Documents/Scripts/CRT/store/index.json")
    if err != nil {
        panic(err)
    }
    
    for _, c := range index {
        err := copy(path.Join(store, c.Filename), "/home/pi/Documents/Scripts/CRT/store/"+c.Filename)
        if err != nil {
            panic(err)
        }
    }
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