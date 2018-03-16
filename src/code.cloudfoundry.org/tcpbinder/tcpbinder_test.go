package main_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Binder test", func() {
	var session *gexec.Session
	BeforeSuite(func() {
		bin, err := gexec.Build("code.cloudfoundry.org/tcpbinder")
		Expect(err).NotTo(HaveOccurred())
		session, err = gexec.Start(exec.Command("sudo", "--non-interactive", bin), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Listening on socket at /var/tmp/com.docker.vmnetd.socket"))
	})

	AfterSuite(func() {
		gexec.Start(exec.Command("sudo", "--non-interactive", "kill", fmt.Sprintf("%d", session.Command.Process.Pid)), GinkgoWriter, GinkgoWriter)
		gexec.KillAndWait()
		gexec.CleanupBuildArtifacts()
	})

	var conn *net.UnixConn
	BeforeEach(func() {
		var err error
		conn, err = net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: "/var/tmp/com.docker.vmnetd.socket",
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		conn.Close()
	})

	It("binds ports", func() {
		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		Expect(sendBindAddr(conn, "10.245.0.2", 1888)).To(Succeed())
		ln, _, err := recvBindAddr(conn, "10.245.0.2", 1888)
		Expect(err).NotTo(HaveOccurred())
		defer ln.Close()

		msg := "Hello from test"
		go sendMessage("10.245.0.2", 1888, msg)
		Expect(readFromListener(ln)).To(Equal(msg))
	})

	Context("binding to a bound port", func() {
		var prior net.Listener
		BeforeEach(func() {
			var err error
			prior, err = net.Listen("tcp", "10.245.0.2:1888")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() { prior.Close() })

		It("sends an error", func() {
			Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
			Expect(recvHello(conn)).To(Equal("CFD3V"))
			Expect(sendBindAddr(conn, "10.245.0.2", 1888)).To(Succeed())
			_, b, _ := recvBindAddr(conn, "10.245.0.2", 1888)
			fmt.Println("msg received", b)
			Expect(b[0]).To(Equal(uint8(48)))
		})
	})
})

func sendHello(conn *net.UnixConn, id string, version uint32, sha1 string) error {
	if _, err := conn.Write([]byte(id)); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.LittleEndian, version); err != nil {
		return err
	}
	_, err := conn.Write([]byte(sha1))
	return err
}

func recvHello(conn *net.UnixConn) (string, error) {
	bytes := make([]byte, 49, 49)
	if n, err := io.ReadFull(conn, bytes); err != nil {
		return "", err
	} else if n != 49 {
		return "", fmt.Errorf("Expected to read 49 bytes, read %d", n)
	}
	return string(bytes[0:5]), nil
}

func sendBindAddr(conn *net.UnixConn, ip string, port uint16) error {
	conn.Write([]byte{0x6})
	b := []byte(net.ParseIP(ip).To4())
	conn.Write(append([]byte{}, b[3], b[2], b[1], b[0]))
	return binary.Write(conn, binary.LittleEndian, port)
}

func recvBindAddr(conn *net.UnixConn, ip string, port uint16) (net.Listener, []byte, error) {
	b := make([]byte, 8, 8)
	oob := make([]byte, 16, 16)
	if _, _, _, _, err := conn.ReadMsgUnix(b, oob); err != nil {
		return nil, b, err
	}
	if b[0] != 0 {
		return nil, b, errors.New("Look at b")
	}
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, b, err
	}
	fds, err := syscall.ParseUnixRights(&scms[0])
	if err != nil {
		return nil, b, err
	}
	file := os.NewFile(uintptr(fds[0]), fmt.Sprintf("tcp:%s:%d"))
	defer file.Close()
	ln, err := net.FileListener(file)
	return ln, b, err
}

func sendMessage(host string, port uint16, mesg string) {
	defer GinkgoRecover()
	wconn, err := net.Dial("tcp", "10.245.0.2:1888")
	Expect(err).NotTo(HaveOccurred())
	defer wconn.Close()
	wconn.Write([]byte(mesg))
}

func readFromListener(ln net.Listener) (string, error) {
	conn, err := ln.Accept()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	received := make([]byte, 15, 15)
	_, err = conn.Read(received)
	if err != nil {
		return "", err
	}
	return string(received), nil
}