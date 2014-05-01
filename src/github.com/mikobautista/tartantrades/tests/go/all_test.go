package gotest

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func Test_All(t *testing.T) {
	performBuild(t, "main/tradeservertest.go")
	performBuild(t, "../../server/main/resolverRunner.go")
	performBuild(t, "../../server/main/tradeServerRunner.go")

	performTest("../python/setup.py", "Performing Setup", t)

	performTest("../python/testSingleServer.py", "Test Single Server", t)
	performTest("../python/testMultipleServer.py", "Test Multiple Server", t)
	performTest("../python/testBuy.py", "Test Buy", t)
	performTest("../python/testRegisterUser.py", "Test Register User", t)
	performTest("../python/testLateStart.py", "Test Late Start", t)
	performTest("../python/testRecovery.py", "Test Recovery", t)

	performTest("../python/teardown.py", "Performing Teardown", t)

	t.Log("all tests passed")
}

func performBuild(t *testing.T, s string) {
	cmd := exec.Command("go", "build", s)
	err := cmd.Run()
	if err != nil {
		t.Error("error occurred")
	}
}

func performTest(s, msg string, t *testing.T) {
	fmt.Println("-------" + msg + "-------")
	cmd := exec.Command("python", s, "-v")
	var out bytes.Buffer
    cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Error("error occurred")
	}
	fmt.Println(out.String())
	fmt.Println("Success!")
}
