package step

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type StepCliAdapter struct {
	provisionerPassword string
	rootCert            string
	caUrl               string
}

func NewStepCliAdapter() *StepCliAdapter {
	provisionerPassword := "/etc/step-ca/scion-ca.pw"
	provisionerPasswordEnv := os.Getenv("SCION_CA_PROVISIONER_PASSWORD")
	if provisionerPasswordEnv != "" {
		provisionerPassword = provisionerPasswordEnv
	}

	rootCert := "/etc/step-ca/.step/certs/root_ca.crt"
	rootCertEnv := os.Getenv("SCION_CA_ROOT_CERT")
	if rootCertEnv != "" {
		rootCert = rootCertEnv
	}

	caUrl := "https://127.0.0.1:8443"
	caUrlEnv := os.Getenv("SCION_CA_URL")
	if caUrlEnv != "" {
		caUrl = caUrlEnv
	}

	return &StepCliAdapter{
		provisionerPassword,
		rootCert,
		caUrl,
	}
}

func (sca *StepCliAdapter) SignCert(scrPath, outputPath, duration string) error {
	cmd := exec.Command("step", "ca", "sign", fmt.Sprintf("--provisioner-password-file=%s", sca.provisionerPassword), fmt.Sprintf("--not-after=%s", duration), scrPath, outputPath, fmt.Sprintf("--ca-url=%s", sca.caUrl), fmt.Sprintf("--root=%s", sca.rootCert))

	var out bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &out
	log.Infof("Executing: %s", cmd.String())
	err := cmd.Run()
	if err == nil {
		log.Infof("Execute successful")
	} else {
		log.Infof("Execute failed %s", err.Error())
		log.Error(stdErr.String())
		return err
	}
	return nil
}
