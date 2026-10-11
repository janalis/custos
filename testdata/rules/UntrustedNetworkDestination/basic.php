<?php
$ch = curl_init($_GET['url']); <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;

function qualifiedOption() {
    $ch = curl_init();
    curl_setopt($ch, \CURLOPT_URL, $_GET['u']);
    curl_setopt($ch, CURLOPT_TIMEOUT, 5);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function optionArray() {
    $ch = curl_init();
    curl_setopt_array($ch, [CURLOPT_URL => $_GET['u'], CURLOPT_RETURNTRANSFER => true]);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function unknownOptionArrayKeepsEarlierUrl(array $options) {
    $ch = curl_init($_GET['u']);
    curl_setopt_array($ch, $options);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function hostWithoutDelimiter() {
    $ch = curl_init('https://api.example.com' . $_GET['suffix']);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function requestChoosesHost() {
    $url = "https://{$_GET['host']}/status";
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function eitherRequestValue($flag) {
    $url = $flag ? $_GET['a'] : $_GET['b'];
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function fallbackOperand() {
    $ch = curl_init($_GET['u'] ?? 'https://example.com/');
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function armFromRequest() {
    $url = match ($_GET['env']) { 'eu' => 'https://eu.example.com/', default => $_GET['u'] };
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function armWithoutDelimiter() {
    $url = match ($_GET['env']) { 'eu' => 'https://eu.example.com' . $_GET['x'], default => 'https://us.example.com/' };
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function nestedArmFromRequest() {
    $url = match ($_GET['env']) { 'eu' => match ($_GET['zone']) { default => $_GET['u'] }, default => 'https://us.example.com/' };
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function selectionThenSuffix() {
    $url = match ($_GET['env']) { 'eu' => 'https://eu.example.com', default => 'https://us.example.com' } . $_GET['x'];
    $ch = curl_init($url);
    return <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;
}

function streamReads() {
    echo strlen(<error descr="Validate the outbound network destination and redirects.">file_get_contents($_GET['u'])</error>);
    $h = <error descr="Validate the outbound network destination and redirects.">fopen('http://' . $_GET['host'] . '/status', 'r')</error>;
    <error descr="Validate the outbound network destination and redirects.">readfile(trim($_GET['u']))</error>;
    $lines = <error descr="Validate the outbound network destination and redirects.">file("{$_GET['u']}/feed")</error>;
    <error descr="Validate the outbound network destination and redirects.">copy($_GET['from'], '/tmp/mirror')</error>;
    <error descr="Validate the outbound network destination and redirects.">readfile(`printf %s {$_GET['u']}`)</error>;
    return [$h, $lines];
}

function download($url) {
    $ch = curl_init($url);
    return curl_exec($ch);
}

<error descr="Validate the outbound network destination and redirects.">download($_GET['u'])</error>;
