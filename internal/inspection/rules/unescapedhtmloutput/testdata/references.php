<?php
function retainOutput(&$output, $ignored) {}
function replaceOutput(&$output, $ignored) { $output = ['label' => 'safe']; }

function unsafePositionalReference() {
    retainOutput($_GET, []);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>;
}
function unsafeNamedReference() {
    retainOutput(ignored: [], output: $_GET);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>;
}
function safePositionalReference() {
    replaceOutput($_GET, []);
    echo '<p>' . $_GET['label'];
}
function safeNamedReference() {
    replaceOutput(ignored: [], output: $_GET);
    echo '<p>' . $_GET['label'];
}
