package aifs

import (
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil)
	if s == nil { t.Fatal("NewServer returned nil") }
	if len(s.models.models) == 0 { t.Error("default models not registered") }
}

func TestSessionLifecycle(t *testing.T) {
	s := NewServer(nil)
	sess := s.NewSession("testuser")
	if sess.ID == 0 { t.Error("session ID should be > 0") }
	if sess.User != "testuser" { t.Errorf("user = %q, want %q", sess.User, "testuser") }
	sess.AddMessage("user", "hello")
	sess.AddMessage("assistant", "hi there")
	history := sess.History()
	if !strings.Contains(history, "hello") { t.Error("history missing user message") }
	if !strings.Contains(history, "hi there") { t.Error("history missing assistant message") }
	got, ok := s.GetSession(sess.ID)
	if !ok { t.Error("GetSession failed") }
	if got.ID != sess.ID { t.Error("GetSession returned wrong session") }
	if len(s.ListSessions()) != 1 { t.Errorf("ListSessions = %d, want 1", len(s.ListSessions())) }
}

func TestSessionMeta(t *testing.T) {
	s := NewServer(nil)
	sess := s.NewSession("bob")
	meta := sess.Meta()
	if !strings.Contains(meta, "bob") { t.Error("meta missing user") }
	if !strings.Contains(meta, "messages: 0") { t.Error("meta should show 0 messages") }
}

func TestModelRegistry(t *testing.T) {
	r := NewModelRegistry()
	r.Register(ModelConfig{Name: "test-model", Provider: "test", MaxTokens: 1000})
	if r.Active() != "test-model" { t.Errorf("active = %q", r.Active()) }
	m, ok := r.Get("test-model")
	if !ok { t.Error("Get failed") }
	if m.Provider != "test" { t.Errorf("provider = %q", m.Provider) }
	r.Register(ModelConfig{Name: "other-model", Provider: "test", MaxTokens: 2000})
	if err := r.SetActive("other-model"); err != nil { t.Errorf("SetActive: %v", err) }
	if r.Active() != "other-model" { t.Error("active not changed") }
	if err := r.SetActive("nonexistent"); err == nil { t.Error("should fail for nonexistent") }
}

func TestKnowledgeGraph(t *testing.T) {
	g := NewKnowledgeGraph()
	cat := g.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := g.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	g.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})
	cat2 := g.AddNode(ConceptNode, "Cat", TruthValue{0.8, 0.5})
	if cat2.ID != cat.ID { t.Error("duplicate node should return existing") }
	if len(g.ListConcepts()) != 2 { t.Errorf("concepts = %d", len(g.ListConcepts())) }
	if len(g.GetByName("Cat")) != 1 { t.Error("GetByName failed") }
	if len(g.GetIncoming(animal.ID)) != 1 { t.Error("incoming failed") }
	if len(g.Query(InheritanceLink, animal.ID, 1)) != 1 { t.Error("query failed") }
	if !strings.Contains(g.Stats(), "total_atoms: 3") { t.Error("stats wrong") }
}

func TestTruthValueMerge(t *testing.T) {
	merged := TruthValue{0.8, 0.9}.Merge(TruthValue{0.6, 0.7})
	if merged.Strength < 0.6 || merged.Strength > 0.8 { t.Errorf("strength %.3f", merged.Strength) }
	if merged.Confidence <= 0 || merged.Confidence >= 1 { t.Errorf("confidence %.3f", merged.Confidence) }
}

func TestTopology(t *testing.T) {
	top := NewTopology()
	id0 := top.AddVertex("alice", [4]float64{1, 0, 0, 0})
	id1 := top.AddVertex("bob", [4]float64{0, 1, 0, 0})
	id2 := top.AddVertex("carol", [4]float64{0, 0, 1, 0})
	top.AddEdge(id0, id1, 1.0); top.AddEdge(id1, id2, 0.5); top.AddEdge(id0, id2, 0.8)
	top.ComputeInfluence()
	top.AddCell("team-alpha", []int{id0, id1, id2})
	if !strings.Contains(top.ListVertices(), "alice") { t.Error("missing alice") }
	if !strings.Contains(top.ListEdges(), "1.000") { t.Error("missing weight") }
	if !strings.Contains(top.ListCells(), "team-alpha") { t.Error("missing cell") }
	top.SetPhase(PhaseGrowth)
	if top.Phase() != PhaseGrowth { t.Error("phase wrong") }
	if !strings.Contains(top.LifecycleStatus(), "growth") { t.Error("status wrong") }
}

func TestStereographicProject(t *testing.T) {
	p := StereographicProject([4]float64{1, 2, 3, 0})
	if p[0] != 1 || p[1] != 2 || p[2] != 3 { t.Errorf("got %v", p) }
}
