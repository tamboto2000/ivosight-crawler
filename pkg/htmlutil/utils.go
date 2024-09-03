package htmlutil

import "golang.org/x/net/html"

type Filter interface {
	IsMatch(node *html.Node) bool
}

type DefaultFilter struct {
	Type       html.NodeType
	Data       string
	Namespace  string
	Attributes []html.Attribute
}

func (df DefaultFilter) IsMatch(node *html.Node) bool {
	if node.Type != df.Type {
		return false
	}

	if df.Data != "" && node.Data != df.Data {
		return false
	}

	if df.Namespace != "" && node.Namespace != df.Namespace {
		return false
	}

	for _, fattr := range df.Attributes {
		match := false

		for _, attr := range node.Attr {
			if fattr.Namespace != "" && attr.Namespace != fattr.Namespace {
				continue
			}

			if attr.Key != fattr.Key {
				continue
			}

			if attr.Val != fattr.Val {
				continue
			}

			match = true
			break
		}

		if !match {
			return false
		}
	}

	return true
}

func FindNode(node *html.Node, filter Filter) *html.Node {
	return findSingleNode(node, filter)
}

func findSingleNode(node *html.Node, filter Filter) *html.Node {
	if filter.IsMatch(node) {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		foundNode := findSingleNode(c, filter)
		if foundNode != nil {
			return foundNode
		}
	}

	return nil
}

func FindAllNode(node *html.Node, filter Filter) []*html.Node {
	return findAllNode(node, filter)
}

func findAllNode(node *html.Node, filter Filter) []*html.Node {
	var foundNodes []*html.Node

	if filter.IsMatch(node) {
		foundNodes = append(foundNodes, node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		foundCnodes := findAllNode(c, filter)
		foundNodes = append(foundNodes, foundCnodes...)
	}

	return foundNodes
}

func WalkNodes(node *html.Node, walkf func(node *html.Node) bool) {
	walkNodes(node, walkf)
}

func walkNodes(node *html.Node, walkf func(node *html.Node) bool) bool {
	if !walkf(node) {
		return false
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if !walkNodes(c, walkf) {
			return false
		}
	}

	return true
}

func WalkSkipNodes(node *html.Node, walkf func(node *html.Node) (bool, bool)) {
	walkSkipNodes(node, walkf)
}

func walkSkipNodes(node *html.Node, walkf func(node *html.Node) (bool, bool)) (bool, bool) {
	keepTraverse, keepWalking := walkf(node)
	if !keepWalking {
		return false, false
	}

	if keepTraverse {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			_, keepWalking = walkSkipNodes(c, walkf)
			if !keepWalking {
				return false, false
			}
		}
	}

	return keepTraverse, keepWalking
}
