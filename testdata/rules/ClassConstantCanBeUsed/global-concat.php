<?php
class Item {
    public function name() {
        return <weak_warning descr="Use Item::class instead of the class name string.">__NAMESPACE__ . '\Item'</weak_warning>;
    }
}
