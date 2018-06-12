package btcatom_test

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/republicprotocol/atom-go/adapters/btcatom"
	"github.com/republicprotocol/atom-go/adapters/btcclient"
	"github.com/republicprotocol/atom-go/drivers/btc/regtest"
	"github.com/republicprotocol/atom-go/services/atom"
)

const CHAIN = "regtest"
const RPC_USERNAME = "testuser"
const RPC_PASSWORD = "testpassword"

func randomBytes32() [32]byte {
	randString := [32]byte{}
	_, err := rand.Read(randString[:])
	if err != nil {
		panic(err)
	}
	return randString
}

var _ = Describe("Bitcoin", func() {

	// Don't run on CI
	atom.LocalContext("atom swap", func() {

		var connection btcclient.Connection
		var cmd *exec.Cmd
		var aliceAddr, bobAddr string // btcutil.Address
		var bobAtom atom.Atom
		var aliceBobAtom, bobAliceAtom atom.Atom
		var aliceData, bobData []byte
		var aliceBobData, bobAliceData []byte
		var secret, hashLock [32]byte
		validity := time.Now().Unix() + 10000

		BeforeSuite(func() {
			var err error

			cmd = regtest.Start()
			time.Sleep(5 * time.Second)

			connection, err = btcclient.Connect("regtest", RPC_USERNAME, RPC_PASSWORD)
			Ω(err).Should(BeNil())

			go func() {
				err = regtest.Mine(connection)
				Ω(err).Should(BeNil())
			}()

			time.Sleep(5 * time.Second)

			_aliceAddr, err := regtest.NewAccount(connection, "alice", 1000000000)
			Ω(err).Should(BeNil())
			aliceAddr = _aliceAddr.EncodeAddress()

			_bobAddr, err := regtest.NewAccount(connection, "bob", 1000000000)
			Ω(err).Should(BeNil())
			bobAddr = _bobAddr.EncodeAddress()

			bobAliceAtom = NewBitcoinAtom([]byte(bobAddr), connection)
			bobAliceData, err = bobAliceAtom.Serialize()
			Ω(err).Should(BeNil())

			aliceBobAtom = NewBitcoinAtom([]byte(aliceAddr), connection)
			aliceBobData, err = aliceBobAtom.Serialize()
			Ω(err).Should(BeNil())

			fmt.Println("Alice")
			fmt.Println(aliceAddr)
			fmt.Println("Bob")
			fmt.Println(bobAddr)
		})

		AfterSuite(func() {
			connection.Shutdown()
			regtest.Stop(cmd)
		})

		It("can initiate a btc atomic swap", func() {
			aliceAtom := NewBitcoinAtom([]byte(aliceAddr), connection)
			err := aliceAtom.Deserialize(bobAliceData)
			Ω(err).Should(BeNil())
			secret = randomBytes32()
			hashLock = sha256.Sum256(secret[:])
			err = aliceAtom.Initiate(hashLock, big.NewInt(3000000), validity)
			Ω(err).Should(BeNil())
			aliceData, err = aliceAtom.Serialize()
			Ω(err).Should(BeNil())
		})

		It("can audit and initiate an atomic swap", func() {
			err := bobAliceAtom.Deserialize(aliceData)
			Ω(err).Should(BeNil())
			_secretHash, _from, _to, _value, _expiry, err := bobAliceAtom.Audit()

			fmt.Println(string(_from), string(_to))
			Ω(err).Should(BeNil())
			Ω(_secretHash).Should(Equal(hashLock))
			Ω(_to).Should(Equal([]byte(bobAddr)))
			Ω(_value).Should(Equal(big.NewInt(3000000)))
			Ω(_expiry).Should(Equal(validity))

			bobAtom = NewBitcoinAtom([]byte(bobAddr), connection)
			err = bobAtom.Deserialize(aliceBobData)
			Ω(err).Should(BeNil())
			err = bobAtom.Initiate(_secretHash, big.NewInt(3000000), validity)
			Ω(err).Should(BeNil())
			bobData, err = bobAtom.Serialize()
			Ω(err).Should(BeNil())
		})

		It("can audit atom details and reveal the secret", func() {
			err := aliceBobAtom.Deserialize(bobData)
			Ω(err).Should(BeNil())
			_secretHash, _from, _to, _value, _expiry, err := aliceBobAtom.Audit()
			Ω(err).Should(BeNil())
			Ω(_secretHash).Should(Equal(hashLock))
			Ω(_to).Should(Equal([]byte(bobAddr)))
			Ω(_value).Should(Equal(big.NewInt(3000000)))
			Ω(_expiry).Should(Equal(validity))

			err = aliceBobAtom.Redeem(aliceSecret)
			Ω(err).Should(BeNil())
		})

		// It("can retrieve the secret from his contract and complete the swap", func() {
		// 	secret, err := bobAtom.AuditSecret()
		// 	Ω(err).Should(BeNil())

		// 	err = bobAliceAtom.Redeem(secret)
		// 	Ω(err).Should(BeNil())
		// })
	})
})