# 2026-11-14 初回デプロイ

<!-- markdownlint-disable MD013 -->

AWS に初めてデプロイした。構成の決定から、DB のバックアップの設定まで行った。

## 決めたこと

- **構成**: EC2 1台で、API・PostgreSQL・Caddy を Docker Compose で動かす。月 $13〜14 程度。
  - ECS + ALB + RDS（東京で月 $60 程度）や、Fargate（月 $20〜26 程度）も検討したが、費用を優先した。
- **リージョン**: ap-southeast-2（シドニー）。新しいサインアップ方式で、アカウントのリージョンが固定されていたため。
- **インスタンス**: t4g.micro（Arm64）。同じ性能の x86 より安いため。
- **イメージの受け渡し**: レジストリ（ECR）は使わず、`docker save` → scp → `docker load` で送る。
- **イメージのバージョン**: git のタグと合わせる。最初は v1.0.0。
- **HTTPS**: Caddy で Let's Encrypt の証明書を自動取得する。ALB + ACM は月 $16 以上かかるため使わない。
- **TRUSTED_PROXIES**: 設定する。フロント経由の利用者は INTERNAL_API_SECRET で IP を判断できるが、API を直接呼ぶ相手の IP を正しく扱うため。
- **バックアップ**: `pg_dump` のスクリプトを systemd timer で毎日実行し、非公開の S3 バケットに送る。EBS スナップショット、他人のバックアップ用イメージ、RDS も検討した。

## 行った作業

1. AWS アカウントを作成した。MFA と回復用メールを設定し、有料プランに切り替えた。支出上限は「Stop new resource launches」のみ有効にした。
2. EC2（t4g.micro、Amazon Linux 2023、20GB gp3）を起動し、Elastic IP を付けた。
3. サーバーに Docker、Docker Compose プラグイン、git、vim-enhanced を入れた。
4. リポジトリに `Dockerfile`、`.dockerignore`、`scripts/build-image.sh`、`docker-compose.prod.yml` を作成した。v1.0.0 のイメージを作ってサーバーに送り、API と PostgreSQL を起動した。
5. 画像用の S3 バケットと、EC2 用の IAM ロール `movie-log-api-ec2`（インラインポリシー `movie-log-api-s3`）を作成し、EC2 に付けた。
6. `api.cinelier.app` の A レコードを Elastic IP に向けた。`Caddyfile` を作成し、compose に caddy を追加した。ネットワークのサブネットを `172.30.0.0/24` に固定し、`.env` に `TRUSTED_PROXIES` を追加した。
7. バックアップ用の S3 バケット、IAM ポリシー `movie-log-api-backup`、`scripts/backup-db.sh`、systemd の `movie-log-backup.service` と `movie-log-backup.timer`、30日で削除するライフサイクルルールを作成した。

## 確認したこと

- PC から `curl -i https://api.cinelier.app/` で `HTTP/2 404`、`via: 1.1 Caddy` が返った。
- Caddy のログに `certificate obtained successfully` が出た。
- API のログに、外部の本当の IP が出た（TRUSTED_PROXIES が正しく働いている）。証明書の取得直後から、WordPress などの弱点を探すボットのアクセスが来ていた。すべて 404。
- バックアップを手動と systemd 経由で1回ずつ実行し、S3 に届いた。
- ダウンロードしたバックアップを試験用の DB `restore_test` に戻し、9つのテーブルがあることを確認した。確認後に削除した。

## 起きた問題

### SSH が応答しない

セキュリティグループの SSH ルールで「My IP」を選んだら、IPv6 の範囲が入っていた。自宅からは IPv4 で接続しているため、許可されなかった。そのルールを削除し、IPv4 の `/32` で追加し直して解決した。

### API が DB のパスワード認証に失敗する

API のログに `password authentication failed` が出た。原因は2つあった。

1. `.env` が `POSTGRES_USER=postgres` なのに、`POSTGRES_DSN` は `movie_log` ユーザーを指していた。postgres のログの `DETAIL` 行に `Role "movie_log" does not exist` と出ていた。`POSTGRES_USER=movie_log` に直し、`docker compose down -v` で DB を作り直した。
2. それでも失敗した。原因は、確認のためにシェルで `set -a; . ./.env` を実行し、古い値を環境変数に読み込んでいたこと。compose はシェルの環境変数を `.env` より優先する。SSH を切って入り直し、`down -v` と `up -d` をやり直して解決した。

最初は API 側のログだけを見て原因を推測し、時間がかかった。postgres 側のログの `DETAIL` 行を先に見るべきだった。

### distroless の debian12 が更新されていない

Dockerfile の実行用イメージに `gcr.io/distroless/static-debian12` を使おうとしたが、更新が止まっていた。`static-debian13:nonroot` に変えた。

## 残した作業

- OS の自動更新（8）。`/etc/dnf/vars/releasever` に `latest` を書き、`dnf-automatic` を入れて `dnf-automatic-install.timer` を有効にする予定。途中で中断したため、どのコマンドまで実行したかは未確認。`/etc/dnf/automatic.conf` がないことは確認済み（dnf-automatic は入っていない）。
- カーネル更新後の再起動の方法を決める。
- Web アプリのデプロイ時に、同じ `INTERNAL_API_SECRET` を渡す。
