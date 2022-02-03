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
