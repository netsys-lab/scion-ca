# scion-ca
Repo for smallstep CA to run in the SCION Education Network, i.e. in the GEANT Core AS

## Setup
The CA Service can be configured in the .toml file of the control service in a SCION AS. There is a mode called "delegating" which makes the CS talk to an external CA Service via HTTP implementing [this interface](https://github.com/scionproto/scion/blob/master/spec/ca.gen.yml).

![geant-scion-ca](https://user-images.githubusercontent.com/32448709/185569635-a538f8bc-965a-4a6c-a166-1673f2a66b0f.jpg)

## Get Started
The initial setup consists of a few steps that need to be done before you can deploy your own CA.

### Create your own root certs and TRCs
**Note:** This step should be performed according to this documentation only for testing purposes. Please refer to Anapaya's official guide for TRC creation and signing ceremonies.

Create your root certs and TRCs via the following commands:
`cd ca-conf`
`./gen-trc.sh`
`cd ..`

### Create step-ca folder for all the CA configuration files
Now create your step-ca and step-internal folders which will serve as root directory for your scion CA.

`mkdir step-ca`
`mkdir step-internal`

Copy your root cert, key and your trc into step-ca. Root certs and keys are named the following `ca-conf/ISD1-999.root.crt/key` for ISD_AS 1-999. The TRC is `ISD1-B1-S1.trc` for this ISD-AS.

Now generate secrets that will be used for the CA and for communication between CS and CA:

Generate a jwt-secret by copying the jwt-secret.template from `ca-conf` to `step-ca` and insert a proper symmetric key via `openssl rand -base64 256` and name it `jwt-secret.pem`. It will be used by the scion-step-proxy to sign JWT tokens.

Generate a shared secret by copying the shared-secret.pem.template from `ca-conf` to `step-ca` and insert a proper symmetric key via `openssl rand -base64 256` and name it `shared-secret.pem`. It will be used by the CS to obtain JWT tokens from the scion-step-proxy.

Generate files that contain passwords for you SCION-ca in `step-ca/scion.pw` and for the internal CA `step-internal/step-ca.pw`.

Copy `seeds.json` from `ca-conf` into `step-ca` and add a random client id to your admin user.

### Run your CA to create the initial configuration

Start by running the `smallstep-cli-scion` container that creates the initial configuration: `docker-compose up -d smallstep-cli-scion`

Check the log output if everything was done properly: `docker-compose logs -f smallstep-cli-scion`. It should print something like `Your PKI is ready to go` two times.

### Add SCION specific step configuration
**Note:** I hope I can automate this step later, too.

You may need to change ownership of the `step-ca/.step` and `step-internal/.step` folders to your current user to edit files there.

`sudo chown -R $USER step-ca/.step`
`sudo chown -R $USER step-internal/.step`

Add the `leaf.tpl` from `ca-conf` to `step-ca/.step/templates`.

Make the scion-ca use the specific template by adding the following lines into the config file `step-ca/.step/config/ca.json` (starting at line 23 into the provisioners object):

```
"options": {
    "x509": {
        "templateFile": "/root/.step/templates/leaf.tpl"
    }
},
```

Now configure the TLS cert duration to be longe than 24h: Add the following lines after the provisioners array (starting at line 40):

```
"claims": {
    "minTLSCertDuration": "5m",
    "maxTLSCertDuration": "1440h",
    "defaultTLSCertDuration": "24h"
}
```

### Start SCION-CA 
Next step is to start your SCION CA by running `docker-compose up -d smallstep-ca-scion`. Again check the logs if everything is set up properly `docker-compose logs -f smallstep-cli-scion`.

### Start SCION Step Proxy
Next, you need to start the SCION step proxy via `docker-compose up -d scion-step-proxy`.

### Secure your SCION Step Proxy with Caddy
To protect communication between the CS and the SCION step proxy, we propose to configure caddy as reverse-proxy to obtain certificates from your internal step-ca.

Run your internal CA by via `docker-compose up -d step-ca`.

Adapt the IP you want Caddy to listen on via HTTPS in the Caddyfile and also change the TLS email setting.

Next, trust the root cert of your step-internal CA `step-internal/.step/certs/root_ca.crt` either by following this guide (for Ubuntu) or use step for this `sudo step-ca/step certificate install step-internal/.step/certs/root_ca.crt`. Your updated trust store will then be linked into the caddy container

Now add acme support to your internal step-ca `docker-compose exec step-ca step ca provisioner add acme --type ACME` and `docker-compose restart step-ca`

Now run `docker-compose up -d caddy`. It should log that it is capable of obtaining a new cert from your step-internal instance.

### Configuration of your CS
In the CS, there is a section [ca] that needs to be configured the following way:

```yaml
[ca]
mode = "delegating"

[ca.service]
shared_secret = "./step-ca/shared_key.pem" # Update to your location
addr = "https://127.0.0.1:443" # Point to caddys IP
client_id = "YOUR CLIENT ID"
```

### Create and renew certificates
Coming soon...