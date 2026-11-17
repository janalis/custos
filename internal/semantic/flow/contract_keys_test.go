package flow

import (
	"slices"
	"testing"
)

// answerFor returns what the contract key gave while summaries were made.
func (s *summaryResolver) answerFor(key string) answer {
	if s.keys == nil {
		return answer{}
	}
	if id, known := s.keys.known(key); known && int(id) < len(s.answers) {
		return s.answers[id]
	}
	return answer{}
}

// readsKey reports that reads holds the contract key.
func (s *summaryResolver) readsKey(reads *readSet, key string) bool {
	if s.keys == nil {
		return false
	}
	id, known := s.keys.known(key)
	return known && reads.has(id)
}

func TestContractKeysNumberEachKeyOnce(t *testing.T) {
	var keys contractKeys
	if _, known := keys.known("a"); known {
		t.Fatal("a numbered before use")
	}
	a, b := keys.id("a"), keys.id("b")
	if a == b || keys.id("a") != a || keys.key(b) != "b" {
		t.Fatalf("a=%d b=%d", a, b)
	}
}

func TestReadSets(t *testing.T) {
	var none *readSet
	if none.has(1) || none.meets(&readSet{}) || !none.empty() {
		t.Fatal("nil set holds nothing")
	}
	none.each(func(uint32) { t.Fatal("nil set has no member") })
	s := &readSet{}
	s.addAll(nil)
	s.addAll(&readSet{})
	if !s.empty() || s.meets(nil) {
		t.Fatal("empty set")
	}
	s.add(3)
	s.add(200)
	s.add(64)
	other := &readSet{}
	other.add(700)
	if s.meets(other) {
		t.Fatal("disjoint sets meet")
	}
	other.addAll(s)
	if !other.meets(s) || !other.has(700) || !other.has(200) || other.has(4) || other.has(5000) || other.empty() {
		t.Fatal("union")
	}
	var got []uint32
	other.each(func(id uint32) { got = append(got, id) })
	if !slices.Equal(got, []uint32{3, 64, 200, 700}) {
		t.Fatalf("members %v", got)
	}
	small := &readSet{words: make([]uint64, 1, 8)}
	small.add(130)
	if !small.has(130) || cap(small.words) != 8 {
		t.Fatal("growth within capacity")
	}
}

// A key numbered for another reader of the table is no answer of this one.
func TestChangedAnswersSkipKeysNotRead(t *testing.T) {
	var keys contractKeys
	keys.id("unread")
	read := keys.id("read")
	answers := make([]answer, read+1)
	answers[read] = answer{known: true, ok: true}
	lookup := func(key string) (Summary, bool) {
		if key == "unread" {
			t.Fatal("unread key looked up")
		}
		return Summary{}, false
	}
	differ := changedAnswers(&keys, answers, map[string]Summary{}, lookup)
	if !differ.has(read) || differ.has(0) {
		t.Fatal("only the read key differs")
	}
	var e summaryResolver
	if e.answeredKey("read") {
		t.Fatal("nothing answered")
	}
}
