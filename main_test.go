package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func createUserScore(s string, v int) *UserScore {
	var u UserScore
	u.UserId = s
	u.Score = int64(v)
	return &u
}

func testHouseCup(t *testing.T, s string, p *UserScore) {
	r := houseCupHelper(s)
	if r == nil || p == nil {
		if r == p {
			return
		}
		t.Fatalf("Expected %+v not %+v", p, r)
	}

	if !proto.Equal(r, p) {
		t.Fatalf("Expected %+v not %+v", p, r)
	}
}

func TestHouseCupHandler(t *testing.T) {
	testHouseCup(t, "abc", nil)
	testHouseCup(t,
		"<@523502514092638217> <@523502514092638218> penalizeded 211 points",
		createUserScore("523502514092638218", -211))
	testHouseCup(t, "No match <@523502514092638218> penalizeded 211 points",
		createUserScore("523502514092638218", -211))
	testHouseCup(t, "<@523653040410984458> has been awarded 10 points!",
		createUserScore("523653040410984458", 10))
}
