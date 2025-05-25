# audition-marker_2_mp3-id3-tag

このツールは、Adobe Audition が出力するマーカーの CSV ファイルを読み取り、MP3 ファイルに ID3v2 形式のチャプターとして埋め込みます。

## 機能

- Adobe Audition のマーカー CSV ファイルを解析
- MP3 ファイルに ID3v2 チャプタータグを追加
- 入力ファイルを上書きするか、別のファイルとして保存するかを選択可能

## 使用方法

```sh
go run ./... -csv <CSVファイルパス> -input <入力MP3パス> [-output <出力MP3パス>]
```

### オプション

- `-csv`: Adobe Audition のマーカー CSV ファイルのパス（必須）
- `-input`: チャプターを追加する元の MP3 ファイルのパス（必須）
- `-output`: チャプターを追加した MP3 ファイルの出力パス（指定しない場合は "ファイル名_with_chapters.mp3" として出力）

## 例

チャプターを追加して "podcast_with_chapters.mp3" として保存:

```sh
go run ./... -csv "marker.csv" -input "podcast.mp3"
```

別ファイルとして保存:

```sh
go run ./... -csv "marker.csv" -input "podcast.mp3" -output "podcast_with_chapters.mp3"
```
