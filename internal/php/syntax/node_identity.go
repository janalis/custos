package syntax

// NodeIdentity is a comparable key for one node within a parsed tree. It retains
// the node's storage and distinguishes nodes with identical kinds and spans.
// It is not a persistent identifier across parses.
type NodeIdentity struct{ base *nodeBase }

// Identity returns a compact identity key without retaining the node's dynamic
// interface type. A nil interface has the zero identity.
func Identity(node Node) NodeIdentity {
	if node == nil {
		return NodeIdentity{}
	}
	return NodeIdentity{base: node.base()}
}
