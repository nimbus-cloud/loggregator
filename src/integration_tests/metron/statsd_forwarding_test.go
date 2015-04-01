package integration_test

import (
	"net"
	"time"
	"os/exec"

	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/storeadapter"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Statsd support", func() {
	var fakeDoppler net.PacketConn

	BeforeEach(func() {
		fakeDoppler = eventuallyListensForUDP("localhost:3457")

		node := storeadapter.StoreNode{
			Key:   "/healthstatus/doppler/z1/doppler_z1/0",
			Value: []byte("localhost"),
		}

		adapter := etcdRunner.Adapter()
		adapter.Create(node)
		adapter.Disconnect()
		time.Sleep(200 * time.Millisecond) // FIXME: wait for metron to discover the fake doppler ... better ideas welcome
	})

	AfterEach(func() {
		fakeDoppler.Close()
	})

	Context("with a fake statsd client", func() {
		It("outputs gauges as signed value metric messages", func(done Done) {
			connection, err := net.Dial("udp", "localhost:51162")
			Expect(err).ToNot(HaveOccurred())
			defer connection.Close()

			statsdmsg := []byte("fake-origin.test.gauge:23|g")
			_, err = connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			readBuffer := make([]byte, 65535)
			readCount, _, _ := fakeDoppler.ReadFrom(readBuffer)
			readData := make([]byte, readCount)
			copy(readData, readBuffer[:readCount])
			readData = readData[32:]

			var receivedEnvelope events.Envelope
			Expect(proto.Unmarshal(readData, &receivedEnvelope)).To(Succeed())

			Expect(receivedEnvelope.GetValueMetric()).To(Equal(basicValueMetric()))
			Expect(receivedEnvelope.GetOrigin()).To(Equal("fake-origin"))
			close(done)
		}, 5)
	})

	Context("with a Ruby statsd client", func() {
		It("forwards gauges as signed value metric messages", func(done Done) {
			clientCommand := exec.Command("/bin/sh", "startStatsdRubyClient.sh")

			clientInput, err := clientCommand.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			clientSession, err := gexec.Start(clientCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			clientInput.Write([]byte("gauge test.gauge 23\n"))
			Eventually(metronSession).Should(gbytes.Say("StatsdListener: Read "))

			readBuffer := make([]byte, 65535)
			readCount, _, _ := fakeDoppler.ReadFrom(readBuffer)
			readData := make([]byte, readCount)
			copy(readData, readBuffer[:readCount])
			Expect(len(readData)).To(BeNumerically(">", 32))
			readData = readData[32:]

			var receivedEnvelope events.Envelope
			Expect(proto.Unmarshal(readData, &receivedEnvelope)).To(Succeed())

			Expect(receivedEnvelope.GetValueMetric()).To(Equal(basicValueMetric()))
			Expect(receivedEnvelope.GetOrigin()).To(Equal("testNamespace"))

			clientInput.Close()
			clientSession.Kill().Wait()
			close(done)
		}, 5)
	})
})
