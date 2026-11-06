<?php
function construction_fill($value) { return array_fill(3, 2, $value); }
function construction_keys($keys, $value) { return array_fill_keys($keys, $value); }
function construction_flip($values) { return array_flip($values); }
function construction_chunks($values) { return array_chunk($values, 1); }
function construction_sum($row) { return array_sum($row['nums']); }
function construction_output(): void {
    echo '<p>' . construction_fill(['label' => 'fixed', 'other' => $_GET['unused']])[4]['label'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">construction_fill($_GET['label'])[4]</error>;
    echo '<p>' . construction_keys(['x'], 'fixed')['x'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">construction_keys(['x'], $_GET['label'])['x']</error>;
    echo '<p>' . construction_flip(['first' => 'x', 'last' => 'x'])['x'];
    echo '<p>' . construction_chunks(['fixed', $_GET['unused']])[0][0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">construction_chunks(['fixed', $_GET['label']])[1][0]</error>;
    echo '<p>' . construction_sum(['nums' => [2, 3], 'other' => $_GET['unused']]);
}
