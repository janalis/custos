<?php
function checked_divide($left, $right) { return $left / $right; }
function checked_remainder($left, $right) { return $left % $right; }
function checked_bits($left, $right) { return ($left & $right) << 1; }
function checked_replace($text) { return str_replace('a', 'b', $text); }
function checked_repeat($text, $times) { return str_repeat($text, $times); }
function checked_reverse($rows) { return array_reverse($rows); }
function checked_switch($selector, $fallback) {
    switch ($selector) {
        default: return $fallback;
        case 1: return 'fixed';
    }
}
function checked_output(): void {
    echo '<p>' . checked_divide(6, 3);
    echo '<p>' . checked_remainder(7, 3);
    echo '<p>' . checked_bits(7, 3);
    echo '<p>' . checked_replace('a,a');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">checked_replace($_GET['label'])</error>;
    echo '<p>' . checked_repeat('fixed', 2);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">checked_repeat($_GET['label'], 2)</error>;
    echo '<p>' . checked_reverse([$_GET['unused'], 'fixed'])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">checked_reverse(['fixed', $_GET['label']])[0]</error>;
    echo '<p>' . checked_switch(1, $_GET['unused']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">checked_switch(2, $_GET['label'])</error>;
}
