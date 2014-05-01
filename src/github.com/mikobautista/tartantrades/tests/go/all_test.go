package gotest

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func Test_All(t *testing.T) {
	verbose := true
	db_user := "root"
	db_pw := "password"

	performBuild(t, "main/tradeservertest.go")
	performBuild(t, "../../server/main/resolverRunner.go")
	performBuild(t, "../../server/main/tradeServerRunner.go")

	performTest("../python/setup.py", "Performing Setup", t, verbose, db_user, db_pw)

	performTest("../python/testSingleServer.py", "Test Single Server", t, verbose, db_user, db_pw)
	performTest("../python/testMultipleServer.py", "Test Multiple Server", t, verbose, db_user, db_pw)
	performTest("../python/testBuy.py", "Test Buy", t, verbose, db_user, db_pw)
	performTest("../python/testRegisterUser.py", "Test Register User", t, verbose, db_user, db_pw)
	performTest("../python/testLateStart.py", "Test Late Start", t, verbose, db_user, db_pw)
	performTest("../python/testRecovery.py", "Test Recovery", t, verbose, db_user, db_pw)

	performTest("../python/teardown.py", "Performing Teardown", t, verbose, db_user, db_pw)

	t.Log("all tests passed")
}

func performBuild(t *testing.T, s string) {
	cmd := exec.Command("go", "build", s)
	err := cmd.Run()
	if err != nil {
		t.Error("error occurred")
	}
}

func performTest(s, msg string, t *testing.T, v bool, db_user, db_pw string) {
	fmt.Println("-------" + msg + "-------")
	var cmd *exec.Cmd
	if v {
		cmd = exec.Command("python", s, db_user, db_pw, "-v")
	} else {
		cmd = exec.Command("python", s, db_user, db_pw)
	}
	
	var out bytes.Buffer
    cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Error("error occurred")
	}
	fmt.Println(out.String())
	fmt.Println("Success!")
}
