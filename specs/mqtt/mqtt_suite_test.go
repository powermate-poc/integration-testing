package devices_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"powermate-integration-testing/configuration"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ProvisioningResponse struct {
	Name        string `json:"name"`
	Arn         string `json:"arn"`
	Pem         string `json:"pem"`
	Public_key  string `json:"public_key"`
	Private_key string `json:"private_key"`
	Root_ca     string `json:"root_ca"`
}

type Measurement struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Message struct {
	Measurements []Measurement `json:"measurements"`
}

func TestSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MQTT")
}

var (
	config     *configuration.Configuration
	deviceName string
)

var _ = BeforeSuite(func() {
	config = configuration.Load()
})

var _ = BeforeEach(func() {
	id := uuid.New()
	deviceName = fmt.Sprintf("it-test-mqtt-%s", id.String())
	fmt.Println(deviceName)
})

var _ = AfterEach(func() {
	path, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s", deviceName))
	Expect(err).NotTo(HaveOccurred())

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

	client := &http.Client{}
	resp, err2 := client.Do(req)
	Expect(err2).NotTo(HaveOccurred())
	defer resp.Body.Close()

	Expect(resp.StatusCode).To(Or(Equal(http.StatusNoContent), Equal(http.StatusNotFound)))
})

var _ = Describe("MQTT Full test", func() {
	It("should save published message to database", func() {
		path, err := url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err := http.NewRequest(http.MethodPut, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		var provisioningResponse ProvisioningResponse
		err = json.Unmarshal(body, &provisioningResponse)
		Expect(err).NotTo(HaveOccurred())

		// create an mqtt client with the credentials
		options := mqtt.NewClientOptions()
		options.AddBroker(fmt.Sprintf("ssl://%s:8883", config.Broker))
		options.SetClientID(deviceName)

		// Load AWS IoT Core certificates
		cert, err := tls.X509KeyPair([]byte(provisioningResponse.Pem), []byte(provisioningResponse.Private_key))
		if err != nil {
			Fail(fmt.Sprintf("Error loading certificate/key: %s", err))
		}

		// Load the Root CA certificate
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(provisioningResponse.Root_ca))

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
		}

		options.SetTLSConfig(tlsConfig)

		mqtt := mqtt.NewClient(options)
		if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
			Fail(fmt.Sprintf("Error connecting to AWS IoT Core: %s", token.Error()))
		}

		message := Message{
			Measurements: []Measurement{
				{
					Name:  "x",
					Value: 420.69,
				},
			},
		}

		jsonMessage, err := json.Marshal(message)
		Expect(err).NotTo(HaveOccurred())

		token := mqtt.Publish(fmt.Sprintf("%s/data", deviceName), 0, false, jsonMessage)
		token.Wait()

		Expect(token.Error()).NotTo(HaveOccurred())

		fmt.Println("Published message")

		//// check data arrived correctly
		path, err = url.JoinPath(config.Host, fmt.Sprintf("/api/devices/%s/sensors/x/current", deviceName))
		Expect(err).NotTo(HaveOccurred())

		req, err = http.NewRequest(http.MethodGet, path, nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))

		response := make([]byte, 1000)

		iterations := 0

		// request data as long as it does not contain the value
		for !bytes.Contains(response, []byte("420.69")) {
			fmt.Println("Waiting for data to arrive")
			client := &http.Client{}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			_, err = resp.Body.Read(response)
			Expect(err).NotTo(HaveOccurred())
			// wait a bit
			time.Sleep(5 * time.Second)

			iterations++

			if iterations > 20 {
				Fail("Data did not arrive")
			}
		}
	})
})
