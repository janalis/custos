<?php
class Shape {
    public function scale(
        $factor,
        <weak_warning descr="Prefer null as the default value for this parameter.">$min = 1</weak_warning>,
        $max = null,
        <weak_warning descr="Prefer null as the default value for this parameter.">$tags = ['a']</weak_warning>
    ) {}

    private function tint(<weak_warning descr="Prefer null as the default value for this parameter.">$hue = 'red'</weak_warning>) {}
}

class Circle extends Shape {
    public function scale($factor, $min = 1, $max = null, $tags = ['a']) {}

    private function tint(<weak_warning descr="Prefer null as the default value for this parameter.">$hue = 'blue'</weak_warning>) {}
}

interface Drawable {
    public function draw(<weak_warning descr="Prefer null as the default value for this parameter.">$times = 2</weak_warning>);
}

function pad(<weak_warning descr="Prefer null as the default value for this parameter.">?int $width = 8</weak_warning>, int $fill = 0) {}
function wrap(<weak_warning descr="Prefer null as the default value for this parameter.">string|null $glue = ','</weak_warning>, string $edge = '') {}
/** @param int|null $n */
function count_up(<weak_warning descr="Prefer null as the default value for this parameter.">$n = 10</weak_warning>, $m = NULL) {}
$cb = function (<weak_warning descr="Prefer null as the default value for this parameter.">$limit = 3</weak_warning>) {};
$af = fn(<weak_warning descr="Prefer null as the default value for this parameter.">&$x = false</weak_warning>) => $x;
