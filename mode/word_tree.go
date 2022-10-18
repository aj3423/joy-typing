package mode

// A tree of many nodes is used to parse the speech words
// it works by searching the words one by one until a leaf is found,
// and use that leaf to execute the words
// To build a tree for commands like:
// 1. "turn off screen"       -> handler_1
// 2. "turn off sound"        -> handler_2
// 3. "turn off the computer" -> handler_3
// A tree is generated as:
//
//	turn
//	└── off
//	    ├── screen       -> handler_1
//	    ├── sound        -> handler_2
//	    └── the
//	        └── computer -> handler_3
type node[T any] struct {
	children map[string]*node[T]

	leaf T
	// The `leaf T` above cannot be compared to `nil` like `if leaf == nil`,
	// use this variable to indicate if it's nil
	// this variable can be removed once Golang has `nilable` type
	hasLeaf bool
}

func newNode[T any]() *node[T] {
	return &node[T]{children: make(map[string]*node[T])}
}

// Add a word path
func (n *node[T]) Set(words []string, leaf T) {
	if len(words) == 0 {
		n.leaf = leaf
		n.hasLeaf = true
		return
	}
	first := words[0]

	ch, exist := n.children[first]
	if !exist {
		ch = newNode[T]()
		n.children[first] = ch
	}
	ch.Set(words[1:], leaf)
}

// return value:
// 0: cost count of words that been scanned
// 1: the found leaf
// 2: `true` if a leaf is found
func (n *node[T]) Scan(words []string) (cost int, leaf T, hasLeaf bool) {
	var nilT T

	size := len(words)
	if size == 0 {
		return 0, nilT, false
	}
	ch, exist := n.children[words[0]]
	if !exist {
		return 0, nilT, false
	}

	restCost, leaf, hasLeaf := ch.Scan(words[1:]) // greedy scan as deep as possible

	if hasLeaf { // found a deeper result
		return 1 + restCost, leaf, true
	} else { // no more match, return current result
		return 1, ch.leaf, ch.hasLeaf
	}
}

// a `node` with a fallback leaf
// the fallback leaf is returned when `node.parse` doesn't match any result
type fallbackNode[T any] struct {
	node[T]

	hasFallback bool // cannot compare `fallback` to `nil`
	fallback    T
}

func newFallbackNode[T any]() *fallbackNode[T] {
	f := &fallbackNode[T]{}
	f.node = *newNode[T]()
	return f
}

func (n *fallbackNode[T]) SetFallback(fb T) {
	n.hasFallback = true
	n.fallback = fb
}

func (n *fallbackNode[T]) Scan(words []string) (int, T, bool) {
	cost, leaf, ok := n.node.Scan(words)
	if !ok && n.hasFallback {
		return 0, n.fallback, true
	}
	return cost, leaf, ok
}
