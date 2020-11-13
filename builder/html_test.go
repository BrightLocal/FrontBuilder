package builder

import "testing"

func TestHTML(t *testing.T) {
	h := NewHTMLFile("test.html")
	j := NewJSFile("test.js", []byte("import"))
	h.InjectJS(j)
	if err := h.Render("test_2.html", true); err != nil {
		t.Fatal(err)
	}
}
