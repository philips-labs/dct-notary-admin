package targets

import (
	"context"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os/exec"
)

func streamNotaryTargets(ctx context.Context) (<-chan Target, <-chan error, error) {
	cmd := exec.Command("notary", "key", "export")
	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter
	targetChan := make(chan Target)
	errorChan := make(chan error, 1)

	go processPrivateKeys(ctx, pipeReader, targetChan, errorChan)

	return targetChan, errorChan, cmd.Run()
}

func processPrivateKeys(ctx context.Context, reader io.ReadCloser, stream chan<- Target, errChan chan<- error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		errChan <- err
		return
	}

	go func() {
		defer close(stream)
		defer reader.Close()
		for {
			block, rest := pem.Decode(data)
			if len(data) == 0 {
				return
			}
			if block != nil && isTarget(block) {
				path, gun := getPathAndGun(block)
				if path != "" && gun != "" {
					select {
					case stream <- Target{Path: path, Gun: gun}:
					case <-ctx.Done():
						return
					}
				}
			}
			data = rest
		}
	}()
}

func isTarget(block *pem.Block) bool {
	role, ok := block.Headers["role"]
	return ok && role == "targets"
}

func getPathAndGun(block *pem.Block) (string, string) {
	return block.Headers["path"], block.Headers["gun"]
}

func listNotaryTargets(ctx context.Context) ([]Target, error) {
	targetChan, errChan, err := streamNotaryTargets(ctx)
	if err != nil {
		return nil, err
	}

	targets, err := getTargets(targetChan, errChan)

	return targets, err
}

func getTargets(targetChan <-chan Target, errChan <-chan error) ([]Target, error) {
	targets := make([]Target, 0)
	for {
		select {
		case target, open := <-targetChan:
			if !open {
				return targets, nil
			}
			targets = append(targets, target)
		case err := <-errChan:
			return nil, err
		}
	}
}
