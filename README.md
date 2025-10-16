### Register Ilugc

Service to gather registration details.

#### Requirements

1. GNU Make
2. Golang
3. Podman-Compose (optional)

Use the following command to install requirement in Ubuntu
```console
sudo apt install -y make golang-go podman-compose
```

#### Run

1. Clone this repo
2. From the Cloned directory, run following command, this will build `src/cmd/register` binary
```console
make
```
3. Run following command to start this service.  This service listen by default on port `2203`.
```console
src/cmd/register
```

#### Podman Run

from `container` directory, run
```console
podman-compose up
```

#### Accessing Webservice

Using the following command we can check if this service is accessable
```console
curl -v http://localhost:2203
```

### Endpoints

- `/` :: Primary Registration Form
- `/config` :: Show form to dynamically change maximum participants limit or to stop registration (Auth required)
- `/csv` :: To get the current participants list as csv (Auth required)

### Config

You can create `default.config` file in the current directory where you run `register` binary to configure it.
```json
{
    "hostport": ":2203",
    "domain": "https://register.ilugc.in",
    "static": "",
    "defaultmax": 0,
    "stopregistration": false,
    "adminusername": "",
    "adminpassword": ""
}
```

- `domain` :: This field is used for generating `qrcode`.
- `hostport` :: `host` and `port` information to listen.
- `static` :: where the static files are available.
- `defaultmax` :: if zero, no limitation, otherwise only allow participants upto this number.
- `stopregistration` :: simply stop registration.
- `adminusername` :: Administrator Username (used for `Auth Required` endpoints)
- `adminpassword` :: Administrator Password (used for `Auth Required` endpoints). It **should** be **base64 encoded** string. You can easily generate it using following commandline in bash shell
```console
echo -n "password" | base64
```
