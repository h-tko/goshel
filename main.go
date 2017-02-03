package main

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "github.com/h-tko/sshconfig-parser"
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
    CONF_FILE = "/.goshel_conf"
)

const (
    HOST = iota
    ALIAS
    PORT
    NAME
    IDENTITYFILE
)

func init() {
    confDir, err := homedir.Dir()

    if err != nil {
        log.Fatalf("%v", err)
    }

    FullFileName = confDir + CONF_FILE
}

func main() {

    f := flag.Bool("l", false, "接続先一覧")

    flag.Parse()

    if *f {
        list, err := sshList()

        if err != nil {
            log.Fatalf("%v", err)
            os.Exit(1)
        }

        showList(list)

        os.Exit(0)
    }

    for {
        var proc string

        println("なにします？")
        println("1) ssh実行")
        println("2) 接続先追加")
        println("3) ssh_config読み込み")
        println("8) 接続先削除")
        println("99) 初期化")
        println("q) Exit")

        fmt.Scanln(&proc)

        switch proc {
        case "1":
            if err := startssh(); err != nil {
                fmt.Errorf("%v", err)
            }

            os.Exit(1)
        case "2":
            if err := configure(); err != nil {
                log.Fatalf("%v", err)
            }
        case "3":
            hosts, err := loadSSHConfig()

            if err != nil {
                println("ssh_configファイルを見つけることができませんでした、ごめんね")
                log.Fatalf("%v", err)

                os.Exit(1)
            }

            if err := addFromSSHConfig(hosts); err != nil {
                log.Fatalf("%v", err)

                os.Exit(1)
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

    println("接続ユーザーを設定してください。未入力の場合、現在のユーザー名が設定されます")

    var name string
    fmt.Scanln(&name)

    if name == "" {
        curUser, err := user.Current()

        if err != nil {
            return err
        }

        name = curUser.Username
    }

    println("鍵認証の場合、認証ファイルのファイル名をフルパスで指定してください。未入力の場合、パスワード認証とみなされます")

    var identityfile string
    fmt.Scanln(&identityfile)

    if err := addConfig(host, alias, port, name, identityfile); err != nil {
        return err
    }

    return nil
}

func addConfig(host, alias, port, name, identityfile string) error {
    line := []byte(fmt.Sprintf("%s,%s,%s,%s,%s\n", host, alias, port, name, identityfile))

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
        fmt.Printf("%v", err)
        return nil
    }

    if err := execssh(list[selected-1]); err != nil {
        return err
    }

    return nil
}

func showList(list [][]string) {
    for index, rec := range list {

        key := ""

        if len(rec[IDENTITYFILE]) > 0 {
            key = "🔑"
        }

        fmt.Printf("%d) %s [%s]%s\n", index+1, rec[ALIAS], rec[HOST], key)
    }
}

func showAndSelectList(list [][]string) (int, error) {

    showList(list)

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
    identityfile := target[IDENTITYFILE]

    var cmd *exec.Cmd

    if len(identityfile) > 0 {
        cmd = exec.Command("ssh", fmt.Sprintf("%s@%s", name, host), fmt.Sprintf("-p%s", port), fmt.Sprintf("-i%s", identityfile))
    } else {
        cmd = exec.Command("ssh", fmt.Sprintf("%s@%s", name, host), fmt.Sprintf("-p%s", port))
    }

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

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

    list = deleteElement(list, selected-1)

    clearConfig()

    for _, data := range list {
        if err := addConfig(data[HOST], data[ALIAS], data[PORT], data[NAME], data[IDENTITYFILE]); err != nil {
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

func loadSSHConfig() ([]*sshconfig.SSHConfig, error) {
    println("ssh_configのフルパスを指定してください（未入力の場合「~/.ssh/config」もしくは「〜/.ssh/ssh_config」を読み込みます")

    var hosts []*sshconfig.SSHConfig
    var sshConfigPath string
    fmt.Scanln(&sshConfigPath)

    if len(sshConfigPath) > 0 {
        var err error
        hosts, err = sshconfig.Parse(sshConfigPath)

        if err != nil {
            return nil, err
        }
    } else {
        confDir, err := homedir.Dir()

        if err != nil {
            log.Fatalf("%v", err)
        }

        hosts, err = sshconfig.Parse(confDir + "/.ssh/config")

        if err != nil {

            hosts, err = sshconfig.Parse(confDir + "/.ssh/ssh_config")

            if err != nil {
                return nil, err
            }
        }
    }

    return hosts, nil
}

func addFromSSHConfig(hosts []*sshconfig.SSHConfig) error {
    println("登録モードを選んでください")
    println("1) 追記")
    println("2) スクラップアンドビルド")

    var div string
    fmt.Scan(&div)

    switch div {
    case "1":
        if err := addSSHHostList(hosts); err != nil {
            return err
        }

    case "2":
        clearConfig()

        if err := addSSHHostList(hosts); err != nil {
            return err
        }

    default:
        return errors.New("範囲外が選択されました")
    }

    return nil
}

func addSSHHostList(hosts []*sshconfig.SSHConfig) error {
    for _, host := range hosts {

        fmt.Printf("%sを%sとして追加\n", host.HostName, host.Host)

        if len(host.Host) > 0 && len(host.HostName) < 1 {
            host.HostName = host.Host
        }

        if err := addConfig(host.HostName, host.Host, strconv.Itoa(host.Port), host.User, host.IdentityFile); err != nil {
            return err
        }
    }

    return nil
}

func showHostList(hostname string, host []string) (int, error) {
    fmt.Printf("%sのホスト名が複数設定されているので、選んでください", hostname)

    for index, d := range host {
        fmt.Printf("%d) %s\n", index+1, d)
    }

    var selected string
    fmt.Scan(&selected)

    selectedIndex, err := strconv.Atoi(selected)

    if err != nil || len(host)+1 < selectedIndex {
        return -1, errors.New("範囲外が選択されました")
    }

    return selectedIndex, nil
}
