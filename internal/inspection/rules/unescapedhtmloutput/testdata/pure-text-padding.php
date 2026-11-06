<?php
function padding_rows($rows, $fill) { return array_pad($rows, 3, $fill); }
function padding_label($label) { return strtoupper(str_pad(rtrim($label, '/'), 6, '.')); }
function padding_steps($last) { return range(1, $last); }
function padding_output(): void {
    echo '<p>' . padding_rows(['fixed'], $_GET['fill'])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">padding_rows(['fixed'], $_GET['fill'])[2]</error>;
    echo '<p>' . padding_rows(['a' => ['label' => 'fixed', 'other' => $_GET['unused']]], 'x')['a']['label'];
    echo '<p>' . padding_label('ab/');
    echo '<p>' . padding_steps(3)[2];
}
