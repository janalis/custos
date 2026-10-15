<?php
function alternate_request_names() {
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . ${'_GET'}['title'];</error>
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $GLOBALS['_POST']['title'];</error>
}
function global_reader() {
    global $round_global;
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $round_global;</error>
}
function global_writer() {
    global $round_global;
    $round_global = $_GET['title'];
}
function escaped_request() { echo htmlspecialchars(${'_GET'}['title']); }
