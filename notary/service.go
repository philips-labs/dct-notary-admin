package notary

import (
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Key holds Path and GUN to keys
type Key struct {
	Path string `json:"id"`
	Gun  string `json:"gun"`
}

// Service notary service exposes notary operations
type Service struct {
	configFile string
}

// NewService creates a new notary service object
func NewService(configFile string) *Service {
	return &Service{configFile}
}

// GetTarget retrieves a target by its path/id
func (s *Service) GetTarget(ctx context.Context, path string) (*Key, error) {
	if len(path) < 7 {
		return nil, fmt.Errorf("you must provide at least 7 characters of the path: %w", ErrInvalidID)
	}

	targets, err := s.ListTargets(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range targets {
		if strings.HasPrefix(t.Path, path) {
			return &t, nil
		}
	}

	return nil, ErrItemNotFound(path)
}

// ListTargets lists all the notary target keys
func (s *Service) ListTargets(ctx context.Context) ([]Key, error) {
	targetChan, errChan, err := s.StreamTargets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run notary command: %w", err)
	}

	targets, err := getTargets(targetChan, errChan)

	return targets, err
}

// StreamTargets streams all the notary target keys on a channel
func (s *Service) StreamTargets(ctx context.Context) (<-chan Key, <-chan error, error) {
	cmd := exec.Command("notary", "-c", s.configFile, "key", "export")
	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter
	targetChan := make(chan Key)
	errorChan := make(chan error, 1)

	go processPrivateKeys(ctx, pipeReader, targetChan, errorChan)

	return targetChan, errorChan, cmd.Run()
}

func processPrivateKeys(ctx context.Context, reader io.ReadCloser, stream chan<- Key, errChan chan<- error) {
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
					case stream <- Key{Path: path, Gun: gun}:
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

func getTargets(targetChan <-chan Key, errChan <-chan error) ([]Key, error) {
	targets := make([]Key, 0)
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
