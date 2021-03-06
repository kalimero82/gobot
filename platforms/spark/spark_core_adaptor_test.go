package spark

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hybridgroup/gobot"
)

// HELPERS

func createTestServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func getDummyResponseForPath(path string, dummy_response string, t *testing.T) *httptest.Server {
	dummy_data := []byte(dummy_response)

	return createTestServer(func(w http.ResponseWriter, r *http.Request) {
		actualPath := "/v1/devices" + path
		if r.URL.Path != actualPath {
			t.Errorf("Path doesn't match, expected %#v, got %#v", actualPath, r.URL.Path)
		}
		w.Write(dummy_data)
	})
}

func getDummyResponseForPathWithParams(path string, params []string, dummy_response string, t *testing.T) *httptest.Server {
	dummy_data := []byte(dummy_response)

	return createTestServer(func(w http.ResponseWriter, r *http.Request) {
		actualPath := "/v1/devices" + path
		if r.URL.Path != actualPath {
			t.Errorf("Path doesn't match, expected %#v, got %#v", actualPath, r.URL.Path)
		}

		r.ParseForm()

		for key, value := range params {
			if r.Form["params"][key] != value {
				t.Error("Expected param to be " + r.Form["params"][key] + " but was " + value)
			}
		}
		w.Write(dummy_data)
	})
}

func initTestSparkCoreAdaptor() *SparkCoreAdaptor {
	return NewSparkCoreAdaptor("bot", "myDevice", "token")
}

// TESTS

func TestSparkCoreAdaptor(t *testing.T) {
	var _ gobot.Adaptor = (*SparkCoreAdaptor)(nil)

	var a interface{} = initTestSparkCoreAdaptor()
	_, ok := a.(gobot.Adaptor)
	if !ok {
		t.Errorf("SparkCoreAdaptor{} should be a gobot.Adaptor")
	}
}

func TestNewSparkCoreAdaptor(t *testing.T) {
	// does it return a pointer to an instance of SparkCoreAdaptor?
	var a interface{} = initTestSparkCoreAdaptor()
	spark, ok := a.(*SparkCoreAdaptor)
	if !ok {
		t.Errorf("NewSparkCoreAdaptor() should have returned a *SparkCoreAdaptor")
	}

	gobot.Assert(t, spark.APIServer, "https://api.spark.io")
}

func TestSparkCoreAdaptorConnect(t *testing.T) {
	a := initTestSparkCoreAdaptor()
	gobot.Assert(t, len(a.Connect()), 0)
}

func TestSparkCoreAdaptorFinalize(t *testing.T) {
	a := initTestSparkCoreAdaptor()

	a.Connect()

	gobot.Assert(t, len(a.Finalize()), 0)
}

func TestSparkCoreAdaptorAnalogRead(t *testing.T) {
	// When no error
	response := `{"return_value": 5.2}`
	params := []string{"A1"}

	a := initTestSparkCoreAdaptor()
	testServer := getDummyResponseForPathWithParams("/"+a.DeviceID+"/analogread", params, response, t)

	a.setAPIServer(testServer.URL)

	val, _ := a.AnalogRead("A1")
	gobot.Assert(t, val, 5)

	testServer.Close()

	// When error
	testServer = createTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	defer testServer.Close()

	val, _ = a.AnalogRead("A1")
	gobot.Assert(t, val, 0)

}

func TestSparkCoreAdaptorPwmWrite(t *testing.T) {
	response := `{}`
	params := []string{"A1,1"}

	a := initTestSparkCoreAdaptor()
	testServer := getDummyResponseForPathWithParams("/"+a.DeviceID+"/analogwrite", params, response, t)
	defer testServer.Close()

	a.setAPIServer(testServer.URL)
	a.PwmWrite("A1", 1)
}

func TestSparkCoreAdaptorAnalogWrite(t *testing.T) {
	response := `{}`
	params := []string{"A1,1"}

	a := initTestSparkCoreAdaptor()
	testServer := getDummyResponseForPathWithParams("/"+a.DeviceID+"/analogwrite", params, response, t)
	defer testServer.Close()

	a.setAPIServer(testServer.URL)
	a.AnalogWrite("A1", 1)
}

func TestSparkCoreAdaptorDigitalWrite(t *testing.T) {
	// When HIGH
	response := `{}`
	params := []string{"D7,HIGH"}

	a := initTestSparkCoreAdaptor()
	testServer := getDummyResponseForPathWithParams("/"+a.DeviceID+"/digitalwrite", params, response, t)

	a.setAPIServer(testServer.URL)
	a.DigitalWrite("D7", 1)

	testServer.Close()
	// When LOW
	params = []string{"D7,LOW"}

	testServer = getDummyResponseForPathWithParams("/"+a.DeviceID+"/digitalwrite", params, response, t)
	defer testServer.Close()

	a.setAPIServer(testServer.URL)
	a.DigitalWrite("D7", 0)
}

func TestSparkCoreAdaptorDigitalRead(t *testing.T) {
	// When HIGH
	response := `{"return_value": 1}`
	params := []string{"D7"}

	a := initTestSparkCoreAdaptor()
	testServer := getDummyResponseForPathWithParams("/"+a.DeviceID+"/digitalread", params, response, t)

	a.setAPIServer(testServer.URL)

	val, _ := a.DigitalRead("D7")
	gobot.Assert(t, val, 1)
	testServer.Close()

	// When LOW
	response = `{"return_value": 0}`

	testServer = getDummyResponseForPathWithParams("/"+a.DeviceID+"/digitalread", params, response, t)

	a.setAPIServer(testServer.URL)

	val, _ = a.DigitalRead("D7")
	gobot.Assert(t, val, 0)

	testServer.Close()

	// When error
	testServer = createTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	defer testServer.Close()

	val, _ = a.DigitalRead("D7")
	gobot.Assert(t, val, -1)
}

func TestSparkCoreAdaptorSetAPIServer(t *testing.T) {

	a := initTestSparkCoreAdaptor()
	apiServer := "new_api_server"
	gobot.Refute(t, a.APIServer, apiServer)

	a.setAPIServer(apiServer)
	gobot.Assert(t, a.APIServer, apiServer)
}

func TestSparkCoreAdaptorDeviceURL(t *testing.T) {
	// When APIServer is set
	a := initTestSparkCoreAdaptor()
	a.setAPIServer("http://server")
	a.DeviceID = "devID"
	gobot.Assert(t, a.deviceURL(), "http://server/v1/devices/devID")

	//When APIServer is not set
	a = &SparkCoreAdaptor{name: "sparkie", DeviceID: "myDevice", AccessToken: "token"}

	gobot.Assert(t, a.deviceURL(), "https://api.spark.io/v1/devices/myDevice")
}

func TestSparkCoreAdaptorPinLevel(t *testing.T) {

	a := initTestSparkCoreAdaptor()

	gobot.Assert(t, a.pinLevel(1), "HIGH")
	gobot.Assert(t, a.pinLevel(0), "LOW")
	gobot.Assert(t, a.pinLevel(5), "LOW")
}

func TestSparkCoreAdaptorPostToSpark(t *testing.T) {

	a := initTestSparkCoreAdaptor()

	// When error on request
	vals := url.Values{}
	vals.Add("error", "error")
	resp, err := a.postToSpark("http://invalid%20host.com", vals)
	if err == nil {
		t.Errorf("postToSpark() should return an error when request was unsuccessful but returned", resp)
	}

	// When error reading body
	// Pending

	// When response.Status is not 200
	testServer := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	defer testServer.Close()

	resp, err = a.postToSpark(testServer.URL+"/existent", vals)
	if err == nil {
		t.Errorf("postToSpark() should return an error when status is not 200 but returned", resp)
	}

}
