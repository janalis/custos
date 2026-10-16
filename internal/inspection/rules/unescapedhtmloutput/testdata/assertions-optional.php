<?php
function optional_html_sanitizer(): void {
    $value = $_GET['title'];
    assert(($value = htmlspecialchars($value)) !== '');
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $value;</error>
}
function optional_numeric_cast(): void {
    $value = $_GET['count'];
    assert(($value = (int) $value) >= 0);
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $value;</error>
}
function optional_back_edge($again): void {
    $value = 1;
    do {
        <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $value;</error>
        assert(($value = $_GET['count']) !== '');
    } while ($again);
}
