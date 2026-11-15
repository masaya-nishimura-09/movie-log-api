# 2026-11-15 ログを30日で削除する設定

<!-- markdownlint-disable MD013 -->

本番のコンテナのログが削除されずに残り続けていたため、30日で削除するようにした（issue #31）。

## 決めたこと

- **消す方法**: 容量ではなく日数で消す。プライバシーポリシーの「一定の期間を過ぎると削除します」に合わせるため。容量で消す方法（json-file の `max-size` / `max-file`）では、アクセスが少ないとログがいつまでも残る。
- **期間**: 30日。障害の調査には足りる。ポリシーの「一定の期間」も「30日」に書き換える。
- **仕組み**: コンテナのログを journald に保存し、journald の保存期間で消す。json-file には日数で消す設定がない。
- **対象**: api、postgres、caddy の3つ。api のログには IP アドレスと `userId` が出ている。

## 行った作業

1. `/etc/docker/daemon.json` がないことを確認した（Docker の初期設定の json-file で、削除なし）。
2. `/var/log/journal/` があることを確認した（journald はログをディスクに保存している）。
3. `/etc/systemd/journald.conf.d/retention.conf` を作り、`MaxRetentionSec=30day` と `MaxFileSec=1day` を書いた。
4. `sudo systemctl restart systemd-journald` で反映した。
5. `docker-compose.prod.yml` の3つのサービスに `logging.driver: journald` を加えた。
6. サーバーに送り、`docker compose up -d` でコンテナを作り直した。json-file に残っていた古いログは、コンテナと一緒に消えた。

## 確認したこと

- `systemctl status systemd-journald` が `active (running)` だった。
- `systemd-analyze cat-config systemd/journald.conf` の最後に `retention.conf` の内容が出た。
- `docker compose ps` で3つのコンテナが `Up` だった。
- `docker inspect -f '{{.HostConfig.LogConfig.Type}}'` で、3つとも `journald` だった。

## 起きた問題

### 保存期間が30秒になっていた

最初は `MaxRetentionSec=30` と単位を付けずに書いていた。単位がないと秒として扱われる。反映する前に気づき、`30day` に直した。

## 残した作業

- Web（movie-log）のプライバシーポリシーの「一定の期間」を「30日」に書き換える。
