<?php
class Item {
    public function name() {
        // the namespace is empty: the value is '\Item', which ::class cannot produce
        return <weak_warning descr="Use Item::class instead of the class name string (::class has no leading backslash).">__NAMESPACE__ . '\Item'</weak_warning>;
    }
}
