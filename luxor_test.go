package luxor_test

import (
	"encoding/json"
	"net"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"vbeffa/luxor"
)

var _ = Describe("Luxor RPC Server", func() {
	var s luxor.Server

	It("Should accept authorize and subscribe calls", func() {
		signal := make(chan (int), 1)

		ready := func() {
			signal <- 1
		}

		go s.Start(ready)

		<-signal

		conn, err := net.Dial("tcp", ":1234")
		Expect(err).To(BeNil())
		defer conn.Close()

		_, err = conn.Write([]byte(`{"params": ["slush.miner1", "hungryhippo123"], "id": 2, "method": "mining.authorize"}` + "\n"))
		Expect(err).To(BeNil())

		buf := make([]byte, 1024)
		resp := make([]byte, 0)
		size := 0
		for {
			n, err := conn.Read(buf)
			Expect(err).To(BeNil())
			size += n
			resp = append(resp, buf...)
			if strings.Contains(string(resp), "\n") {
				break
			}
		}
		resp = resp[:size]
		Expect(strings.TrimSpace(string(resp))).To(Equal(
			`{"error": null, "id": 2, "result": true}`))

		_, err = conn.Write([]byte(`{"id": 1, "method": "mining.subscribe", "params": []}` + "\n"))
		Expect(err).To(BeNil())

		buf = make([]byte, 1024)
		resp = make([]byte, 0)
		size = 0
		for {
			n, err := conn.Read(buf)
			Expect(err).To(BeNil())
			size += n
			resp = append(resp, buf...)
			if strings.Contains(string(resp), "\n") {
				break
			}
		}
		resp = resp[:size]

		type subscribeResp struct {
			ID     int
			Result []interface{}
			Error  string
		}
		var subResp subscribeResp
		err = json.Unmarshal(resp, &subResp)
		Expect(err).To(BeNil())
		Expect(subResp.ID).To(Equal(1))
		Expect(subResp.Result[1]).To(Equal("31303030303030303031"))
		Expect(subResp.Result[2]).To(Equal(float64(4)))
	})
})
