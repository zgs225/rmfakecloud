The configuration is made through environment variables.

## General configuration

| Variable name     | Description |
|-------------------|-------------|
| `JWT_SECRET_KEY`  | The secret key used to sign the authentication token.<br>If you don't provide it, a random secret is generated, invalidating all connections established previously to be closed.<br>A good secret is for example: `openssl rand -base64 48` |
| `STORAGE_URL`     | It controls whether file upload/download goes through the local proxy or to an external server. It's the full address (protocol, host, port, path) of rmfakecloud **as visible from the tablet**, especially if the host is behind a reverse proxy or in a container. Example: `http://192.168.2.3:3000` (default: `https://local.appspot.com`) |
| `PORT`            | listening port number (default: 3000) |
| `DATADIR`         | Set data/files directory (default: `data/` in current dir) |
| `LOGLEVEL`        | Set the log verbosity. Default is **info**, set to **debug** for more logging or **warn**, **error** for less |
| `RM_HTTPS_COOKIE` | For the UI, force cookies to be available only via https |
| `RM_TRUST_PROXY`  | Trust the proxy for client ip addresses (X-Forwarded-For/X-Real-IP) default false |

## Handwriting recognition

To use the handwriting recognition feature, you need first to create a free account on <https://developer.myscript.com/> (up to 2000 free recognitions per month).

Then you'll obtains an application key and its corresponding HMAC to give to rmfakecloud:

| Variable name              | Description |
|----------------------------|-------------|
| `RMAPI_HWR_APPLICATIONKEY` | Application key obtained from myscript |
| `RMAPI_HWR_HMAC`           | HMAC obtained from myscript |

## Email settings

To be able to send email from your reMarkable, fill the following variables:

| Variable name          | Description |
|------------------------|-------------|
| `RM_SMTP_SERVER`       | The SMTP server address in  host:port format |
| `RM_SMTP_USERNAME`     | The username/email for login |
| `RM_SMTP_PASSWORD`     | Plaintext password (application password should work) |
| `RM_SMTP_FROM`         | Custom `From:` header for the mails (eg. `ReMarkable self-hosted <remarkable@my.example.net>`). If this override is set, the user's email address is instead put as `Reply-To` |
| `RM_SMTP_HELO`         | Custom HELO, if your email provider needs it |
| `RM_SMTP_NOTLS` | don't use tls |
| `RM_SMTP_STARTTLS` | use starttls command, should be combined with NOTLS. in most cases port 587 should be used |
| `RM_SMTP_INSECURE_TLS` | If set, don't check the server certificate (not recommended) |
