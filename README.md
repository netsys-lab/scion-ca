# scion-ca
Repo for smallstep CA to run in the SCION Education Network, i.e. in the GEANT Core AS

## Setup
The CA Service can be configured in the .toml file of the control service in a SCION AS. There is a mode called "delegating" which makes the CS talk to an external CA Service via HTTP implementing [this interface](https://github.com/scionproto/scion/blob/master/spec/ca.gen.yml).

![geant-scion-ca](https://user-images.githubusercontent.com/32448709/185569635-a538f8bc-965a-4a6c-a166-1673f2a66b0f.jpg)
