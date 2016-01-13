# tscutmp4
tsファイルをaviutlでカットしてx264に変換する

## 構成
現時点で以下の2段階の構成により変換している。

```
1. 入力のtsファイルを、マルチバイトを含まないファイルにコピー
2. DGIndex.exeで処理する
3. BonTsDemuxC.exeでDemux(wav)する
4. aviutl.exeで開く
   →その後、手動でtrim()を作成し、input.ts.avs内に追記する
```

```
1. input.ts.avsをx264.exeでencodeする
2. mp4boxで組み立てる
```

## 新しくつくりたい部分

- aviutlで開直前までの処理を行い、キューイングする
- encode + mp4box の部分を指定時間から実行する
