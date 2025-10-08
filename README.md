### Register Ilugc

Project to gather registration details.

#### Run
1. Clone this repo
2. From the Cloned directory, run following command, this will build `src/cmd/register-ilugc` binary
```console
make
```
3. Run following command to start this service.  This service listen by default on port `2203`.
```console
src/cmd/register-ilugc
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