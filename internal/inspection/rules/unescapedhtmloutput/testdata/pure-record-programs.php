<?php
function html_record_truth($rows, $fallback) {
    return $rows ? 'fixed' : $fallback;
}
function html_quiet_empty($rows, $fallback) {
    return empty($rows['label']) ? 'fixed' : $fallback;
}
function html_computed_record($key, $value) {
    return [$key => $value, 'fixed' => 'safe'];
}
function html_keyed_record($rows) {
    $out = [];
    foreach ($rows as $row) {
        $out[$row['id']] = $row['label'];
    }
    return $out;
}
function html_record_union($rows) {
    return ['fixed' => 'safe'] + $rows;
}
function html_record_pipeline($rows) {
    $selected = [];
    foreach ($rows as $row) {
        if ($row['keep']) {
            $selected[] = $row['label'];
        }
    }
    $out = [];
    foreach ($selected as $label) {
        $out[] = $label;
    }
    return $out;
}
function html_record_program_output(): void {
    echo '<p>' . html_record_truth([1], $_GET['fallback']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_record_truth([], $_GET['fallback'])</error>;
    echo '<p>' . html_quiet_empty([], $_GET['fallback']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_quiet_empty(['label' => [1]], $_GET['fallback'])</error>;
    echo '<p>' . html_computed_record('x', $_GET['value'])['fixed'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_computed_record('x', $_GET['value'])['x']</error>;
    echo '<p>' . html_keyed_record([['id' => 'x', 'label' => 'fixed', 'unused' => $_GET['unused']]])['x'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_keyed_record([['id' => 'x', 'label' => $_GET['label']]])['x']</error>;
    echo '<p>' . html_record_union(['fixed' => $_GET['unused']])['fixed'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_record_union(['unsafe' => $_GET['label']])['unsafe']</error>;
    echo '<p>' . html_record_pipeline([['keep' => true, 'label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_record_pipeline([['keep' => true, 'label' => $_GET['label']]])[0]</error>;
}
