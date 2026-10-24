# japan-mobile-career-ip-address
日本のモバイルキャリアIPアドレスを、jsonで提供します。

IPv4 確認日: **2026-10-24**。今回公式一覧から更新できたのは docomo と au です。
SoftBank は本文を取得できず、楽天は携帯回線向けの公式一覧を確認できなかったため、既存データを維持しています。
全キャリアの最新・完全な一覧を保証するものではありません。
確認範囲、保留した範囲、キャリア別差分は [SOURCES.md](SOURCES.md) を参照してください。
JSON の `ipv4` / `ipv6` 形式は維持しています。今回 `ipv6` は更新していません。

# キャリア

https://github.com/yas-nyan/IP-Address-Ranges-for-Japanese-Mobile-Carrers
を参考にさせていただきました。

## NTTドコモ

https://www.docomo.ne.jp/service/developer/smart_phone/spmode/index.html

## au

https://www.au.com/mobile/charge/internet-connection/network/

LTE NET／5G NET の公式一覧（2025年1月現在）。`106.134.0.0/16` と `106.135.0.0/16` の除外指定を適用します。
LTE NET for DATA／5G NET for DATA の範囲は非公開です。

## softbank

https://s4.itc.softbank.jp/partner/home_tech1/index.cfm
https://s4.itc.softbank.jp/partner_st/home_tech1/ios/index.cfm
https://s4.itc.softbank.jp/partner_st/home_tech1/X_series/index.cfm

旧 `www.support.softbankmobile.co.jp` の出典も今回再確認を試みましたが、本文を取得できませんでした。
取得失敗を公開終了とは判定していません。


## 楽天モバイル


既存データの由来は第三者の ASN 一覧です（今回の更新根拠には使用していません）。

- https://ipinfo.io/AS138384
- https://ipinfo.io/AS23720
- https://ipinfo.io/AS10012

携帯回線専用と確認できておらず、固定回線等を誤検知する可能性があります。新たな ASN 帯域は追加していません。

# 更新・テスト

`calc/` の入力から IPv4 を生成します。ドコモの `docomo_v4.json` は歴史的なファイル名で、内容は1行1 CIDRです。
生成時には、公式情報の用途・追加完了日を確認し、予定・メール専用の範囲を分けてください。
入力と JSON の対応は `calc/main_test.go` で検証します。

```sh
cd calc
go run . docomo_v4.json
go run . au_v4.ip
go test ./...
go vet ./...
```

上記コマンドの `ipv4` を `data/<carrier>.json` に反映します。今回対象外の `ipv6` はそのまま維持してください。
Go モジュールへの同期手順は [SOURCES.md](SOURCES.md) に記載しています。
