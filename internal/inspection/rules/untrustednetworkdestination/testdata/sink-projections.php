<?php
function numberedDestination($rows) { curl_init(implode('', $rows)); }
$rows = array_map('intval', $_GET['rows']);
<error descr="Validate the outbound network destination and redirects.">numberedDestination($rows)</error>;
final class DestinationGateway {
    public function send($url) { curl_init($url); }
    public static function dispatch($url) { curl_init($url); }
}
class OpenDestinationGateway {
    public function send($url) { curl_init($url); }
}
function destinations(DestinationGateway $exact, OpenDestinationGateway $open) {
    <error descr="Validate the outbound network destination and redirects.">$exact->send($_GET['url'])</error>;
    <error descr="Validate the outbound network destination and redirects.">DestinationGateway::dispatch(url: $_GET['url'])</error>;
    $open->send($_GET['url']);
}
