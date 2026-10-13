<?php
echo $_GET['url'];

function fixedHostEncodedQuery() {
    $ch = curl_init('https://api.example.com/search?q=' . urlencode($_GET['q']));
    return curl_exec($ch);
}

function fixedHostNumericSegment() {
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, 'https://api.example.com/items/' . (int) $_GET['id']);
    return curl_exec($ch);
}

function fixedHostBuiltQuery() {
    $base = 'https://api.example.com/search?';
    $ch = curl_init($base . http_build_query(['q' => $_GET['q']]));
    return curl_exec($ch);
}

function fixedHostInterpolated() {
    $ch = curl_init("https://api.example.com/items/{$_GET['id']}");
    return curl_exec($ch);
}

function selectedLiteralHost() {
    $url = match ($_GET['env']) {
        'eu' => 'https://eu.example.com/',
        'asia' => match ($_GET['zone']) { 'jp' => 'https://jp.example.com/', default => 'https://sg.example.com/' },
        'beta' => ('https://beta.example.com/v2/' . $_GET['tenant']) . '/status',
        default => 'https://us.example.com/',
    };
    $ch = curl_init($url);
    return curl_exec($ch);
}

function laterFixedUrlWins() {
    $ch = curl_init($_GET['u']);
    curl_setopt_array($ch, [CURLOPT_URL => 'https://example.com/', CURLOPT_POSTFIELDS => $_POST['body']]);
    return curl_exec($ch);
}

function otherOptionsOnly() {
    $ch = curl_init('https://example.com/');
    curl_setopt($ch, CURLOPT_USERAGENT, $_SERVER['HTTP_USER_AGENT']);
    curl_setopt_array($ch, [CURLOPT_POSTFIELDS => $_POST['body']]);
    return curl_exec($ch);
}

function localAndFixedStreams() {
    readfile('/var/data/' . $_GET['name']);
    readfile(__DIR__ . '/' . $_GET['name']);
    readfile("/var/data/{$_GET['name']}");
    readfile((int) $_GET['id'] . '.txt');
    readfile(basename($_GET['name']));
    readfile('/var/data/static.txt');
    copy('/var/data/' . $_GET['name'], '/tmp/copy');
    return file_get_contents('https://api.example.com/items/' . rawurlencode($_GET['id']));
}

function download($url) {
    $ch = curl_init($url);
    return curl_exec($ch);
}

function announce($value) {
    header('X-Trace: ' . $value);
}

download('https://example.com/');
download(lookup($_GET['key']));
announce($_GET['trace']);
