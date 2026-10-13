<?php
<error descr="Keep untrusted input out of shell command syntax.">system('viewer ' . $_GET['path'])</error>;
$listing = <error descr="Keep untrusted input out of shell command syntax.">`ls {$_GET['dir']}`</error>;
$process = <error descr="Keep untrusted input out of shell command syntax.">proc_open('convert ' . $_GET['file'] . ' out.png', [], $pipes)</error>;

function pickTool()
{
    $tool = match ($_GET['kind']) {
        'raw' => $_GET['tool'],
        default => 'viewer',
    };
    <error descr="Keep untrusted input out of shell command syntax.">system('run ' . $tool)</error>;
}

function fallbacks()
{
    <error descr="Keep untrusted input out of shell command syntax.">passthru('viewer ' . (getenv('VIEW') ?: $_GET['path']))</error>;
    <error descr="Keep untrusted input out of shell command syntax.">passthru('viewer ' . ($_GET['path'] ?? 'index'))</error>;
    <error descr="Keep untrusted input out of shell command syntax.">exec($_GET['left'] + $_GET['right'])</error>;
    <error descr="Keep untrusted input out of shell command syntax.">exec('viewer ' . ~$_GET['mask'])</error>;
    <error descr="Keep untrusted input out of shell command syntax.">exec('viewer ' . ($_GET['mask'] | 'a'))</error>;
    <error descr="Keep untrusted input out of shell command syntax.">shell_exec('viewer ' . trim($_GET['path']))</error>;
    <error descr="Keep untrusted input out of shell command syntax.">shell_exec("viewer {$_GET['path']}")</error>;
    <error descr="Keep untrusted input out of shell command syntax.">shell_exec('viewer ' . (isset($_GET['raw']) ? $_GET['path'] : 'index'))</error>;
}

function deeplyNested()
{
    <error descr="Keep untrusted input out of shell command syntax.">system('viewer ' . (((((((((((((((($_GET['a'] . $_GET['b'])))))))))))))))))</error>;
}

function runTool($command)
{
    system($command);
}

function runRequested()
{
    <error descr="Keep untrusted input out of shell command syntax.">system($_GET['command'])</error>;
}

<error descr="Keep untrusted input out of shell command syntax.">runTool('viewer ' . $_GET['path'])</error>;
<error descr="Keep untrusted input out of shell command syntax.">runRequested()</error>;
<error descr="Keep untrusted input out of shell command syntax.">runTool('viewer ' . strlen($_GET['path']) . $_GET['path'])</error>;

function runLabelled($label, $command)
{
    system($command);
}

<error descr="Keep untrusted input out of shell command syntax.">runLabelled('viewer', 'viewer ' . $_GET['path'])</error>;
