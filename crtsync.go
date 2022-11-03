package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path"
    "runtime"
    "context"

    "github.com/jibble330/cli"
    "example/crtsync/padding"
    "golang.org/x/crypto/ssh"
    scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
)

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
    cli.RegisterCommand("init", storeinit)

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

    err = os.Mkdir(path.Join(dir, "store"), os.ModeDir)
    if err != nil && !os.IsExist(err) {
        panic(err)
    }

    file, err := os.Create(path.Join(dir, "store", "index.json"))
    if err != nil && !os.IsExist(err) {
        panic(err)
    }
    defer file.Close()

    _, err = file.WriteString("[]")
    if err != nil {
        panic(err)
    }

    keyfile, err := os.Create(path.Join(dir, "store", "id_rsa"))
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
    filestr, err := os.ReadFile(path.Join(dir, "store/index.json"))
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("index file does not exist. please run the init command to initialize the directory.")
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

    fmt.Printf(" %v | %v | %v | %v \n", padding.Pad("Name", maxlen[0], padding.EDGES), padding.Pad("Path", maxlen[1], padding.EDGES), padding.Pad("Key", maxlen[2], padding.EDGES), padding.Pad("Loop", maxlen[3], padding.EDGES))
    fmt.Printf("-%v-|-%v-|-%v-|-%v-\n", padding.Fill("", maxlen[0], '-', padding.RIGHT), padding.Fill("", maxlen[1], '-', padding.RIGHT), padding.Fill("", maxlen[2], '-', padding.RIGHT), padding.Fill("", maxlen[3], '-', padding.RIGHT))
    for _, c := range commands {
        fmt.Printf(" %v | %v | %v | %v \n", padding.Pad(c.Name, maxlen[0], padding.RIGHT), padding.Pad(c.Path, maxlen[1], padding.RIGHT), padding.Pad(c.Key, maxlen[2], padding.RIGHT), padding.Pad(c.Loop, maxlen[3], padding.RIGHT))
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

    filestr, err := os.ReadFile(path.Join(dir, "store/index.json"))
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("index file does not exist. please run the init command to initialize the directory.")
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
    
    err = os.WriteFile(path.Join(dir, "store/index.json"), marshaled, 0644)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("index file does not exist. please run the init command to initialize the directory.")
            os.Exit(1)
        }
        panic(err)
    }
}

func remove(args, flags []string) {

}

func sync(args, flags []string) {
    clientConfig, _ := auth.PrivateKey("pi", path.Join(dir, "store", "id_rsa"), ssh.InsecureIgnoreHostKey())

    client := scp.NewClient("raspberrypi:22", &clientConfig)

    err := client.Connect()
    if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		os.Exit(1)
	}
    defer client.Close()
    
    f, _ := os.Open(path.Join(dir, "store", "index.json"))
    defer f.Close()
    err = client.CopyFromFile(context.Background(), *f, "/home/pi/Documents/Scripts/CRT/CRT/store/index.json", "0655")
    if err != nil {
        panic(err)
    }
}