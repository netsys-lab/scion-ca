package step

import (
	"bytes"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type StepCliAdapter struct {
}

func NewStepCliAdapter() *StepCliAdapter {
	return &StepCliAdapter{}
}

func (sca *StepCliAdapter) SignCert(scrPath, outputPath, duration string) error {
	// step ca sign --not-after=1440h switch.csr switch-^Cw.crt
	//  "--ca-config=/home/scion/config/ca.json", "--root=/home/scion/.step/root_ca.crt", "--ca-url=https://127.0.0.1:443",
	cmd := exec.Command("step", "ca", "sign", "--provisioner-password-file=/home/scion/.step/pw.key", fmt.Sprintf("--not-after=%s", duration), scrPath, outputPath)
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
