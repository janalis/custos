<?php
function rejectsWithTruthyMatch() {
    $label = $_GET['label'];
    if (preg_match('/[\r\n]/', $label)) {
        exit;
    }
    header('X-Label: ' . $label);
}

function rejectsWithCharacterList() {
    $label = $_GET['label'];
    if (strpbrk($label, "\r\n") !== false) {
        throw new InvalidArgumentException('line break');
    }
    header("X-Label: {$label}");
}

function emitsOnlyWhenClean() {
    $label = $_GET['label'];
    if (!preg_match("#[\n\r]#", $label)) {
        header('X-Label: ' . $label);
    }
}

function numericResults() {
    header('Content-Length: ' . strlen($_POST['body']));
    header('X-Total-Count: ' . count($_GET['ids']));
    header('X-Page: ' . ($_GET['page'] + 1));
    header('X-Digits: ' . ctype_digit($_GET['page']));
}

function finiteChoice() {
    $direction = match ($_GET['sort']) {
        'desc' => 'DESC',
        default => 'ASC',
    };
    header('X-Sort: ' . $direction);
}

function wrongPattern() {
    $label = $_GET['label'];
    if (preg_match('/[^\r\n]/', $label)) {
        exit;
    }
    <error descr="Reject line breaks and validate this header value.">header('X-Label: ' . $label)</error>;
}

function guardDoesNotLeave() {
    $label = $_GET['label'];
    if (preg_match('/[\r\n]/', $label)) {
        error_log('line break');
    }
    <error descr="Reject line breaks and validate this header value.">header('X-Label: ' . $label)</error>;
}

function reassignedAfterGuard() {
    $label = $_GET['label'];
    if (preg_match('/[\r\n]/', $label)) {
        exit;
    }
    $label = $_GET['other'];
    <error descr="Reject line breaks and validate this header value.">header('X-Label: ' . $label)</error>;
}

function armKeepsRequestText() {
    $value = match ($_GET['mode']) {
        'raw' => $_GET['value'],
        default => 'none',
    };
    <error descr="Reject line breaks and validate this header value.">header('X-Value: ' . $value)</error>;
}

function arrayUnion() {
    <error descr="Reject line breaks and validate this header value.">header('X-Ids: ' . implode(',', $_GET['a'] + $_GET['b']))</error>;
}
