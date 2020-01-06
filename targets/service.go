package targets

import (
	"encoding/pem"
	"io"
	"io/ioutil"
	"os/exec"
)

func listNotaryTargets() ([]Target, error) {
	cmd := exec.Command("notary", "key", "export")
	pipeReader, pipeWriter := io.Pipe()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	targets := make([]Target, 0)
	done := make(chan bool, 1)
	go func() {
		defer close(done)
		data, err := ioutil.ReadAll(pipeReader)
		if err != nil {
			return
		}

		decodeRecurse(&targets, data)
	}()

	err := cmd.Run()
	if err != nil {
		pipeWriter.CloseWithError(err)
		return nil, err
	}
	pipeWriter.Close()

	<-done
	return targets, nil
}

func decodeRecurse(targets *[]Target, data []byte) {
	if len(data) == 0 {
		return
	}
	block, rest := pem.Decode(data)
	if block != nil {
		if role, ok := block.Headers["role"]; ok && role == "targets" {
			if gun, ok := block.Headers["gun"]; ok {
				if path, ok := block.Headers["path"]; ok {
					*targets = append(*targets, Target{path, gun})
				}
			}
		}
	}
	decodeRecurse(targets, rest)
}
