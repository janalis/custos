<?php
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'] . '</p>';</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<p>{$_GET['label']}</p>\n";</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<p>$_GET[label] for $name</p>";</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>', $_GET['label'], '</p>', render();</error>
<error descr="Escape untrusted text for its HTML output context.">print '<p>' . ($_GET['label'] . '</p>')</error>;
<error descr="Escape untrusted text for its HTML output context.">echo <<<HTML
    <li class="row">{$_POST['label']}</li>
    HTML;</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . ($_GET['label'] ?? 'none');</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . ($_GET['short'] ? $_GET['label'] : 'none');</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . strtoupper($_GET['label']);</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . sprintf('%s items', $_GET['count']);</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . sprintf('%c', $_GET['count']);</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . implode(',', $_GET['a'] + $_GET['b']);</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . match ($_GET['sort']) { 'desc' => 'DESC', default => $_GET['sort'] };</error>

function viaVariable($flag)
{
    $label = $_GET['label'];
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $label;</error>
    if ($flag) {
        $either = $_GET['a'];
    } else {
        $either = $_POST['a'];
    }
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $either;</error>
}

function markupResponses($type)
{
    header('X-Frame-Options: DENY');
    header($type);
    header('Content-Type: text/plain');
    header('Content-Type: image/svg+xml');
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'];</error>
}

function xhtmlResponse($response)
{
    $response->send();
    header('Content-Type: application/xhtml+xml; charset=UTF-8');
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'];</error>
}

function unrelatedWrites(array $counts, $item)
{
    $counts[$_GET['type']] = 1;
    $item->label = 'x';
    $later = function () { $_GET['label'] = 'safe'; };
    register($_GET);
    strlen($_GET['label']);
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'];</error>
    $_GET['label'] = 'safe';
}
