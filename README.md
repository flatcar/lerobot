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

```
sudo adduser --system --disabled-password --home /home/lerobot --shell /bin/bash --gecos '' --group lerobot
sudo chmod 0700 /home/lerobot
```

Add an environment file `/etc/lerobot-env` with AWS credentials:

```
AWS_ACCESS_KEY_ID=ABCD
AWS_SECRET_ACCESS_KEY=1234
AWS_HOSTED_ZONE_ID=ZXXXL
AWS_REGION=eu-central-1
```

Add a `lerobot.service` systemd unit:

```
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

```
accounts:
  - email: michael+letsencrypt-testing@kinvolk.io
    ssh_public_key: |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDdGAEc1/vnNEGHhIfCXvX4afywNzXep9vbHQLNGnYw/0cHe5MH8wi0srkyKdisXsKmIcNv5mCrV8Y8KAc7r775gJtRzMjY8HuxFZuVqk8mUbol56qGiQbLnF4YFUURv1qPebmXmkaU6u9S1usMlD28iSPO+U0acY2f2LXm5oKqclQz7rGUC7kjkw6A8c//N9ZvyKSuX6+7mfED9bW6bRtS7Dv7/q9J1RWwWckyDR9NXyaFnUKaBexdwmGxe5wBspAsYYeg7M7clNMiolybNw5Aqj0lchywux5LgbTgR5zqUF3B4pQ1mY/oi21VBWw+i9uU+4qJuAu9OmQd8yYLcjBh81zzN60Jr0OqHwq45i14Vgz8oY97oKkNzxhP4Py/PHDhkn3lRJeJRipAWhms+wvrRcuPJ9q87EVNXheuxmY3KYRhvBmaTj2A8XLKndnUaFAh82Kw5kcq0C4XiQruWv0HqjH2kNAaXQKaVAMgOUO0wRutsklnwFc5l79jmfxldL69egCD4EiToKk5oFDtQHPfo6bTkqg/O+5pkMoymGVBfw5XEjw6d4dvH+JMoJ1WNnLni7bP9jeeIOrPTWMlaxQ2fFyJ88TnPrGsrrPxcO6uliDTtpstdH8LgbdXNV0CBa4nXT7RjXCzuvkWTDu69ckUHd2QgHBnXSDYGnbwduo0zxepNbf9B+IBQoXbhM/UQenHxo1SdwRyJdq7mUpSRiIk4HNH3OrUmlszjM0jXje6UuIWewTs+vl8Ehm3S493pzC3Td1n1DL8uuz/4937Kx8u7iNLEqzbazi2k03eL+w+MXjNu3JstrjlPwNbhBVMKbIQJpJQwchmFO046jP4AqAdTcZxQarDofofkwXk4XbXAMgy/hFRp+UcDbu+xHpOMP9WJsPxl+9F8whMjCxEoTzDM9BEAJoqbHDlA4JD7ST828+Sbplzp4unHUlO/fV2kkSwjV+6ok7N3xiLThr3PvzFcEy1eGTuwxQWF/9B6dh4eoDzusqDOKXZEEb8BUJMvfhxCmV6sEEKHlVe4H4k9Klh73ftZGXR4sMPxsAYU8Y0+UhcI6lOs/xJ0rrVIFFJdc7dL5L5zpqt9VAYoKg/+smlkP6BNTFgRruJSgY21taK1yS94jYCrw5fNzDD4W05XLRO0DIaWTIPFJxiWQsN4/PwD5712x+OLVFRNCsiD9Dlk+qyLYYM4pxe/LM1EganUDaHCiPvHAjCUdNbIRXSq+KD+eY/YrLBYi9ALKPiuTgkQ/Y+SAg/iWDONR6LqYPv/t7hBWiLW5BCQ2p1XEIuIUzFvRoAHxarbb6ntxidV2orKqqGTfo4pOzPplj5wPFXa6xe7aKtP2dtlpDWW+h9W/nn michael@kinvolk.io

certificates:
  - account: michael+letsencrypt-testing@kinvolk.io
    common_name: foo.example.com
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

```
sudo systemctl enable lerobot
sudo systemctl start lerobot
```

## Testing

When testing, make sure to **not** use the Let's Encrypt production API but
staging (can be set with `--le-url`):

```
https://acme-staging-v02.api.letsencrypt.org/directory
```

Example `lerobot daemon` invocation for testing:

```
./bin/lerobot daemon --le-api https://acme-staging-v02.api.letsencrypt.org/directory --le-config lets-encrypt.yaml
```

Example `lets-encrypt.yaml' file for testing:

```
accounts:
  - email: michael+lerobot-testing@kinvolk.io
certificates:
  - account: michael+lerobot-testing@kinvolk.io
    common_name: "*.example.com"
    subject_alternative_names: []
```

## Certificate consumers

Users are allowed to rsync all certificates for their account to their
machines. rsync is the only allowed command. In the default configuration,
the remote source path must be set exactly like shown below, i.e.
`certificates/<email>/`. It's not possible to use a different path or
to only sync a particular file.

Example:

```
rsync -ave "ssh -i /etc/lerobot.pem" lerobot@example.com:certificates/michael+letsencrypt-testing@kinvolk.io/ /etc/certificates/
```

This can be put into a systemd service triggered by a systemd timer once
per day.

[lego]: https://github.com/go-acme/lego
