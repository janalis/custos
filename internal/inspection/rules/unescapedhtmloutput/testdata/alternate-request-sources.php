<?php
function alternate_request_names() {
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">${'_GET'}['title']</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$GLOBALS['_POST']['title']</error>;
}
function global_reader() {
    global $round_global;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$round_global</error>;
}
function global_writer() {
    global $round_global;
    $round_global = $_GET['title'];
}
function escaped_request() { echo htmlspecialchars(${'_GET'}['title']); }
