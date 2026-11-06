<?php
function html_selected_rows($rows) {
    $out = [];
    foreach ($rows as $row) {
        if ($row['keep']) {
            $out[] = $row['label'];
        }
    }
    return $out;
}
function html_tail_choice($rows, $fallback) {
    $sum = 0;
    foreach ($rows as $row) {
        $sum += $row;
    }
    return $sum === 3 ? 'fixed' : $fallback;
}
function html_type_choice($value, $fallback) {
    return is_string($value) ? 'fixed' : $fallback;
}
function html_ordered_values($rows) {
    return array_values($rows);
}
function html_default_keys($rows) {
    return array_keys($rows)[0];
}
function html_slice_label($value) {
    return substr($value, 1, 2);
}
function html_loop_query_output(): void {
    echo '<p>' . html_selected_rows([['keep' => true, 'label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_selected_rows([['keep' => true, 'label' => $_GET['label']]])[0]</error>;
    echo '<p>' . html_tail_choice([1, 2], $_GET['fallback']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_tail_choice([1, 1], $_GET['fallback'])</error>;
    echo '<p>' . html_type_choice('value', $_GET['fallback']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_type_choice([], $_GET['fallback'])</error>;
    echo '<p>' . html_ordered_values(['a' => 'fixed', 'b' => $_GET['unused']])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_ordered_values(['a' => 'fixed', 'b' => $_GET['label']])[1]</error>;
    echo '<p>' . html_default_keys(['fixed' => $_GET['unused']]);
    echo '<p>' . html_slice_label('fixed');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_slice_label($_GET['label'])</error>;
}
