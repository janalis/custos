<?php
trait OptionalNode {
    public const self|null NODE = null;
}

class TreeNode {
    use OptionalNode;
    public function label(): string { return 'node'; }
}

class LeafNode extends TreeNode {}

function readNode() {
    $node = LeafNode::NODE;
    return <warning descr="Possible null dereference.">$node</warning>->label();
}

function readCheckedNode() {
    $node = TreeNode::NODE;
    if ($node !== null) {
        return $node->label();
    }
    return '';
}
