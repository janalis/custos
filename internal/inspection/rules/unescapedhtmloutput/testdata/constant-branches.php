<?php
function constant_html_paths(): void {
    $title = 'safe';
    if (false) { $title = $_GET['title']; }
    echo '<p>' . $title;
    if (true) { $title = $_GET['title']; }
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$title</error>;
}
function selected_html_arms(): void {
    $title = false ? $_GET['title'] : 'safe';
    echo '<p>' . $title;
    $title = true ? $_GET['title'] : 'safe';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$title</error>;
}
function dead_html_output(): void {
    if (false) { echo '<p>' . $_GET['title']; }
}
function goto_html_label(): void {
    $title = 'safe';
    goto populate;
    if (false) {
    populate:
        $title = $_GET['title'];
    }
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$title</error>;
}
