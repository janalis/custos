<?php
function transform_arrow($rows) { return array_map(fn($row) => $row['label'], $rows); }
function transform_slice($rows) { return array_slice($rows, 1); }
function transform_last($carry, $value) { return $value; }
function transform_fold($rows) { return array_reduce($rows, 'transform_last', 'fixed'); }
function transform_fold_fixed_callback($carry, $value) { return 'fixed'; }
function transform_fold_fixed($rows) { return array_reduce($rows, 'transform_fold_fixed_callback', 'fixed'); }
function transform_keys($rows) { return array_keys($rows, [1], true); }
function transform_destructure($rows) {
    $out = [];
    foreach ($rows as ['label' => $label, 'unused' => $unused]) {
        $out[] = $label;
    }
    return $out;
}
function transform_output(): void {
    echo '<p>' . transform_arrow([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">transform_arrow([['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . transform_slice([$_GET['unused'], 'fixed'])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">transform_slice(['fixed', $_GET['label']])[0]</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">transform_fold([$_GET['label']])</error>;
    echo '<p>' . transform_fold_fixed([$_GET['unused']]);
    echo '<p>' . transform_keys(['safe' => [1], 'other' => [2]])[0];
    echo '<p>' . transform_destructure([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">transform_destructure([['label' => $_GET['label'], 'unused' => 'fixed']])[0]</error>;
}
