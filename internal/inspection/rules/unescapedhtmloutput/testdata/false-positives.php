<?php
// Constants selected by request data, and numeric or boolean conversions.
$dir = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
echo '<p>' . $dir . '</p>';
echo '<p>' . ($_GET['sort'] === 'desc' ? strlen($_GET['q']) : 'ASC') . '</p>';
echo '<p>' . strlen($_GET['q']) . ' characters</p>';
echo '<p>' . count($_GET['ids']) . ' selected</p>';
echo '<p>' . ($_GET['page'] + 1) . '</p>';
echo '<p>' . ($_GET['page'] * 2 - 1) . '</p>';
echo '<p>' . sprintf('%d of %05.2f (100%%)', $_GET['page'], $_GET['total']) . '</p>';
echo '<p>' . is_numeric($_GET['page']) . '</p>';
echo '<p>' . ctype_digit($_GET['page']) . '</p>';
print '<p>' . mb_strlen($_GET['q']) . '</p>';

// Unknown output context.
echo `ls`;
echo $_GET['label'], '<p>';
echo '<p>' . $_GET['label'] . audit();
echo "<p>{$_GET['label']}{$logger->flush()}";
echo render(), '<p>', $_GET['label'];
print $_GET['label'];

function plainResponse()
{
    header('Content-Type: text/plain; charset=UTF-8');
    echo '<p>' . $_GET['label'];
}

function jsonResponses()
{
    header('Content-Type: text/html');
    header('content-type: application/problem+json');
    echo '<p>' . $_GET['label'];
}

// The request value is replaced before it is emitted.
function elementOverwritten()
{
    $_GET['v'] = 'safe';
    echo '<p>' . $_GET['v'] . '</p>';
}

function wholeArraySanitised()
{
    $_GET = array_map('htmlspecialchars', $_GET);
    echo '<p>' . $_GET['v'];
}

function rewrittenInLoop()
{
    foreach ($_POST as $key => &$value) {
        $value = htmlspecialchars($value, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8');
    }
    echo '<p>' . $_POST['v'];
}

function rewrittenByReference()
{
    array_walk($_GET, 'sanitise_entry');
    echo '<p>' . $_GET['v'];
}

function writtenThroughAlias()
{
    $alias = &$_COOKIE;
    $alias['v'] = 'safe';
    echo '<p>' . $_COOKIE['v'];
}

function writtenThroughSecondAlias()
{
    $entry = &$_GET['v'];
    $other = &$entry;
    $other = 'safe';
    echo '<p>' . $_GET['v'];
}

function boundToLocal($clean)
{
    $_GET['v'] = &$clean;
    echo '<p>' . $_GET['v'];
}

function filledFromLiteral()
{
    parse_str('v=1', $_GET);
    echo '<p>' . $_GET['v'];
}

function wholeArrayCleared()
{
    $_GET = [];
    echo '<p>' . $_GET['v'];
}

function counted()
{
    $_GET['n']++;
    echo '<p>' . $_GET['n'];
}

function removed()
{
    unset($_GET['v']);
    echo '<p>' . $_GET['v'];
}

function throughGlobals()
{
    $GLOBALS['_GET']['v'] = 'safe';
    echo '<p>' . $_GET['v'];
}
