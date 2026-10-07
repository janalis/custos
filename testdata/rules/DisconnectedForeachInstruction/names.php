<?php
namespace Views;

function compact(...$names) { return []; }

function draw(array $keys, \Renderer $out) {
    foreach ($keys as $key) {
        echo $out->render(\COMPACT('key'));
        <weak_warning descr="Statement does not depend on the loop; move it out.">echo $out->render(compact('key'));</weak_warning>
    }
}
