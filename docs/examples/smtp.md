# SMTP

Cassocial checks for a local SMTP relay on first startup. It probes the
loopback address, Docker bridge gateway, default gateway, detected host name,
global IPv4 address, and `mail.` and `smtp.` host names, using ports 25, 465,
and 587. The first successful EHLO connection is saved in the server
configuration. If no relay responds, email features stay disabled and the
server continues running.

Configure a relay under `server.notifications.email` in `server.yml`:

```yaml
server:
  notifications:
    email:
      smtp:
        host: ""
        port: 587
        username: ""
        password: ""
        tls: auto
      from:
        name: ""
        email: ""
```

An empty host enables autodetection. A configured host is tested at startup;
email is disabled with a warning if the connection fails. The password belongs
in `{config_dir}/secrets.yml`, not in `server.yml`. `SMTP_HOST`, `SMTP_PORT`,
`SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_TLS`, `SMTP_FROM_NAME`, and
`SMTP_FROM_EMAIL` override the corresponding settings for containerized setups.

With `tls: auto`, port 465 uses implicit TLS. Other ports begin with SMTP and
EHLO, then upgrade with STARTTLS if the relay offers it. Credentials are sent
only after TLS is established. The explicit modes are `starttls`, `tls`, and
`none`. For IP literals in private, loopback, link-local, or carrier-grade
NAT ranges, TLS encryption remains enabled while certificate verification is
relaxed to support local relays without IP subject alternative names.
