<?php
function build(array $opts) {
    $a = greet(...$opts);
    $b = greet(...['name' => 'Ann']);
}
