<?php
class Acc { public static $all = []; }
function collect(array $chunks) {
    foreach ($chunks as $chunk) {
        Acc::$all = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge(acc::$all, $chunk)</error>;
    }
}
