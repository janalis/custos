<?php
class Store { public static array $rows = []; }
function normalise() {
    foreach (Store::$rows as $key => $value) {
        <weak_warning descr="Iterate '$value' by reference and assign to it directly instead of writing through the key.">store::$rows[$key]</weak_warning> = trim($value);
    }
}
