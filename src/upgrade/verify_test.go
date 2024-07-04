package upgrade

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	checksum, err := os.ReadFile("../test/signing/checksums.txt")
	assert.NoError(t, err)

	signature, err := os.ReadFile("../test/signing/checksums.txt.sig")
	assert.NoError(t, err)

	OK := validateSignature(checksum, signature)
	assert.True(t, OK)
}

func TestVerifyFail(t *testing.T) {
	checksum, err := os.ReadFile("../test/signing/checksums.txt")
	assert.NoError(t, err)

	signature, err := os.ReadFile("../test/signing/checksums.txt.invalid.sig")
	assert.NoError(t, err)

	OK := validateSignature(checksum, signature)
	assert.False(t, OK)
}
