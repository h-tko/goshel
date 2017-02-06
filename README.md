# Goshel

対話形式でssh先を設定・保存しておいて、選択することでsshするツールです。
接続先が多くてわちゃってきた時用。

## Installation

```
$ go get -u github.com/h-tko/goshel
```

## What I can do

現時点で実装が完了しているのは以下のあたりです。

+ ssh先を登録する
   + パスワード接続 
   + 鍵認証接続
   + ssh_configからの一括取り込み
+ 登録済みのssh先を一覧表示する
+ 登録されているssh先に接続する

## How to Use

- 使う

```
$ goshel
```

- 登録されている接続先一覧を見る

```
$ goshel -l
```

## License

[MIT](/LICENSE)
