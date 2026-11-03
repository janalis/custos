<?php
// A known descendant returning null rules out a type without null above it.
interface Labeled {
    /** @return string */
    public function label();

    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">name</weak_warning>();
}
class Node {
    /** @return static */
    public function setName($name) { return $this; }
    public function <weak_warning descr="Declare ': ?\Node' as the return type (update the whole hierarchy with a signature refactoring).">maybe</weak_warning>(bool $b) { return $b ? $this : null; }
    public function <weak_warning descr="Declare ': static' as the return type (update the whole hierarchy with a signature refactoring).">kept</weak_warning>() { return $this; }
}
