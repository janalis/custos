<?php
function batches(array $rows, string $sep) {
    foreach ($rows as $row) {
        // A generator created inside a closure is not this loop's output.
        <weak_warning descr="Statement does not depend on the loop; move it out.">register(function () use ($sep) { yield $sep; });</weak_warning>
        yield $sep;
        yield $row;
    }
}
