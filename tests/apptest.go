package tests

import (
	//"testing"
	"github.com/revel/revel/testing"
)

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	println("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) After() {
	println("Tear down")
}

/*
func (t *AppTest) TestMike() {
	println("Mike")
	t.Assertf("foo" == "bar", "Expected '%s', got '%s'", "foo", "bar")
}

func (t *AppTest) TestMike2() {
	println("Mike2")
	loc := "1"
	exp := "12345"
	got := LocationToSerialNumber(loc)
	t.Assertf(got != exp, "Expected '%s', got '%s'", exp, got)
}
*/
