# !/bin/sh

if [ -d "/etc/step-ca/.step" ]; then
    echo "/etc/step-ca/.step exists, SCION CA was already initialized"
else
echo "Creating new SCION step-ca"
rm -rf /root/.step
step ca init --key=$CA_ROOT_KEY --root=$CA_ROOT_CERT --provisioner=$SCION_CA_PROVISIONER --deployment-type=standalone --name=$SCION_CA_SUBJECT_COMMONNAME --dns=127.0.0.1 --address=127.0.0.1:8443 --password-file /etc/step-ca/scion-ca.pw
cp /bin/step /etc/step-ca/step
cp -R /root/.step/ /etc/step-ca/
fi

if [ -d "/etc/step-internal/.step" ]; then
    echo "/etc/step-internal/.step exists, CA was already initialized"
else
echo "Creating new step-ca"
rm -rf /root/.step
step ca init --provisioner=$STEP_CA_PROVISIONER --deployment-type=standalone --name=$STEP_CA_SUBJECT_COMMONNAME --dns=127.0.0.1 --address=127.0.0.1:9443 --password-file /etc/step-internal/step-ca.pw
cp -R /root/.step/ /etc/step-internal/
fi

sleep infinity

# step ca certificate --kty=RSA --not-after=24h --provisioner-password-file=/etc/step-ca/scion-ca.pw --ca-url=https://127.0.0.1:8443 --root=/etc/step-ca/.step/certs/root_ca.crt "1-999 AS Certificate" /etc/step-ca/tmp.crt /etc/step-ca/tmp.key