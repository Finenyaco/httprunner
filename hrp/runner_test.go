package hrp

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/httprunner/httprunner/hrp/internal/scaffold"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func buildHashicorpGoPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", templatesDir+"debugtalk.bin", templatesDir+"plugin/debugtalk.go")
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("build hashicorp go plugin failed")
		os.Exit(1)
	}
}

func removeHashicorpGoPlugin() {
	log.Info().Msg("[teardown] remove hashicorp go plugin")
	os.Remove(templatesDir + "debugtalk.bin")
}

func buildHashicorpPyPlugin() {
	log.Info().Msg("[init] prepare hashicorp python plugin")
	pluginFile := templatesDir + "debugtalk.py"
	err := scaffold.CopyFile("templates/plugin/debugtalk.py", pluginFile)
	if err != nil {
		log.Error().Err(err).Msg("build hashicorp python plugin failed")
		os.Exit(1)
	}
}

func removeHashicorpPyPlugin() {
	log.Info().Msg("[teardown] remove hashicorp python plugin")
	os.Remove(templatesDir + "debugtalk.py")
}

func TestRunCaseWithGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	assertRunTestCases(t)
}

func TestRunCaseWithPythonPlugin(t *testing.T) {
	buildHashicorpPyPlugin()
	defer removeHashicorpPyPlugin()

	assertRunTestCases(t)
}

func assertRunTestCases(t *testing.T) {
	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").
			SetBaseURL("http://httpbin.org"),
		TestSteps: []IStep{
			NewStep("testcase1-step1").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			NewStep("testcase1-step2").
				GET("/user-agent").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
			NewStep("testcase1-step3").CallRefCase(
				&TestCase{
					Config: NewConfig("testcase1-step3-ref-case").SetBaseURL("http://httpbin.org"),
					TestSteps: []IStep{
						NewStep("ip").
							GET("/ip").
							Validate().
							AssertEqual("status_code", 200, "check status code").
							AssertEqual("headers.\"Content-Type\"", "application/json", "check http response Content-Type"),
					},
				},
			),
			NewStep("testcase1-step4").CallRefCase(&demoTestCaseWithPluginJSONPath),
		},
	}
	testcase2 := &TestCase{
		Config: NewConfig("TestCase2").SetWeight(3),
	}

	r := NewRunner(t)
	r.SetPluginLogOn()
	err := r.Run(testcase1, testcase2)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestRunCaseWithThinkTime(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testcases := []*TestCase{
		{
			Config: NewConfig("TestCase1"),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(2),
			},
		},
		{
			Config: NewConfig("TestCase2").
				SetThinkTime(thinkTimeIgnore, nil, 0),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(0.5),
			},
		},
		{
			Config: NewConfig("TestCase3").
				SetThinkTime(thinkTimeRandomPercentage, nil, 0),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: NewConfig("TestCase4").
				SetThinkTime(thinkTimeRandomPercentage, map[string]interface{}{"min_percentage": 2, "max_percentage": 3}, 2.5),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: NewConfig("TestCase5"),
			TestSteps: []IStep{
				// think time: 3s, random pct: {"min_percentage":1, "max_percentage":1.5}, limit: 4s
				NewStep("thinkTime").CallRefCase(&demoTestCaseWithThinkTimePath),
			},
		},
	}
	expectedMinValue := []float64{2, 0, 0.5, 2, 3}
	expectedMaxValue := []float64{2.5, 0.5, 2, 3, 10}
	for idx, testcase := range testcases {
		r := NewRunner(t)
		startTime := time.Now()
		err := r.Run(testcase)
		if err != nil {
			t.Fatalf("run testcase error: %v", err)
		}
		duration := time.Since(startTime)
		minValue := time.Duration(expectedMinValue[idx]*1000) * time.Millisecond
		maxValue := time.Duration(expectedMaxValue[idx]*1000) * time.Millisecond
		if duration < minValue || duration > maxValue {
			t.Fatalf("failed to test think time, expect value: [%v, %v], actual value: %v", minValue, maxValue, duration)
		}
	}
}

func TestRunCaseWithPluginJSON(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	err := NewRunner(nil).Run(&demoTestCaseWithPluginJSONPath) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}

func TestRunCaseWithPluginYAML(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	err := NewRunner(nil).Run(&demoTestCaseWithPluginYAMLPath) // hrp.Run(testCase)
	if err != nil {
		t.Fail()
	}
}

func TestRunCaseWithRefAPI(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	err := NewRunner(nil).Run(&demoTestCaseWithRefAPIPath)
	if err != nil {
		t.Fail()
	}

	testcase := &TestCase{
		Config: NewConfig("TestCase").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("run referenced api").CallRefAPI(&demoAPIGETPath),
		},
	}

	r := NewRunner(t)
	err = r.Run(testcase)
	if err != nil {
		t.Fail()
	}
}

func TestLoadTestCases(t *testing.T) {
	// load test cases from folder path
	tc := TestCasePath("../examples/demo-with-py-plugin/testcases/")
	testCases, err := loadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, len(testCases), 3) {
		t.Fail()
	}

	// load test cases from folder path, including sub folders
	tc = TestCasePath("../examples/demo-with-py-plugin/")
	testCases, err = loadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, len(testCases), 3) {
		t.Fail()
	}

	// load test cases from single file path
	tc = demoTestCaseWithPluginJSONPath
	testCases, err = loadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, 1, len(testCases)) {
		t.Fail()
	}

	// load test cases from TestCase instance
	testcase := &TestCase{
		Config: NewConfig("TestCase").SetWeight(3),
	}
	testCases, err = loadTestCases(testcase)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, len(testCases), 1) {
		t.Fail()
	}
}