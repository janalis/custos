<?php
function html_callback_field($row) {
    return $row['label'];
}
function html_callback_selected($row) {
    return $row['keep'];
}
function html_callback_map($rows) {
    return array_map('html_callback_field', $rows);
}
function html_callback_filter($rows) {
    return array_filter($rows, 'html_callback_selected');
}
function html_find_label($rows) {
    foreach ($rows as $row) {
        if ($row['keep']) {
            return $row['label'];
        }
    }
    return 'fixed';
}
function html_continue_labels($rows) {
    $out = [];
    foreach ($rows as $row) {
        if (!$row['keep']) {
            continue;
        }
        $out[] = $row['label'];
    }
    return $out;
}
function html_array_identity($rows, $fallback) {
    return $rows === ['label' => 'fixed'] ? 'fixed' : $fallback;
}
function html_explicit_cast($value) {
    return (int) $value;
}
function html_callback_control_output(): void {
    echo '<p>' . html_callback_map([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_callback_map([['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_callback_filter(['x' => ['keep' => true, 'label' => $_GET['label']]])['x']['label']</error>;
    echo '<p>' . html_find_label([['keep' => false, 'label' => $_GET['unused']]]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_find_label([['keep' => true, 'label' => $_GET['label']]])</error>;
    echo '<p>' . html_continue_labels([['keep' => true, 'label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_continue_labels([['keep' => true, 'label' => $_GET['label']]])[0]</error>;
    echo '<p>' . html_array_identity(['label' => 'fixed'], $_GET['fallback']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_array_identity(['label' => 'other'], $_GET['fallback'])</error>;
    echo '<p>' . html_explicit_cast('12');
}
