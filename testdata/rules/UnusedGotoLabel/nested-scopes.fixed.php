<?php
function process(array $items)
{
    $each = function ($item) {
        inner:
        if (!$item) {
            goto inner;
        }
    };
    $handler = new class {
        public function run() {
            goto retry;
            retry:
        }
    };
    foreach ($items as $i) {
        if ($i) { goto finish; }
    }
    finish:
    return [$each, $handler];
}
