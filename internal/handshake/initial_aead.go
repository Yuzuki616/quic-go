package handshake

import (
	"crypto"

	"github.com/lucas-clemente/quic-go/internal/protocol"
)

var quicVersion1Salt = []byte{0xaf, 0xbf, 0xec, 0x28, 0x99, 0x93, 0xd2, 0x4c, 0x9e, 0x97, 0x86, 0xf1, 0x9c, 0x61, 0x11, 0xe0, 0x43, 0x90, 0xa8, 0x99}

// NewInitialAEAD creates a new AEAD for Initial encryption / decryption.
func NewInitialAEAD(connID protocol.ConnectionID, pers protocol.Perspective) (LongHeaderSealer, LongHeaderOpener) {
	clientSecret, serverSecret := computeSecrets(connID)
	var mySecret, otherSecret []byte
	if pers == protocol.PerspectiveClient {
		mySecret = clientSecret
		otherSecret = serverSecret
	} else {
		mySecret = serverSecret
		otherSecret = clientSecret
	}
	myKey, myIV := computeInitialKeyAndIV(mySecret)
	otherKey, otherIV := computeInitialKeyAndIV(otherSecret)

	encrypter := qtlsAEADAESGCMTLS13(myKey, myIV)
	decrypter := qtlsAEADAESGCMTLS13(otherKey, otherIV)

	return newLongHeaderSealer(encrypter, newHeaderProtector(initialSuite, mySecret, true)),
		newLongHeaderOpener(decrypter, newAESHeaderProtector(initialSuite, otherSecret, true))
}

func computeSecrets(connID protocol.ConnectionID) (clientSecret, serverSecret []byte) {
	initialSecret := hkdfExtract(crypto.SHA256, connID, quicVersion1Salt)
	clientSecret = hkdfExpandLabel(crypto.SHA256, initialSecret, []byte{}, "client in", crypto.SHA256.Size())
	serverSecret = hkdfExpandLabel(crypto.SHA256, initialSecret, []byte{}, "server in", crypto.SHA256.Size())
	return
}

func computeInitialKeyAndIV(secret []byte) (key, iv []byte) {
	key = hkdfExpandLabel(crypto.SHA256, secret, []byte{}, "quic key", 16)
	iv = hkdfExpandLabel(crypto.SHA256, secret, []byte{}, "quic iv", 12)
	return
}
