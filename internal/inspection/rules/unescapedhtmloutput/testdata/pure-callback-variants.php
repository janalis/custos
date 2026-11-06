<?php
function variant_keep_key($key) { return $key === 'label'; }
function variant_keep_both($value, $key) { return $key === 'label'; }
function variant_filter_key($rows) { return array_filter($rows, 'variant_keep_key', ARRAY_FILTER_USE_KEY); }
function variant_filter_both($rows) { return array_filter($rows, 'variant_keep_both', ARRAY_FILTER_USE_BOTH); }
function variant_pick_second($left, $right) { return $right['label']; }
function variant_map($left, $right) { return array_map('variant_pick_second', $left, $right); }
function variant_trim($rows) { return array_map('trim', $rows); }
function variant_length($rows) { return array_map('strlen', $rows); }
function variant_match($rows, $fallback) { return match ($rows) { [1] => 'fixed', default => $fallback }; }
function variant_member($rows, $fallback) { return in_array($rows, [[1]], true) ? 'fixed' : $fallback; }
function variant_break($rows) {
    $out = [];
    foreach ($rows as $row) {
        if (!$row['keep']) { break; }
        $out[] = $row['label'];
    }
    return $out;
}
function variant_output(): void {
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_filter_key(['label' => $_GET['label']])['label']</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_filter_both(['label' => $_GET['label']])['label']</error>;
    echo '<p>' . variant_map([$_GET['unused']], [['label' => 'fixed']])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_map([1], [['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_trim([$_GET['label']])[0]</error>;
    echo '<p>' . variant_length(['fixed'])[0];
    echo '<p>' . variant_match([1], $_GET['unused']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_match([2], $_GET['label'])</error>;
    echo '<p>' . variant_member([1], $_GET['unused']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_member([2], $_GET['label'])</error>;
    echo '<p>' . variant_break([['keep' => true, 'label' => 'fixed'], ['keep' => false], []])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">variant_break([['keep' => true, 'label' => $_GET['label']], ['keep' => false]])[0]</error>;
}
