package main

import (
    "bufio"
    "errors"
    "fmt"
    "github.com/mitchellh/go-homedir"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "os/user"
    "strconv"
    "strings"
)

var FullFileName string

const (
    CONF_FILE = "/.gossh_conf"
)

const (
    HOST = iota
    ALIAS
    PORT
    NAME
)

func init() {
    confDir, err := homedir.Dir()

    if err != nil {
        log.Fatalf("%v", err)
    }

    FullFileName = confDir + CONF_FILE
}

func main() {

    for {
        var proc string

        println("なにします？")
        println("1) ssh実行")
        println("2) 接続先追加")
        println("8) 接続先削除")
        println("99) 初期化")
        println("q) Exit")

        fmt.Scanln(&proc)

        switch proc {
        case "1":
            if err := startssh(); err != nil {
                panic(err)
            }

        case "2":
            if err := configure(); err != nil {
                log.Fatalf("%v", err)
            }

        case "8":
            deleteConfig()

        case "99":
            clearConfig()

            println("設定を初期化しました")
        case "q":
            os.Exit(0)

        default:
            usage()
        }

        print("\n\n\n\n\n")
    }
}

func usage() {
    println("想定外の選択がされました")
}

func configure() error {
    println("接続先のIPアドレス、またはホスト名を指定してください")

    var host string
    fmt.Scan(&host)

    println("接続先のポートを指定してください。未入力の場合は22が設定されます")

    var port string
    fmt.Scanln(&port)

    if port == "" {
        port = "22"
    }

    println("接続先に名前を設定してください。未入力の場合はIPアドレス、またはホスト名が設定されます")

    var alias string
    fmt.Scanln(&alias)

    if alias == "" {
        alias = host
    }

    var name string
    fmt.Scanln(&name)

    println("接続ユーザーを設定してください。未入力の場合、現在のユーザー名が設定されます")

    if name == "" {
        curUser, err := user.Current()

        if err != nil {
            return err
        }

        name = curUser.Username
    }

    if err := addConfig(host, alias, port, name); err != nil {
        return err
    }

    return nil
}

func addConfig(host, alias, port, name string) error {
    line := []byte(fmt.Sprintf("%s,%s,%s,%s\n", host, alias, port, name))

    file, err := os.OpenFile(FullFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)

    writer.Write(line)
    writer.Flush()

    return nil
}

func clearConfig() {
    ioutil.WriteFile(FullFileName, []byte(""), os.ModePerm)
}

func startssh() error {
    println("接続先を選択してください")
    println("")

    list, err := sshList()

    if err != nil {
        return err
    }

    selected, err := showAndSelectList(list)

    if err != nil {
        return err
    }

    if err := execssh(list[selected-1]); err != nil {
        return err
    }

    return nil
}

func showAndSelectList(list [][]string) (int, error) {
    for index, rec := range list {
        fmt.Printf("%d) %s [%s]\n", index+1, rec[ALIAS], rec[HOST])
    }

    var selectedIndex string
    fmt.Scanln(&selectedIndex)

    selected, err := strconv.Atoi(selectedIndex)
    if err != nil {
        return -1, errors.New("数字で選択してください")
    }

    return selected, nil
}

func sshList() ([][]string, error) {
    var list [][]string

    file, err := os.Open(FullFileName)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := bufio.NewReaderSize(file, 4096)
    for line := ""; err == nil; line, err = reader.ReadString('\n') {
        if line != "" {
            line = strings.TrimRight(line, "\n")
            list = append(list, strings.Split(line, ","))
        }
    }

    return list, nil
}

func execssh(target []string) error {
    host := target[HOST]
    port := target[PORT]
    name := target[NAME]

    cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", name, host), fmt.Sprintf("-p%s", port))
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout

    if err := cmd.Run(); err != nil {
        return err
    }

    return nil
}

func deleteConfig() error {
    list, err := sshList()

    if err != nil {
        return err
    }

    selected, err := showAndSelectList(list)

    if err != nil {
        return err
    }

    list = deleteElement(list, selected - 1)

    clearConfig()

    for _, data := range list {
        if err := addConfig(data[HOST], data[ALIAS], data[PORT], data[NAME]); err != nil {
            return err
        }
    }

    return nil
}

func deleteElement(list [][]string, target int) [][]string {
    var result [][]string

    for index, data := range list {
        if index != target {
            result = append(result, data)
        }
    }

    return result
}
