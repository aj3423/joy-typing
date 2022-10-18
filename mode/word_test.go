package mode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestMappingSpace(t *testing.T) {

	words := wordArray{"space", "space"}

	mappingTree := newNode[*replace]()

	mappingTree.Set([]string{"space"}, &replace{to: []string{" "}})

	newWords := wordArray{}
	for !words.empty() {
		scanCost, replacer, isMapped := mappingTree.Scan(words)
		if isMapped { // replace this word
			replacer.exec(&newWords) // this appends replaced word to `newWords`
			words.skip(scanCost)     // skip the words being scanned
		} else { // don't replace this word
			newWords.add1(words[0])
			words.skip(1) // skip the word being scanned
		}
	}

	ok := slices.Compare(newWords, wordArray{" ", " "})
	assert.Equal(t, ok, 0, fmt.Sprintf("newWords.len: %d, %v", len(newWords), newWords))

}
func TestMappingTree(t *testing.T) {

	words := wordArray{"aaa", "hello", "world", "super", "awesome"}

	mappingTree := newNode[*replace]()

	mappingTree.Set([]string{"hello", "world"}, &replace{to: []string{"good", "night"}})
	mappingTree.Set([]string{"super", "awesome"}, &replace{to: []string{"nice"}})

	newWords := wordArray{}
	for !words.empty() {
		scanCost, replacer, isMapped := mappingTree.Scan(words)
		if isMapped { // replace this word
			replacer.exec(&newWords) // this appends replaced word to `newWords`
			words.skip(scanCost)     // skip the words being scanned
		} else { // don't replace this word
			newWords.add1(words[0])
			words.skip(1) // skip the word being scanned
		}
	}

	ok := slices.Compare(newWords, wordArray{"aaa", "good", "night", "nice"})
	assert.Equal(t, ok, 0, fmt.Sprintf("%v", newWords))

}

func TestExecTree(t *testing.T) {

	words := wordArray{"1", "2", "3"}
	execTree := newFallbackNode[executorFactory]()

	execTree.SetFallback(&typingFactory{noSpace: true})

	var executors []executor

	for !words.empty() {
		scanCost, factory, isMapped := execTree.Scan(words)
		if isMapped { // executor found for current word(s)
			words.skip(scanCost)

			cost, ex, e := factory.parse(words)
			if e == nil {
				executors = append(executors, ex)
				words.skip(cost)
			}
		} else {
			// if no executor handles this word, skip this word
			words.skip(1)
		}
	}

	ok1 := slices.Compare(executors[0].(*typing).words, wordArray{"1"})
	ok2 := slices.Compare(executors[1].(*typing).words, wordArray{"2"})
	ok3 := slices.Compare(executors[2].(*typing).words, wordArray{"3"})
	assert.Equal(t, ok1, 0, "!ok1")
	assert.Equal(t, ok2, 0, "!ok2")
	assert.Equal(t, ok3, 0, "!ok3")
}

func TestExecTreeCrash_LaunchTerminal(t *testing.T) {

	words := wordArray{"launch"}
	execTree := newFallbackNode[executorFactory]()

	execTree.SetFallback(&typingFactory{noSpace: true})

	execTree.Set([]string{"launch", "terminal"}, &hotkeyFixFactory{keys: []string{"t", "control", "alt"}})

	var executors []executor

	for !words.empty() {
		scanCost, factory, isMapped := execTree.Scan(words)
		if isMapped { // executor found for current word(s)
			words.skip(scanCost)

			cost, ex, e := factory.parse(words)
			if e == nil {
				executors = append(executors, ex)
				words.skip(cost)
			}
		} else {
			// if no executor handles this word, skip this word
			words.skip(1)
		}
	}

	assert.Equal(t, executors[0].(*typing).words, wordArray{"1"}, "!ok1")
	assert.Equal(t, executors[1].(*typing).words, wordArray{"2"}, "!ok2")
	assert.Equal(t, executors[2].(*typing).words, wordArray{"3"}, "!ok3")
}
