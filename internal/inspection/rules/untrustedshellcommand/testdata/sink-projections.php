<?php
function sendRows($rows) { system(implode(' ', $rows)); }
function sendKeys($rows) { system(implode(' ', array_keys($rows))); }
function sendEncodedKeys($rows) { system(implode(' ', array_map('intval', array_keys($rows)))); }
function forwardRows($rows) { sendRows($rows); }
$rows = array_map('intval', $_GET['rows']);
sendRows($rows);
forwardRows(rows: $rows);
sendEncodedKeys($rows);
<error descr="Keep untrusted input out of shell command syntax.">sendKeys($rows)</error>;
$rows['new'] = $_GET['raw'];
<error descr="Keep untrusted input out of shell command syntax.">sendRows($rows)</error>;

final class CommandGateway {
    public function send($value) { system($value); }
    public static function dispatch($value) { system($value); }
}
class OpenGateway {
    public function send($value) { system($value); }
    final public function locked($value) { system($value); }
    private function secret($value) { system($value); }
    public function local() { <error descr="Keep untrusted input out of shell command syntax.">$this->secret($_GET['raw'])</error>; }
}
function gatewayRequests(CommandGateway $exact, OpenGateway $open) {
    <error descr="Keep untrusted input out of shell command syntax.">$exact->send($_GET['raw'])</error>;
    <error descr="Keep untrusted input out of shell command syntax.">CommandGateway::dispatch(value: $_GET['raw'])</error>;
    <error descr="Keep untrusted input out of shell command syntax.">$open->locked($_GET['raw'])</error>;
    $open->send($_GET['raw']);
    $open->secret($_GET['raw']);
}
