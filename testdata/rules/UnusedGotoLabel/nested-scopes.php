<?php
function process(array $items)
{
    <weak_warning descr="Label 'outer' is never targeted by a goto; remove it.">outer:</weak_warning>
    $each = function ($item) {
        inner:
        if (!$item) {
            goto inner;
        }
    };
    <weak_warning descr="Label 'inner' is never targeted by a goto; remove it.">inner:</weak_warning>
    $handler = new class {
        public function run() {
            goto retry;
            retry:
        }
    };
    <weak_warning descr="Label 'retry' is never targeted by a goto; remove it.">retry:</weak_warning>
    foreach ($items as $i) {
        if ($i) { goto finish; }
    }
    finish:
    return [$each, $handler];
}
