<?php
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error> . '</p>';
echo "<p>{<error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>}</p>\n";
echo "<p><error descr="Escape untrusted text for its HTML output context.">$_GET[label]</error> for $name</p>";
echo '<p>', <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>, '</p>', render();
print '<p>' . (<error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error> . '</p>');
echo <<<HTML
    <li class="row">{<error descr="Escape untrusted text for its HTML output context.">$_POST['label']</error>}</li>
    HTML;
echo '<p>' . (<error descr="Escape untrusted text for its HTML output context.">$_GET['label'] ?? 'none'</error>);
echo '<p>' . (<error descr="Escape untrusted text for its HTML output context.">$_GET['short'] ? $_GET['label'] : 'none'</error>);
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">strtoupper($_GET['label'])</error>;
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">sprintf('%s items', $_GET['count'])</error>;
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">sprintf('%c', $_GET['count'])</error>;
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">implode(',', $_GET['a'] + $_GET['b'])</error>;
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">match ($_GET['sort']) { 'desc' => 'DESC', default => $_GET['sort'] }</error>;

function viaVariable($flag)
{
    $label = $_GET['label'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$label</error>;
    if ($flag) {
        $either = $_GET['a'];
    } else {
        $either = $_POST['a'];
    }
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$either</error>;
}

function markupResponses($type)
{
    header('X-Frame-Options: DENY');
    header($type);
    header('Content-Type: text/plain');
    header('Content-Type: image/svg+xml');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>;
}

function xhtmlResponse($response)
{
    $response->send();
    header('Content-Type: application/xhtml+xml; charset=UTF-8');
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>;
}

function unrelatedWrites(array $counts, $item)
{
    $counts[$_GET['type']] = 1;
    $item->label = 'x';
    $later = function () { $_GET['label'] = 'safe'; };
    register($_GET);
    strlen($_GET['label']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>;
    $_GET['label'] = 'safe';
}

// Writes that keep request text in the superglobal replace nothing.
function wholeArrayFromRequest()
{
    $_GET = $_POST;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error> . '</p>';
}

function wholeArrayMerged()
{
    $_GET += $_POST;
    $_REQUEST = array_merge($_REQUEST, $_COOKIE);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_REQUEST['v']</error>;
}

function dynamicElementFromRequest($key)
{
    $_GET[$key] = $_POST['v'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error>;
}

function aliased()
{
    $alias = &$_COOKIE;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_COOKIE['v']</error>;
}

function requestWrittenThroughAlias($item)
{
    $entry = &$_GET['v'];
    $entry = $_POST['v'];
    $item->entry = &$_GET['w'];
    $item->entry = 'safe';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error>;
}

function filledFromRequest()
{
    parse_str($_SERVER['QUERY_STRING'], $_GET);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error>;
}

function passedToUnresolvedCode()
{
    fill_request($_GET, $_POST);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['v']</error>;
}
