# !/bin/bash
export $(cat ../.env | sed 's/#.*//g' | xargs)
set -e

# AS
# ISD
# ISD-AS from ../.env

NOW=$(date +%s)
NAME="Sample ISD"

replaceVars () {
  FILE=$1
  # echo "replacing in file ${FILE}"
  sed -i "s/{{ISD}}/${ISD}/g" $FILE
  sed -i "s/{{AS}}/${AS}/g" $FILE
  sed -i "s/{{ISD_AS}}/${ISD_AS}/g" $FILE
  sed -i "s/{{NAME}}/${NAME}/g" $FILE
  sed -i "s/{{NOW}}/${NOW}/g" $FILE 
}

echo "Generating new certs and TRC for ISD ${ISD} and core AS ${AS}"

echo "Generate certificates and keys"
echo "Generating root key pair..."
cp cp-root.tmpl "${ISD_AS}.root.tmpl"
replaceVars "${ISD_AS}.root.tmpl"
scion-pki certificate create "${ISD_AS}.root.tmpl" ISD$ISD_AS.root.crt ISD$ISD_AS.root.key  --profile=cp-root "--not-before=2022-07-08T07:20:50.52Z" "--not-after=2027-07-08T07:20:50.52Z"
echo "Generating root key pair... done."

echo "Generating sensitive-voting key pair..."
cp sensitive.tmpl "${ISD_AS}.sensitive.tmpl"
replaceVars "${ISD_AS}.sensitive.tmpl"
scion-pki certificate create "${ISD_AS}.sensitive.tmpl" ISD$ISD_AS.sensitive.crt ISD$ISD_AS.sensitive.key  --profile=sensitive-voting "--not-before=2022-07-08T07:20:50.52Z" "--not-after=2027-07-08T07:20:50.52Z"
echo "Generating sensitive-voting key pair... done."

echo "Generating regular-voting key pair..."
cp regular.tmpl "${ISD_AS}.regular.tmpl"
replaceVars "${ISD_AS}.regular.tmpl"
scion-pki certificate create "${ISD_AS}.regular.tmpl" ISD$ISD_AS.regular.crt ISD$ISD_AS.regular.key  --profile=regular-voting "--not-before=2022-07-08T07:20:50.52Z" "--not-after=2027-07-08T07:20:50.52Z"
echo "Generating regular-voting key pair... done."

echo "Generating TRC..."
cp trc.toml "${ISD_AS}-trc.toml"
replaceVars "${ISD_AS}-trc.toml"

# Create TRC Payload
scion-pki trcs payload --template "${ISD_AS}-trc.toml" --out ISD$ISD-B1-S1.pld.der

# Sign TRC Payload
scion-pki trc sign ISD$ISD-B1-S1.pld.der ISD$ISD_AS.sensitive.crt ISD$ISD_AS.sensitive.key --out ISD$ISD-B1-S1.sensitive.trc
scion-pki trc sign ISD$ISD-B1-S1.pld.der ISD$ISD_AS.regular.crt ISD$ISD_AS.regular.key --out ISD$ISD-B1-S1.regular.trc

# Combine TRC Payload
scion-pki trc combine --payload ISD$ISD-B1-S1.pld.der --format pem -o ISD$ISD-B1-S1.trc ISD$ISD-B1-S1.sensitive.trc ISD$ISD-B1-S1.regular.trc
echo "Generating TRC... done."