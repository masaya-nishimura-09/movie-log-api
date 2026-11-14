# 本番環境（AWS）

<!-- markdownlint-disable MD013 -->

2026-11-14 に構築した本番環境の構成と、運用の手順をまとめる。

## 構成

```text
インターネット
   │ 80 / 443
   ▼
EC2（t4g.micro / Amazon Linux 2023 / ap-southeast-2）
   └ Docker のネットワーク（172.30.0.0/24）
       ├ caddy     … HTTPS の終端。api.cinelier.app → api:8080
       ├ api       … movie-log-api（127.0.0.1:8080 のみ公開）
       └ postgres  … PostgreSQL 16（外部に非公開）

S3
   ├ cinelier-media-648358186766-ap-southeast-2-an   … 画像（誰でも読める）
   └ cinelier-backup-648358186766-ap-southeast-2-an  … DB のバックアップ（非公開）
```

## AWS のリソース

### アカウント

- 新しいサインアップ方式（Builder ID）。リージョンは ap-southeast-2（シドニー）で固定。
- MFA は settings.aws.com の「Manage security and sign-in」で管理する。
- 有料プランに切り替え済み。プロジェクトの支出上限で「Stop new resource launches」のみ有効。

### EC2

- インスタンスタイプ: t4g.micro（Arm64）
- OS: Amazon Linux 2023
- ディスク: 20GB gp3
- Elastic IP を割り当て済み
- キーペア: PC の `~/.ssh/movie-log-api.pem`
- IAM ロール: `movie-log-api-ec2`
- セキュリティグループ:
  - SSH（22）: 自宅の IPv4 アドレス（/32）のみ
  - HTTP（80）、HTTPS（443）: すべて許可

自宅の IP が変わると SSH できなくなる。その場合は、セキュリティグループの SSH ルールを新しい IP に変更する。
「My IP」を選ぶと IPv6 の範囲が入ることがあり、その場合は接続できない。IPv4 の `/32` を手で入れる。

### IAM ロール `movie-log-api-ec2` のインラインポリシー

| ポリシー名 | 許可 | 対象 |
|---|---|---|
| `movie-log-api-s3` | `s3:PutObject`、`s3:DeleteObject` | 画像バケットの `/*` |
| `movie-log-api-s3` | `s3:ListBucket` | 画像バケット |
| `movie-log-api-backup` | `s3:PutObject` のみ | バックアップバケットの `/*` |

バックアップバケットは書き込みのみ許可している。サーバーが乗っ取られても、バックアップを読まれたり消されたりしないようにするため。

### S3

#### 画像バケット（`cinelier-media-...`）

- Block Public Access は、バケットポリシーに関する2項目のみオフ。
- バケットポリシーで、全オブジェクトの `s3:GetObject` を誰にでも許可している。

#### バックアップバケット（`cinelier-backup-...`）

- Block Public Access はすべてオン。
- ライフサイクルルール `delete-after-30-days`: 作成から30日を過ぎたファイルを自動で削除する。

### DNS

- `api.cinelier.app` の A レコード → Elastic IP

## サーバーの中身

```text
/home/ec2-user/movie-log-api/
├ compose.yaml          … リポジトリの docker-compose.prod.yml をコピーしたもの
├ Caddyfile             … リポジトリの Caddyfile をコピーしたもの
├ .env                  … 本番の環境変数（chmod 600。リポジトリには入れない）
└ scripts/
   ├ create_tables.sql  … DB の初回起動時にテーブルを作る
   └ backup-db.sh       … DB のバックアップを S3 に送る

/etc/systemd/system/
├ movie-log-backup.service
└ movie-log-backup.timer
```

### サーバーに入れたもの

- `docker`（`sudo dnf install -y docker`、`sudo systemctl enable --now docker`）
- ec2-user を docker グループに追加（`sudo usermod -aG docker ec2-user`）
- Docker Compose プラグイン（GitHub のリリースから `/usr/local/lib/docker/cli-plugins/docker-compose` に配置）
- `git`、`vim-enhanced`

### `.env` の項目

値はサーバーの `.env` にのみある。秘密の値はチャットやリポジトリに貼らない。

| 項目 | 内容 |
|---|---|
| `APP_VERSION` | 起動するイメージのタグ（例: `v1.0.0`） |
| `GIN_MODE` | `release` |
| `POSTGRES_DB` / `POSTGRES_USER` | `movie_log` / `movie_log` |
| `POSTGRES_PASSWORD` | `openssl rand -hex` で生成 |
| `POSTGRES_DSN` | `postgres://movie_log:<パスワード>@postgres:5432/movie_log?sslmode=disable` |
| `JWT_SECRET` | `openssl rand -hex` で生成（32バイト以上） |
| `CLEANUP_ENABLE` / `CLEANUP_INTERVAL_SECOND` | `true` / `3600` |
| `TMDB_*` | PC の `.env` と同じ値 |
| `AWS_REGION` | `ap-southeast-2` |
| `S3_BUCKET` | 画像バケット名 |
| `S3_PUBLIC_BASE_URL` | `https://<画像バケット名>.s3.ap-southeast-2.amazonaws.com` |
| `INTERNAL_API_SECRET` | `openssl rand -hex` で生成。Web アプリをデプロイするときに同じ値を渡す |
| `TRUSTED_PROXIES` | `172.30.0.0/24` |

S3 の認証情報は `.env` に書かない。EC2 の IAM ロールから自動で取得される。

## 仕組みの要点

### Caddy（HTTPS）

- Caddyfile にドメインを書くと、Let's Encrypt から証明書を自動で取得・更新する。
- HTTP（80）へのアクセスは、自動で HTTPS にリダイレクトされる。
- 証明書はボリューム `caddy-data` に保存される。コンテナを作り直しても残る。

### TRUSTED_PROXIES

API はアクセス元の IP ごとに回数制限をしている（ログイン、会員登録、トークン更新）。アクセス元の IP は `internal/middleware/client_key.go` の `ClientKey` で次のように決まる。

1. `X-Internal-Secret` が `INTERNAL_API_SECRET` と一致する → `X-Real-IP` の IP を使う（フロントのサーバーからの呼び出し）
2. 一致しない → Gin の `c.ClientIP()` を使う（API を直接呼んだ場合）

2 では、接続相手が Caddy になる。`TRUSTED_PROXIES` に Caddy のいるネットワークの範囲を登録すると、Caddy が付ける `X-Forwarded-For` から本当の IP を取る。この範囲を固定するため、compose で `networks.default` のサブネットを `172.30.0.0/24` に固定している。

## 運用の手順

### SSH で接続する

```bash
ssh -i ~/.ssh/movie-log-api.pem ec2-user@<Elastic IP>
cd ~/movie-log-api
```

### 新しいバージョンをデプロイする

PC で行う。

```bash
git tag vX.Y.Z
./scripts/build-image.sh                     # movie-log-api:vX.Y.Z を作る（未コミットの変更があると失敗する）
docker save movie-log-api:vX.Y.Z | gzip > movie-log-api-vX.Y.Z.tar.gz
scp -i ~/.ssh/movie-log-api.pem movie-log-api-vX.Y.Z.tar.gz ec2-user@<Elastic IP>:~/
```

サーバーで行う。

```bash
gunzip -c ~/movie-log-api-vX.Y.Z.tar.gz | docker load
cd ~/movie-log-api
vim .env                                     # APP_VERSION=vX.Y.Z に変更
docker compose up -d
rm ~/movie-log-api-vX.Y.Z.tar.gz
```

PC の `movie-log-api-vX.Y.Z.tar.gz` も削除する。

### compose.yaml や Caddyfile を変更する

リポジトリで変更してコミットし、サーバーに送ってから起動し直す。

```bash
scp -i ~/.ssh/movie-log-api.pem docker-compose.prod.yml ec2-user@<Elastic IP>:~/movie-log-api/compose.yaml
scp -i ~/.ssh/movie-log-api.pem Caddyfile ec2-user@<Elastic IP>:~/movie-log-api/
```

```bash
docker compose up -d
```

ネットワークの設定を変えた場合は、`docker compose down` してから `docker compose up -d` する。`down` に `-v` を付けると DB のデータが消えるので付けない。

### 状態を確認する

```bash
docker compose ps                      # 3つのコンテナが Up か
docker compose logs api | tail -n 20   # API のログ
docker compose logs caddy | tail -n 20 # Caddy のログ
curl -i https://api.cinelier.app/      # PC から。404 と {"code":"NOT_FOUND",...} が返れば正常
```

## バックアップ

- 毎日 3:00（日本時間）に `movie-log-backup.timer` が `scripts/backup-db.sh` を実行する。
- `pg_dump -Fc` で DB を書き出し、`movie_log-<UTC日時>.dump` という名前でバックアップバケットに送る。
- 30日を過ぎたファイルは S3 が自動で削除する。

### 確認する

```bash
systemctl list-timers movie-log-backup.timer          # 次回と前回の実行時刻
journalctl -u movie-log-backup.service -n 20 --no-pager # 実行時のログ
```

手動で実行する場合:

```bash
sudo systemctl start movie-log-backup.service
```

### 復元する

サーバーにはバックアップを読む権限がないので、PC 経由で送る。

1. AWS コンソールの S3 でバックアップバケットを開き、`.dump` ファイルをダウンロードする。
2. サーバーに送る。

   ```bash
   scp -i ~/.ssh/movie-log-api.pem ~/Downloads/movie_log-XXXX.dump ec2-user@<Elastic IP>:~/restore.dump
   ```

3. 試験用の DB に戻して中身を確認する場合:

   ```bash
   cd ~/movie-log-api
   docker compose exec -T postgres sh -c 'createdb -U "$POSTGRES_USER" restore_test'
   docker compose exec -T postgres sh -c 'pg_restore -U "$POSTGRES_USER" -d restore_test' < ~/restore.dump
   docker compose exec postgres sh -c 'psql -U "$POSTGRES_USER" -d restore_test -c "\dt"'
   docker compose exec -T postgres sh -c 'dropdb -U "$POSTGRES_USER" restore_test'
   ```

4. 本番の DB に戻す場合は、API を止めてから既存のデータを置き換える。

   ```bash
   docker compose stop api
   docker compose exec -T postgres sh -c 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists' < ~/restore.dump
   docker compose start api
   ```

5. サーバーと PC の `.dump` ファイルを削除する。

## トラブルのときに確認すること

- **API が DB に接続できない**: postgres 側のログの `DETAIL` 行に原因が出る。

  ```bash
  docker compose logs postgres | grep -A1 FATAL
  ```

- **`.env` を直したのに反映されない**: compose は、シェルの環境変数を `.env` より優先する。シェルで `.env` を読み込んだ（`set -a; . ./.env` など）場合は、SSH を切って入り直す。
- **`POSTGRES_USER` などを変えても反映されない**: これらは DB の初回作成時にしか使われない。反映するには `docker compose down -v` で DB を作り直す（データが消える）。
- **SSH が応答しない**: セキュリティグループの SSH ルールが、今の自宅の IPv4 アドレスになっているか確認する。

## 未完了

- OS の自動更新（Amazon Linux 2023 は OS のバージョンが固定されているため、`/etc/dnf/vars/releasever` に `latest` を書き、`dnf-automatic` の `dnf-automatic-install.timer` を有効にする予定）
- カーネルの更新後の再起動をどうするか（手動で定期的に `sudo reboot`、または dnf-automatic の自動再起動）
- Web アプリのデプロイ時に、同じ `INTERNAL_API_SECRET` を渡す

## 作業記録

作業ごとの記録は `logs/` にある。ファイル名は `YYYY-MM-DD-内容.md`。
