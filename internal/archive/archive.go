package archive

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
	pgperrors "github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
	pgppacket "github.com/ProtonMail/go-crypto/openpgp/packet"

	"github.com/canonical/chisel/internal/cache"
	"github.com/canonical/chisel/internal/control"
	"github.com/canonical/chisel/internal/deb"
)

type Archive interface {
	Options() *Options
	Fetch(pkg string) (io.ReadCloser, error)
	Exists(pkg string) bool
}

type Options struct {
	Label      string
	Version    string
	Arch       string
	Suites     []string
	Components []string
	CacheDir   string
	PublicKeys map[uint64][]*pgppacket.PublicKey
}

func Open(options *Options) (Archive, error) {
	var err error
	if options.Arch == "" {
		options.Arch, err = deb.InferArch()
	} else {
		err = deb.ValidateArch(options.Arch)
	}
	if err != nil {
		return nil, err
	}
	return openUbuntu(options)
}

type fetchFlags uint

const (
	fetchBulk    fetchFlags = 1 << iota
	fetchDefault fetchFlags = 0
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

var httpDo = httpClient.Do

var bulkClient = &http.Client{
	Timeout: 5 * time.Minute,
}

var bulkDo = bulkClient.Do

type ubuntuArchive struct {
	options Options
	indexes []*ubuntuIndex
	cache   *cache.Cache
}

type ubuntuIndex struct {
	label     string
	version   string
	arch      string
	suite     string
	component string
	release   control.Section
	packages  control.File
	archive   *ubuntuArchive
}

func (a *ubuntuArchive) Options() *Options {
	return &a.options
}

func (a *ubuntuArchive) Exists(pkg string) bool {
	_, _, err := a.selectPackage(pkg)
	return err == nil
}

func (a *ubuntuArchive) selectPackage(pkg string) (control.Section, *ubuntuIndex, error) {
	var selectedVersion string
	var selectedSection control.Section
	var selectedIndex *ubuntuIndex
	for _, index := range a.indexes {
		section := index.packages.Section(pkg)
		if section != nil && section.Get("Filename") != "" {
			version := section.Get("Version")
			if selectedVersion == "" || deb.CompareVersions(selectedVersion, version) < 0 {
				selectedVersion = version
				selectedSection = section
				selectedIndex = index
			}
		}
	}
	if selectedVersion == "" {
		return nil, nil, fmt.Errorf("cannot find package %q in archive", pkg)
	}
	return selectedSection, selectedIndex, nil
}

func (a *ubuntuArchive) Fetch(pkg string) (io.ReadCloser, error) {
	section, index, err := a.selectPackage(pkg)
	if err != nil {
		return nil, err
	}
	suffix := section.Get("Filename")
	logf("Fetching %s...", suffix)
	reader, err := index.fetch("../../"+suffix, section.Get("SHA256"), fetchBulk)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

const ubuntuURL = "http://archive.ubuntu.com/ubuntu/"
const ubuntuPortsURL = "http://ports.ubuntu.com/ubuntu-ports/"

func openUbuntu(options *Options) (Archive, error) {
	if len(options.Components) == 0 {
		return nil, fmt.Errorf("archive options missing components")
	}
	if len(options.Suites) == 0 {
		return nil, fmt.Errorf("archive options missing suites")
	}
	if len(options.Version) == 0 {
		return nil, fmt.Errorf("archive options missing version")
	}
	if options.PublicKeys == nil {
		return nil, fmt.Errorf("archive has no public keys")
	}

	archive := &ubuntuArchive{
		options: *options,
		cache: &cache.Cache{
			Dir: options.CacheDir,
		},
	}

	for _, suite := range options.Suites {
		var release control.Section
		for _, component := range options.Components {
			index := &ubuntuIndex{
				label:     options.Label,
				version:   options.Version,
				arch:      options.Arch,
				suite:     suite,
				component: component,
				release:   release,
				archive:   archive,
			}
			if release == nil {
				err := index.parseRelease()
				if err != nil {
					return nil, err
				}
				release = index.release
				err = index.checkComponents(options.Components)
				if err != nil {
					return nil, err
				}
			}
			err := index.fetchIndex()
			if err != nil {
				return nil, err
			}
			archive.indexes = append(archive.indexes, index)
		}
	}

	return archive, nil
}

func (index *ubuntuIndex) fetchRelease() ([]byte, error) {
	reader, err := index.fetch("InRelease", "", fetchDefault)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// nameToHash returns a hash for a given OpenPGP name, or 0
// if the name isn't known. See RFC 4880, section 9.4.
func nameToHash(h string) crypto.Hash {
	switch h {
	case "SHA1":
		return crypto.SHA1
	case "SHA224":
		return crypto.SHA224
	case "SHA256":
		return crypto.SHA256
	case "SHA384":
		return crypto.SHA384
	case "SHA512":
		return crypto.SHA512
	case "SHA3-256":
		return crypto.SHA3_256
	case "SHA3-512":
		return crypto.SHA3_512
	}
	return crypto.Hash(0)
}

var (
	SHA1     Hash = cryptoHash{2, crypto.SHA1}
	SHA256   Hash = cryptoHash{8, crypto.SHA256}
	SHA384   Hash = cryptoHash{9, crypto.SHA384}
	SHA512   Hash = cryptoHash{10, crypto.SHA512}
	SHA224   Hash = cryptoHash{11, crypto.SHA224}
	SHA3_256 Hash = cryptoHash{12, crypto.SHA3_256}
	SHA3_512 Hash = cryptoHash{14, crypto.SHA3_512}
)


// HashToHashIdWithSha1 returns an OpenPGP hash id which corresponds the given Hash,
// allowing instances of SHA1
func HashToHashIdWithSha1(h crypto.Hash) (id byte, ok bool) {
	for id, hash := range HashById {
		if hash.HashFunc() == h {
			return id, true
		}
	}

	if h == SHA1.HashFunc() {
		return SHA1.Id(), true
	}

	return 0, false
}

// hashForSignature returns a pair of hashes that can be used to verify a
// signature. The signature may specify that the contents of the signed message
// should be preprocessed (i.e. to normalize line endings). Thus this function
// returns two hashes. The second should be used to hash the message itself and
// performs any needed preprocessing.
func hashForSignature(hashFunc crypto.Hash, sigType pgppacket.SignatureType) (hash.Hash, hash.Hash, error) {
	if _, ok := algorithm.HashToHashIdWithSha1(hashFunc); !ok {
		return nil, nil, pgperrors.UnsupportedError("unsupported hash function")
	}
	if !hashFunc.Available() {
		return nil, nil, pgperrors.UnsupportedError("hash not available: " + strconv.Itoa(int(hashFunc)))
	}
	h := hashFunc.New()

	switch sigType {
	case pgppacket.SigTypeBinary:
		return h, h, nil
	case pgppacket.SigTypeText:
		return h, NewCanonicalTextHash(h), nil
	}

	return nil, nil, pgperrors.UnsupportedError("unsupported signature type: " + strconv.Itoa(int(sigType)))
}

func (index *ubuntuIndex) verifyRelease(signedData []byte) ([]byte, error) {
	var err error

	block, _ := clearsign.Decode(signedData)
	if block == nil {
		return nil, fmt.Errorf("cannot decode InRelease clear-sign block")
	}

	var expectedHashes []crypto.Hash
	for _, v := range block.Headers {
		for _, name := range v {
			expectedHash := nameToHash(name)
			if uint8(expectedHash) == 0 {
				return nil, pgperrors.StructuralError("unknown hash algorithm in cleartext message headers")
			}
			expectedHashes = append(expectedHashes, expectedHash)
		}
	}
	if len(expectedHashes) == 0 {
		expectedHashes = append(expectedHashes, crypto.MD5)
	}

	signature := block.ArmoredSignature.Body

	//return openpgp.CheckDetachedSignatureAndHash(keyring, bytes.NewBuffer(b.Bytes), , expectedHashes, config)

	var issuerKeyId uint64
	var hashFunc crypto.Hash
	var sigType pgppacket.SignatureType
	var keys []*pgppacket.PublicKey
	var p pgppacket.Packet
	var sig *pgppacket.Signature

	expectedHashesLen := len(expectedHashes)
	packets := pgppacket.NewReader(signature)
	for {
		p, err = packets.Next()
		if err == io.EOF {
			return nil, pgperrors.ErrUnknownIssuer
		}
		if err != nil {
			return nil, err
		}

		var ok bool
		sig, ok = p.(*pgppacket.Signature)
		if !ok {
			return nil, pgperrors.StructuralError("non signature packet found")
		}
		if sig.IssuerKeyId == nil {
			return nil, pgperrors.StructuralError("signature doesn't have an issuer")
		}
		issuerKeyId = *sig.IssuerKeyId
		hashFunc = sig.Hash
		sigType = sig.SigType

		for i, expectedHash := range expectedHashes {
			if hashFunc == expectedHash {
				break
			}
			if i+1 == expectedHashesLen {
				return nil, pgperrors.StructuralError("hash algorithm mismatch with cleartext message headers")
			}
		}

		keys = index.archive.options.PublicKeys[issuerKeyId]
		if len(keys) > 0 {
			break
		}
	}

	if len(keys) == 0 {
		panic("unreachable")
	}

	h, wrappedHash, err := hashForSignature(hashFunc, sigType)
	if err != nil {
		return nil, nil, err
	}

	if _, err := io.Copy(wrappedHash, signed); err != nil && err != io.EOF {
		return nil, nil, err
	}

	for _, key := range keys {
		err = key.PublicKey.VerifySignature(h, sig)
		if err == nil {
			return sig, key.Entity, checkSignatureDetails(&key, sig, config)
		}
	}

	return nil, nil, err
	return openpgp.CheckDetachedSignatureAndHash(keyring, bytes.NewBuffer(b.Bytes), b.ArmoredSignature.Body, expectedHashes, config)

	if _, err := block.VerifySignature(index.archive.options.PublicKeys, nil); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}
	return block.Plaintext, nil
}

func (index *ubuntuIndex) parseRelease() error {
	logf("Fetching %s %s %s suite details...", index.label, index.version, index.suite)
	signed, err := index.fetchRelease()
	if err != nil {
		return fmt.Errorf("cannot fetch release: %w", err)
	}
	content, err := index.verifyRelease(signed)
	if err != nil {
		return fmt.Errorf("cannot verify release: %w", err)
	}
	ctrl, err := control.ParseReader("Label", bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("parsing archive Release file: %v", err)
	}
	section := ctrl.Section("Ubuntu")
	if section == nil {
		section = ctrl.Section("UbuntuProFIPS")
		if section == nil {
			return fmt.Errorf("corrupted archive Release file: no Ubuntu section")
		}
	}
	logf("Release date: %s", section.Get("Date"))

	index.release = section
	return nil
}

func (index *ubuntuIndex) fetchIndex() error {
	digests := index.release.Get("SHA256")
	packagesPath := fmt.Sprintf("%s/binary-%s/Packages", index.component, index.arch)
	digest, _, _ := control.ParsePathInfo(digests, packagesPath)
	if digest == "" {
		return fmt.Errorf("%s is missing from %s %s component digests", packagesPath, index.suite, index.component)
	}

	logf("Fetching index for %s %s %s %s component...", index.label, index.version, index.suite, index.component)
	reader, err := index.fetch(packagesPath+".gz", digest, fetchBulk)
	if err != nil {
		return err
	}
	ctrl, err := control.ParseReader("Package", reader)
	if err != nil {
		return fmt.Errorf("parsing archive Package file: %v", err)
	}

	index.packages = ctrl
	return nil
}

func (index *ubuntuIndex) checkComponents(components []string) error {
	releaseComponents := strings.Fields(index.release.Get("Components"))
	for _, c1 := range components {
		found := false
		for _, c2 := range releaseComponents {
			if c1 == c2 {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("archive has no component %q", c1)
		}
	}
	return nil
}

func (index *ubuntuIndex) fetch(suffix, digest string, flags fetchFlags) (io.ReadCloser, error) {
	reader, err := index.archive.cache.Open(digest)
	if err == nil {
		return reader, nil
	} else if err != cache.MissErr {
		return nil, err
	}

	baseURL := ubuntuURL
	if index.arch != "amd64" && index.arch != "i386" {
		baseURL = ubuntuPortsURL
	}

	var url string
	if strings.HasPrefix(suffix, "pool/") {
		url = baseURL + suffix
	} else {
		url = baseURL + "dists/" + index.suite + "/" + suffix
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create HTTP request: %v", err)
	}
	var resp *http.Response
	if flags&fetchBulk != 0 {
		resp, err = bulkDo(req)
	} else {
		resp, err = httpDo(req)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot talk to archive: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// ok
	case 401, 404:
		return nil, fmt.Errorf("cannot find archive data")
	default:
		return nil, fmt.Errorf("error from archive: %v", resp.Status)
	}

	body := resp.Body
	if strings.HasSuffix(suffix, ".gz") {
		reader, err := gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("cannot decompress data: %v", err)
		}
		defer reader.Close()
		body = reader
	}

	writer := index.archive.cache.Create(digest)
	defer writer.Close()

	_, err = io.Copy(writer, body)
	if err == nil {
		err = writer.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("cannot fetch from archive: %v", err)
	}

	return index.archive.cache.Open(writer.Digest())
}
