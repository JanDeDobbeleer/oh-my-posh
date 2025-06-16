package upgrade

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	stdruntime "runtime"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// This is based on the following key generation and validation.
// Generate a private key:
// openssl genpkey -algorithm Ed25519 -out private_key.pem
// Extract the public key:
// openssl pkey -in private_key.pem -pubout -out public_key.pem
// Sign the checksums.txt file:
// openssl pkeyutl -sign -inkey private_key.pem -out checksums.txt.sig -rawin -in checksums.txt
// Verify the signature:
// openssl pkeyutl -verify -pubin -inkey public_key.pem -sigfile checksums.txt.sig -rawin -in checksums.txt
// The public key is embedded in the binary.
// The private key is used to sign the checksums.txt file.
// The signature is embedded in the release.
// The checksums.txt file contains the checksums of the release assets.
// All checks are done in memory.
// Only then the binary is written to disk.

//go:embed public_key.pem
var publicKey []byte

func downloadAndVerify(cfg *Config) ([]byte, error) {
	extension := ""
	if stdruntime.GOOS == runtime.WINDOWS {
		extension = ".exe"
	}

	asset := fmt.Sprintf("posh-%s-%s%s", stdruntime.GOOS, stdruntime.GOARCH, extension)

	log.Debug("downloading asset:", asset)

	data, err := cfg.DownloadAsset(asset)
	if err != nil {
		log.Debug("failed to download asset")
		return nil, err
	}

	setState(verifying)

	err = verify(cfg, asset, data)
	if err != nil {
		log.Debug("failed to verify asset")
		return nil, err
	}

	return data, nil
}

func verify(cfg *Config, asset string, binary []byte) error {
	checksums, err := cfg.DownloadAsset("checksums.txt")
	if err != nil {
		log.Debug("failed to download checksums")
		return err
	}

	signature, err := cfg.DownloadAsset("checksums.txt.sig")
	if err != nil {
		log.Debug("failed to download checksums signature")
		return err
	}

	OK := validateSignature(checksums, signature)
	if !OK {
		log.Debug("failed to verify checksums signature")
		return fmt.Errorf("failed to verify checksums signature")
	}

	return validateChecksum(asset, checksums, binary)
}

func validateSignature(data, signature []byte) bool {
	ed25519PublicKey, err := loadPublicKey()
	if err != nil {
		log.Debug("failed to load public key")
		log.Error(err)
		return false
	}

	return ed25519.Verify(*ed25519PublicKey, data, signature)
}

func loadPublicKey() (*ed25519.PublicKey, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		log.Debug("failed to decode PEM block")
		return nil, fmt.Errorf("error parsing PEM block: key not found")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Debug("failed to parse public key")
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	ed25519PubKey, ok := pubKey.(ed25519.PublicKey)
	if !ok {
		log.Debug("failed to convert public key to ed25519")
		return nil, fmt.Errorf("invalid public key format: %v", err)
	}

	return &ed25519PubKey, nil
}

func validateChecksum(asset string, sha256sums, binary []byte) error {
	var assetChecksum string
	checksums := strings.SplitSeq(string(sha256sums), "\n")

	for line := range checksums {
		if !strings.HasSuffix(line, asset) {
			continue
		}

		assetChecksum = strings.Fields(line)[0]
		break
	}

	if len(assetChecksum) == 0 {
		log.Debug("failed to find checksum for asset")
		return fmt.Errorf("failed to find checksum for asset")
	}

	// calculate the checksum of the binary
	binaryChecksum := fmt.Sprintf("%x", sha256.Sum256(binary))

	if assetChecksum != binaryChecksum {
		log.Debugf("checksum mismatch, expected: %s, got: %s", assetChecksum, binaryChecksum)
		return fmt.Errorf("checksum mismatch")
	}

	return nil
}
