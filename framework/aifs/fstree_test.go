package aifs

import (
	"strings"
	"testing"
)

func TestFSTree(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	if tree.Name != "ai" { t.Errorf("root = %q", tree.Name) }
	for _, name := range []string{"features", "sessions", "models", "knowledge", "topology", "config"} {
		if _, ok := tree.Children[name]; !ok { t.Errorf("missing: %s", name) }
	}
}

func TestFSTreeFeatures(t *testing.T) {
	tree := FSTree(NewServer(nil))
	for _, name := range []string{"chat", "complete", "embed", "rag"} {
		node, ok := tree.Children["features"].Children[name]
		if !ok { t.Errorf("missing feature: %s", name); continue }
		if !strings.Contains(node.ReadFn(), "name: "+name) { t.Errorf("%s missing name", name) }
	}
}

func TestFSTreeModels(t *testing.T) {
	tree := FSTree(NewServer(nil))
	models := tree.Children["models"]
	if !strings.Contains(models.Children["list"].ReadFn(), "claude") { t.Error("list missing claude") }
	if !strings.Contains(models.Children["active"].ReadFn(), "claude") { t.Error("active wrong") }
	if err := models.Children["active"].WriteFn("gpt-4o"); err != nil { t.Error(err) }
	if !strings.Contains(models.Children["active"].ReadFn(), "gpt-4o") { t.Error("not updated") }
}

func TestFSTreeSessions(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	if err := tree.Children["sessions"].Children["new"].Children["ctl"].WriteFn("alice"); err != nil {
		t.Fatal(err)
	}
	sess := srv.ListSessions()
	if len(sess) != 1 { t.Fatalf("got %d sessions", len(sess)) }
	node := SessionFSNode(sess[0])
	if err := node.Children["ctl"].WriteFn("what is 9P?"); err != nil { t.Fatal(err) }
	if !strings.Contains(node.Children["history"].ReadFn(), "what is 9P?") { t.Error("history missing") }
	if err := node.Children["context"].WriteFn("pwd=/home/user"); err != nil { t.Fatal(err) }
	if !strings.Contains(node.Children["context"].ReadFn(), "pwd=/home/user") { t.Error("context missing") }
	if !strings.Contains(node.Children["meta"].ReadFn(), "alice") { t.Error("meta missing user") }
}

func TestFSTreeKnowledge(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	kg := tree.Children["knowledge"]
	cat := srv.graph.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := srv.graph.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	srv.graph.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})
	if !strings.Contains(kg.Children["stats"].ReadFn(), "total_atoms: 3") { t.Error("stats wrong") }
	if !strings.Contains(kg.Children["concepts"].ReadFn(), "Cat") { t.Error("concepts missing") }
	if !strings.Contains(kg.Children["relations"].ReadFn(), "InheritanceLink") { t.Error("relations missing") }
}

func TestFSTreeTopology(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	top := tree.Children["topology"]
	srv.topology.AddVertex("node-a", [4]float64{1, 0, 0, 0})
	srv.topology.AddVertex("node-b", [4]float64{0, 1, 0, 0})
	srv.topology.AddEdge(0, 1, 1.0)
	if !strings.Contains(top.Children["vertices"].ReadFn(), "node-a") { t.Error("missing vertex") }
	if err := top.Children["lifecycle"].WriteFn("growth"); err != nil { t.Error(err) }
	if !strings.Contains(top.Children["lifecycle"].ReadFn(), "growth") { t.Error("phase not set") }
}

func TestFSNodeStat(t *testing.T) {
	node := file("test", func() string { return "hello" })
	stat := node.Stat()
	if !strings.Contains(stat, "test") { t.Error("missing name") }
	if !strings.Contains(stat, "5") { t.Error("missing size") }
}
