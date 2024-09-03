package htmlutil

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const dummyHtml = `
<!DOCTYPE html>
<html lang="id-ID">
	<span class="dtkframebar__menu__kanal__icon">
		<p>
			<img data-src="image1" alt="">
		</p>		
	</span>
	<span class="dtkframebar__menu__kanal__icon">
		<img data-src="image2" alt="">
	</span>
</html>`

func TestFindNode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(dummyHtml))
	if err != nil {
		t.Fatal(err.Error())
	}

	node := FindNode(doc, DefaultFilter{
		Type: html.ElementNode,
		Data: "img",
	})

	if node == nil {
		t.Fatal("expect <img> element is found")
	}

	t.Logf("found node: %#v\n", *node)
}

func TestFindAllNode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(dummyHtml))
	if err != nil {
		t.Fatal(err.Error())
	}

	nodes := FindAllNode(doc, DefaultFilter{
		Type: html.ElementNode,
		Data: "img",
	})

	if len(nodes) == 0 {
		t.Fatal("no <img> element is found")
	}

	if len(nodes) < 2 {
		t.Fatal("there's should be 2 <img> element")
	}

	for i, node := range nodes {
		t.Logf("found node %d: %#v\n", i+1, *node)
	}
}
