package notary

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	certificates = `-----BEGIN ENCRYPTED PRIVATE KEY-----
path: 2c788e511219cf611f23a3cc556183870db9d265501bc5b3f5e9a57af1466a9e
role: root

MIHuMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAifdVSXP80IGAICCAAw
HQYJYIZIAWUDBAEqBBCZk5uhb8e8IUTAqJbR1fY2BIGgmjH0KRfI9iSWGqIf+X1S
JUy/pXUxK4jKO9TOg54T0mTu0BjQCY3ABmvIeqUvDZxZaLLxwiMexhC2FJk6GDyj
1WhJ04IV4KFaclUS4VY8OyooLi4SyKsb1HpMP6kTaVj7kwsO5hZMHmY8bnWnEs+8
rgckk9vEPm65m9pWJLoa/bqrEoccOMLYyfX/+IDLpJyVIGvWNzHCfOlqitiVDCKq
mg==
-----END ENCRYPTED PRIVATE KEY-----
-----BEGIN ENCRYPTED PRIVATE KEY-----
gun: docker.io/marcofranssen/openjdk
path: b34192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185
role: targets

MIHuMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAh8fUZq36LiDQICCAAw
HQYJYIZIAWUDBAEqBBB8rNZ8uhj3e6D8Xbr5CQj4BIGgZSepWEmDuzVKdPj/HjB8
2jwCrUihkot8dkUymgshYGwhj4ggPZbNxYI33ojR9mZEADzzDVYIQTRJUO7zj2YB
/gUrq3lA3Tkjg/ahESnI040VLfunq0JEtAc/SncyL9NvMzxHopoYxLpGSJkrI1+D
DBM1/asre0iIgUrfp2hNoGLCVXMX0v8x1GtrTTtFFG/z6j4CFZB4EFWUQbm+LaAV
hw==
-----END ENCRYPTED PRIVATE KEY-----
-----BEGIN ENCRYPTED PRIVATE KEY-----
gun: docker.io/marcofranssen/whalesay
path: b624efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8
role: targets

MIHuMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAhtql75UhDqdAICCAAw
HQYJYIZIAWUDBAEqCCAyA4euyCW6int7fjCvTaG/BIGgycsOfRmkIGvYLctLKtJC
8SJ1Dm57RezMGuy1608R+p/KN1jvs/VvH/fECEYrQfGMQmu0hlOiGHgYnFt+3Ay1
SPKn03++ktzPMzhKEy7LweztRff8dQnN5TQw/jy3+dJYlLf1CaJUrY5RkzZO6/eq
RXH5T8L3Ekc619mrAHu7wQK17k4dkHxJbsooa13+c9HDB1Py+XZlXgv2sFBNOWg5
rw==
-----END ENCRYPTED PRIVATE KEY-----
-----BEGIN ENCRYPTED PRIVATE KEY-----
gun: docker.io/marcofranssen/nginx
path: d11b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd
role: targets

MIHuMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAjEKbGSqxggQAICCAAw
HQYJYIZIAWUDBAEqBBBb3/LPoK6R/g//5QfNECcHBIGgWaOJy1xm2wQWy0SN6tJo
eKdHhO6EdtRNWCyEoHlJ8rHIO+CXHmr06IAempZq7UG/CY0dDj0mIc3QYIyprHF9
W8i0hP4nUi+v9zn/JOofp7JquCBMfe2BywZJJy8KdkalUAPRfGInyuQVLLWz+x1d
FFSTGwg9ucwAlpZLvVoWAnWfIN6A8HtVL0GVzX1htvM5cGlhgeH7JMZbjpA6n6xi
RQ==
-----END ENCRYPTED PRIVATE KEY-----
-----BEGIN ENCRYPTED PRIVATE KEY-----
path: ea8dd99255f91efeba139941fbfdb629f11c2353704de07a2ad653d22311c88b
role: marcofranssen

MIHuMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAh4HLLi4scNHgICCAAw
HQYJYIZIAWUDBAEqBBC3WlfJhgltrOLmevCn8gcUBIGgFurC850B6kKqn2MoqtsR
83Wa2Aw+peGfQPAo77/JaRDgJSr/TqEigs4JxgcXFcF4+a1jM5Mv2EWBPBNoRs30
GCtUr/tisJZV2ZvDjP17ovv4Y1QGy7f1ezIQ8VA1ly6ozTav0FKuV6/R/KHV4dwq
IumYajGBaqNG7pZGDYcQXNCH/nmToiQ93a1PVPVLDMchCXCyvCuluyyT23KqdTUa
Fw==
-----END ENCRYPTED PRIVATE KEY-----`
)

func TestParsePrivateKeys(t *testing.T) {
	assert := assert.New(t)

	expected := []Key{
		Key{Path: "b34192be4389bac3f49f8feeee2aefc478b36cab1c9f56574d7e29e452fc0185", Gun: "docker.io/marcofranssen/openjdk"},
		Key{Path: "b624efeddff59751e8b6b59abb45383555103d702e7d3f46fbaaa9a8ac144ab8", Gun: "docker.io/marcofranssen/whalesay"},
		Key{Path: "d11b2a4c0651b833f0b1a536068c5ba8588041abe7d058aab95fffc5b78c98bd", Gun: "docker.io/marcofranssen/nginx"},
	}

	reader := ioutil.NopCloser(strings.NewReader(certificates))

	targetChan := make(chan Key)
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processPrivateKeys(ctx, reader, targetChan, errChan)

	targets, err := getTargets(targetChan, errChan)

	assert.NoError(err)
	assert.ElementsMatch(expected, targets)
}
