<?php
namespace Views;

function compact(...$names) { return []; }

function draw(array $keys, \Renderer $out) {
    foreach ($keys as $key) {
        $out->send(\COMPACT('key'));
        <weak_warning descr="Statement does not depend on the loop; move it out.">$out->send(compact('key'));</weak_warning>
    }
}
