<?php
function optional_html_sanitizer(): void {
    $value = $_GET['title'];
    assert(($value = htmlspecialchars($value)) !== '');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$value</error>;
}
function optional_numeric_cast(): void {
    $value = $_GET['count'];
    assert(($value = (int) $value) >= 0);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$value</error>;
}
function optional_back_edge($again): void {
    $value = 1;
    do {
        echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$value</error>;
        assert(($value = $_GET['count']) !== '');
    } while ($again);
}
