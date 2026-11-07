<?php
function scan_labels($labels) { return array_unique($labels); }
function scan_find($labels, $label) { return array_search($label, $labels, true); }
function scan_widest($widths) { return max($widths) . 'px'; }
function scan_area($sides) { return array_product($sides); }
function scan_mirror($text) { return strrev($text) . substr_count($text, 'a'); }
function scan_output(): void {
    echo '<p>' . scan_labels(['fixed', 'other', 'fixed'])[1];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">scan_labels([$_GET['label']])[0]</error>;
    echo '<p>' . scan_find(['a' => 'fixed'], 'fixed');
    echo '<p>' . scan_widest([3, 9, 2]);
    echo '<p>' . scan_area([2, 3, 4]);
    echo '<p>' . scan_mirror('banana');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">scan_mirror($_GET['text'])</error>;
}
