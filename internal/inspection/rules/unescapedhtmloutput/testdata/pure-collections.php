<?php
function html_collect($rows) {
    $out = [];
    foreach ($rows as $row) {
        $out[] = $row['label'];
    }
    return $out;
}
function html_spread($values) {
    return [...$values, 'tail'];
}
function html_predicate($value) {
    return str_contains($value, 'safe') ? 'fixed' : 'fallback';
}
function html_mode($values) {
    return count($values, COUNT_NORMAL) === 1 ? 'one' : 'other';
}
function html_collection_output(): void {
    echo '<p>' . html_collect([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_collect([['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . html_spread(['fixed', $_GET['unused']])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_spread(['fixed', $_GET['label']])[1]</error>;
    echo '<p>' . html_predicate('safe');
    echo '<p>' . html_mode(['fixed']);
}
