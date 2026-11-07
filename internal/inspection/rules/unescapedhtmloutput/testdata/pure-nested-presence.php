<?php
function nested_html_label($row) {
    return ($row['a'] ?? ['label' => 'fallback'])['label'];
}
function nested_html_output(): void {
    echo '<p>' . nested_html_label([]);
    echo '<p>' . nested_html_label(['a' => ['label' => 'fixed'], 'unused' => $_GET['unused']]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">nested_html_label(['a' => ['label' => $_GET['label']]])</error>;
}
