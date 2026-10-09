<?php
class SemanticNode { public function ping(): void {} }
function anonymousNullableMember() {
    $item = new class {
        public function node(): ?SemanticNode { return null; }
    };
    $item->node()<warning descr="Possible null dereference.">-></warning>ping();
}
/**
 * @param SemanticNode $node
 * @psalm-param SemanticNode|null $node
 * @phpstan-param SemanticNode|null $node
 */
function prefixedParameter($node) {
    $node->ping();
}
function prefixedInline() {
    /** @psalm-var SemanticNode|null $node */
    $node = unresolved();
    $node->ping();
}
