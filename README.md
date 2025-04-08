<div style="text-align: center">

[![Flatcar OS](https://img.shields.io/badge/Flatcar-Website-blue?logo=data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4NCjwhLS0gR2VuZXJhdG9yOiBBZG9iZSBJbGx1c3RyYXRvciAyNi4wLjMsIFNWRyBFeHBvcnQgUGx1Zy1JbiAuIFNWRyBWZXJzaW9uOiA2LjAwIEJ1aWxkIDApICAtLT4NCjxzdmcgdmVyc2lvbj0iMS4wIiBpZD0ia2F0bWFuXzEiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiIHg9IjBweCIgeT0iMHB4Ig0KCSB2aWV3Qm94PSIwIDAgODAwIDYwMCIgc3R5bGU9ImVuYWJsZS1iYWNrZ3JvdW5kOm5ldyAwIDAgODAwIDYwMDsiIHhtbDpzcGFjZT0icHJlc2VydmUiPg0KPHN0eWxlIHR5cGU9InRleHQvY3NzIj4NCgkuc3Qwe2ZpbGw6IzA5QkFDODt9DQo8L3N0eWxlPg0KPHBhdGggY2xhc3M9InN0MCIgZD0iTTQ0MCwxODIuOGgtMTUuOXYxNS45SDQ0MFYxODIuOHoiLz4NCjxwYXRoIGNsYXNzPSJzdDAiIGQ9Ik00MDAuNSwzMTcuOWgtMzEuOXYxNS45aDMxLjlWMzE3Ljl6Ii8+DQo8cGF0aCBjbGFzcz0ic3QwIiBkPSJNNTQzLjgsMzE3LjlINTEydjE1LjloMzEuOVYzMTcuOXoiLz4NCjxwYXRoIGNsYXNzPSJzdDAiIGQ9Ik02NTUuMiw0MjAuOXYtOTUuNGgtMTUuOXY5NS40aC0xNS45VjI2MmgtMzEuOVYxMzQuOEgyMDkuNFYyNjJoLTMxLjl2MTU5aC0xNS45di05NS40aC0xNnY5NS40aC0xNS45djMxLjINCgloMzEuOXYxNS44aDQ3Ljh2LTE1LjhoMTUuOXYxNS44SDI3M3YtMTUuOGgyNTQuOHYxNS44aDQ3Ljh2LTE1LjhoMTUuOXYxNS44aDQ3Ljh2LTE1LjhoMzEuOXYtMzEuMkg2NTUuMnogTTQ4Ny44LDE1MWg3OS42djMxLjgNCgloLTIzLjZ2NjMuNkg1MTJ2LTYzLjZoLTI0LjJMNDg3LjgsMTUxTDQ4Ny44LDE1MXogTTIzMywyMTQuNlYxNTFoNjMuN3YyMy41aC0zMS45djE1LjhoMzEuOXYyNC4yaC0zMS45djMxLjhIMjMzVjIxNC42eiBNMzA1LDMxNy45DQoJdjE1LjhoLTQ3Ljh2MzEuOEgzMDV2NDcuN2gtOTUuNVYyODYuMUgzMDVMMzA1LDMxNy45eiBNMzEyLjYsMjQ2LjRWMTUxaDMxLjl2NjMuNmgzMS45djMxLjhMMzEyLjYsMjQ2LjRMMzEyLjYsMjQ2LjRMMzEyLjYsMjQ2LjR6DQoJIE00NDguMywzMTcuOXY5NS40aC00Ny44di00Ny43aC0zMS45djQ3LjdoLTQ3LjhWMzAyaDE1Ljl2LTE1LjhoOTUuNVYzMDJoMTUuOUw0NDguMywzMTcuOXogTTQ0MCwyNDYuNHYtMzEuOGgtMTUuOXYzMS44aC0zMS45DQoJdi03OS41aDE1Ljl2LTE1LjhoNDcuOHYxNS44aDE1Ljl2NzkuNUg0NDB6IE01OTEuNiwzMTcuOXY0Ny43aC0xNS45djE1LjhoMTUuOXYzMS44aC00Ny44di0zMS43SDUyOHYtMTUuOGgtMTUuOXY0Ny43aC00Ny44VjI4Ni4xDQoJaDEyNy4zVjMxNy45eiIvPg0KPC9zdmc+DQo=)](https://www.flatcar.org/)
[![Matrix](https://img.shields.io/badge/Matrix-Chat%20with%20us!-green?logo=matrix)](https://app.element.io/#/room/#flatcar:matrix.org)
[![Slack](https://img.shields.io/badge/Slack-Chat%20with%20us!-4A154B?logo=slack)](https://kubernetes.slack.com/archives/C03GQ8B5XNJ)
[![Twitter Follow](https://img.shields.io/twitter/follow/flatcar?style=social)](https://x.com/flatcar)
[![Mastodon Follow](https://img.shields.io/badge/Mastodon-Follow-6364FF?logo=mastodon)](https://hachyderm.io/@flatcar)
[![Bluesky](https://img.shields.io/badge/Bluesky-Follow-0285FF?logo=bluesky)](https://bsky.app/profile/flatcar.org)

</div>

# lerobot

**Status: experimental**

A simple robot managing Let's Encrypt certificates.

The current version is very limited and only allows DNS verification
via Route53.

The following credentials are expected in environment variables:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `AWS_HOSTED_ZONE_ID`
* `AWS_REGION`

That's by [lego]'s design, which lerobot uses for the ACME part for the
time being.

## Setup

Create a separate `lerobot` user with its own home directory, for example:

```shell
sudo adduser --system --disabled-password --home /home/lerobot --shell /bin/bash --gecos '' --group lerobot
sudo chmod 0700 /home/lerobot
```

Add an environment file `/etc/lerobot-env` with AWS credentials:

```shell
AWS_ACCESS_KEY_ID=ABCD
AWS_SECRET_ACCESS_KEY=1234
AWS_HOSTED_ZONE_ID=ZXXXL
AWS_REGION=eu-central-1
```

Add a `lerobot.service` systemd unit:

```shell
cat <<LEROBOT_SERVICE | sudo tee /etc/systemd/system/lerobot.service
[Unit]
Description=lerobot
After=network-online.target
Wants=network-online.target systemd-networkd-wait-online.service

[Service]
Restart=on-failure

User=lerobot
Group=lerobot

EnvironmentFile=/etc/lerobot-env

WorkingDirectory=%h

ExecStart=/usr/local/bin/lerobot daemon --authorized-keys-file /home/lerobot/.ssh/authorized_keys
ExecReload=/bin/kill -USR1 $MAINPID

TimeoutStopSec=60

PrivateTmp=true
PrivateDevices=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
LEROBOT_SERVICE

sudo systemctl daemon-reload
```

Create a file `lets-encrypt.yaml` in the home directory of the lerobot
user (here `/home/lerobot`) with the following structure:

```yaml
accounts:
  - email: infra@example.com
    ssh_public_key: |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7GEdUWkBr3zRnb53IKR+KBJRrJ4qxscUe6ZuThcNSnqMQZQWoUPc4fjKMRRoTMqDmH6pA+1N6IduXGa79VodnP+u5QRD7bP5wUFO8J2tY1Oer1wyEnXgOC7kL48KWzLuGdr1M3xEWZVYLgnrbT8HkWj09w01DRXhBaIXeDRpif/XQx5F5XuOlspcQq/9bEqiD0lFg8BFCW9oFYr+h8juyqh2/X0D3eL3nRfXAVGxn3HjfgGm4494fvGimeb4qF8kWKoO7QVndbZDSC2a4BkMhf1prS2EmIs3B3sOfFEYC9M9tkBfV3IIaUtuvXxw4fpRg48qCMGqvlvtl98fWmWPeWoMeDEpffgwGrBJm2yUYub/y7hF9X7dBuKj8ouxoOmN8ZkPMoSE9tNTkWDb1eQBQVZoxxZ4222KLWjgPWNNKNdulRCOB8Meop9/w6YkkXukvKrMHw8k+d5Z2owqzmbiAwfm8xuOeFS2y4twD0Z1R+OsCrHoOdgIZFV3H2jaCEVM= infra@example.com

certificates:
  - account: infra@example.com
    common_name: foo.example.com
    preferred_chain: ISRG Root X1
    subject_alternative_names:
      - bar.example.com
      - baz.example.com
```

**Note**: account data must be unique, i.e. do not duplicate `email`
or `ssh_public_key`.


All certificates for an account are stored in the same directory and
available to the same SSH user. I.e. to separate access to certificates,
use different Let's Encrypt users and different SSH keypairs.

Finally, enable and start `lerobot.service`:

```shell
sudo systemctl enable lerobot
sudo systemctl start lerobot
```

## Testing

When testing, make sure to **not** use the Let's Encrypt production API but
staging (can be set with `--le-url`):

```shell
https://acme-staging-v02.api.letsencrypt.org/directory
```

Example `lerobot daemon` invocation for testing:

```shell
./bin/lerobot daemon --le-api https://acme-staging-v02.api.letsencrypt.org/directory --le-config lets-encrypt.yaml
```

Example `lets-encrypt.yaml' file for testing:

```yaml
accounts:
  - email: infra@example.com
certificates:
  - account: infra@example.com
    common_name: "*.example.com"
    preferred_chain: ""
    subject_alternative_names: []
```

## Certificate consumers

Users are allowed to rsync all certificates for their account to their
machines. rsync is the only allowed command. In the default configuration,
the remote source path must be set exactly like shown below, i.e.
`certificates/<email>/`. It's not possible to use a different path or
to only sync a particular file.

Example:

```shell
rsync -ave "ssh -i /etc/lerobot.pem" lerobot@example.com:certificates/infra@example.com/ /etc/certificates/
```

This can be put into a systemd service triggered by a systemd timer once
per day.

[lego]: https://github.com/go-acme/lego
