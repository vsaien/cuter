package stringx

import "github.com/vsaien/cuter/lib/lang"

type (
	scope struct {
		start int
		stop  int
	}

	Node struct {
		children map[rune]*Node
		end      bool
	}
)

func NewTrie(words []string) *Node {
	node := new(Node)
	for _, word := range words {
		node.add(word)
	}

	return node
}

func (n *Node) Filter(text string) (sentence string, keywords []string, found bool) {
	chars := []rune(text)
	if len(chars) == 0 {
		return text, nil, false
	}

	scopes := n.findKeywordScopes(chars)
	keywords = n.collectKeywords(chars, scopes)

	for _, match := range scopes {
		// we don't care about overlaps, not bringing a performance improvement
		n.replaceWithAsterisk(chars, match.start, match.stop)
	}

	return string(chars), keywords, len(keywords) > 0
}

func (n *Node) add(word string) {
	chars := []rune(word)
	if len(chars) == 0 {
		return
	}

	node := n
	for _, char := range chars {
		if node.children == nil {
			child := new(Node)
			node.children = map[rune]*Node{
				char: child,
			}
			node = child
		} else if child, ok := node.children[char]; ok {
			node = child
		} else {
			child := new(Node)
			node.children[char] = child
			node = child
		}
	}

	node.end = true
}

func (n *Node) collectKeywords(chars []rune, scopes []scope) []string {
	set := make(map[string]lang.PlaceholderType)
	for _, v := range scopes {
		set[string(chars[v.start:v.stop])] = lang.Placeholder
	}

	var i int
	keywords := make([]string, len(set))
	for k := range set {
		keywords[i] = k
		i++
	}

	return keywords
}

func (n *Node) findKeywordScopes(chars []rune) []scope {
	var scopes []scope
	size := len(chars)
	start := -1

	for i := 0; i < size; i++ {
		child, ok := n.children[chars[i]]
		if !ok {
			continue
		}

		if start < 0 {
			start = i
		}
		if child.end {
			scopes = append(scopes, scope{
				start: start,
				stop:  i + 1,
			})
		}

		for j := i + 1; j < size; j++ {
			grandchild, ok := child.children[chars[j]]
			if !ok {
				break
			}

			child = grandchild
			if child.end {
				scopes = append(scopes, scope{
					start: start,
					stop:  j + 1,
				})
			}
		}

		start = -1
	}

	return scopes
}

func (n *Node) replaceWithAsterisk(chars []rune, start, stop int) {
	for i := start; i < stop; i++ {
		chars[i] = '*'
	}
}
