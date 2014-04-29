package main

import (
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "regexp"
    "net/http"
    "strings"
    "github.com/mikobautista/tartantrades/logging"
)

type testFunc struct {
    name string
    f    func()
}

var (
    testRegex = flag.String("t", "", "test to run")
    expected  = flag.String("e", "", "expected output")
    testNum   = flag.Int("n", 0, "test number")
    hostPort  = flag.String("hp", "", "server hostport")
    x         = flag.Int("x", 0, "x coord for selling")
    y         = flag.Int("y", 0, "y coord for selling")
    token     = flag.String("token", "", "token for selling")
    item      = flag.Int("item", 0, "item number for buying")
    user      = flag.String("user", "", "username for registering")
    pass      = flag.String("pass", "", "password for registering")
    LOG       = logging.NewLogger(true)
    output    io.Writer
)

/////////////////////////////////////////////
//  test basic tradeserver operations
/////////////////////////////////////////////

func testRegister() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/servers/", *hostPort))
    testHandler(contents, response, err)
}

func testAvailableBlocks() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/availableblocks/", *hostPort))
    testHandler(contents, response, err)
}

func testSell() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/sell/?x=%d&y=%d&token=%s", *hostPort, *x, *y, *token))
    testHandler(contents, response, err)
}

func testStop() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/stop/", *hostPort))
    testHandler(contents, response, err)
}

func testBuy() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/buy/?item=%d&token=%s", *hostPort, *item, *token))
    testHandler(contents, response, err)
}

func testRegisterUser() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/register/?username=%s&password=%s", *hostPort, *user, *pass))
    testHandler(contents, response, err)
}

func testDeleteUser() {
    var contents = make([]byte, 0)
    response, err := http.Get(fmt.Sprintf("http://%s/deleteaccount/?token=%s", *hostPort, *token))
    testHandler(contents, response, err)
}

func main() {
    output = os.Stderr

    tests := []testFunc{
        {"testRegister", testRegister},
        {"testAvailableBlocks", testAvailableBlocks},
        {"testSell", testSell},
        {"testStop", testStop},
        {"testBuy", testBuy},
        {"testRegisterUser", testRegisterUser},
        {"testDeleteUser", testDeleteUser},
    }

    flag.Parse()

    // Run tests
    t := tests[*testNum]
    if b, err := regexp.MatchString(*testRegex, t.name); b && err == nil {
        // fmt.Printf("Running %s:\n", t.name)
        t.f()
    }
}

func testHandler(contents []byte, response *http.Response, err error) {
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    } else {
        defer response.Body.Close()
        contents, err = ioutil.ReadAll(response.Body)
        if err != nil {
            fmt.Printf("%s", err)
            os.Exit(1)
        }
    }
    contentString := strings.Trim(string(contents), "\n")
    expectedString := *expected
    if contentString != expectedString {
        fmt.Printf("content: %s\n", contentString)
        fmt.Printf("expected: %s\n", expectedString)
        fmt.Fprintln(output, "FAIL")
        os.Exit(1)
    } 
}
