package coldmfa

import (
	"bytes"
	"filippo.io/age"
	"filippo.io/age/armor"
	"io"
	"math/rand/v2"
	"strings"
)

func EncryptMfaCodeBackupItems(codes []string, password string) ([]byte, error) {
	addPaddingLines := rand.Int()%(len(codes)) + 1
	paddedCodes := make([]string, len(codes)+addPaddingLines)
	copy(paddedCodes, codes)

	for i := 0; i < addPaddingLines; i++ {
		paddedCodes[len(codes)+i] = RandStringBytes(rand.Int()%30 + 10)
	}

	// Shuffle the padding with the codes to make sure that the positions of known strings aren't predictable
	rand.Shuffle(len(paddedCodes), func(i, j int) {
		paddedCodes[i], paddedCodes[j] = paddedCodes[j], paddedCodes[i]
	})

	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	armorWriter := armor.NewWriter(out)
	writer, err := age.Encrypt(armorWriter, recipient)
	defer func(writer io.WriteCloser) {
		_ = writer.Close()
	}(writer)
	if err != nil {
		return nil, err
	}

	if _, err := io.WriteString(writer, strings.Join(paddedCodes[:], "\n")); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}
	if err = armorWriter.Close(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func DecryptMfaCodeBackupItems(encrypted []byte, password string) ([]string, error) {
	recipient, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, err
	}

	armorReader := armor.NewReader(bytes.NewReader(encrypted))
	dec, err := age.Decrypt(armorReader, recipient)
	if err != nil {
		return nil, err
	}

	outBytes := &bytes.Buffer{}
	if _, err := io.Copy(outBytes, dec); err != nil {
		return nil, err
	}

	lines := strings.Split(outBytes.String(), "\n")

	out := make([]string, 0)
	for _, line := range lines {
		if strings.Index(line, "{") == 0 {
			out = append(out, line)
		}
	}

	return out, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int()%len(letterBytes)]
	}
	return string(b)
}
