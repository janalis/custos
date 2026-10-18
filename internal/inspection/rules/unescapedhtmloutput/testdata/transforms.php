<?php
function decodedRequestOutput()
{
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_entity_decode(htmlentities($_GET['value']))</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">urldecode(rawurlencode($_POST['value']))</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">base64_decode(base64_encode($_COOKIE['value']))</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">str_replace('&lt;', '<', htmlspecialchars($_GET['value']))</error>;
    echo '<p>' . rawurlencode($_GET['value']);
    echo '<p>' . base64_encode($_GET['value']);
    echo '<p>' . htmlspecialchars(html_entity_decode($_GET['value']));
}
function decodedRewrittenRequest()
{
    $_GET['value'] = htmlspecialchars($_GET['value']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars_decode($_GET['value'])</error>;
    echo '<p>' . strtoupper($_GET['value']);
}
