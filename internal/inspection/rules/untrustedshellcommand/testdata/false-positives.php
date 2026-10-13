<?php
namespace Local {
    function system($command) { return $command; }
    system('tool ' . $_GET['name']);
}

namespace {
    // A match subject only selects one of the constant arms.
    $mode = match ($_GET['mode']) { 'full' => '--full', default => '--quick' };
    system('backup ' . $mode);
    system('backup ' . match ($_GET['mode']) { 'full' => '--full', default => '--quick' });

    // Numeric and boolean results cannot carry shell syntax.
    system('head -c ' . strlen($_POST['body']) . ' /dev/zero');
    system('seq ' . count($_GET['ids']));
    system('sleep ' . ($_GET['n'] % 5));
    system('sleep ' . ($_GET['n'] - 1) . ' ' . ($_GET['n'] * 2) . ' ' . ($_GET['n'] / 2) . ' ' . ($_GET['n'] ** 2));
    system('sleep ' . -$_GET['n'] . ' ' . +$_GET['n']);
    system('sleep ' . ($_GET['n'] + 1) . ' ' . (strlen($_GET['a']) + $_GET['n']) . ' ' . ($_GET['n'] + (2)));
    system('seek ' . strpos($_GET['text'], 'x'));
    $size = strlen($_GET['body']);
    system('head -c ' . $size . ' /dev/zero');
    `head -c {$size} /dev/zero`;
    `uptime`;
    system('sleep ' . ($_GET['fast'] ? 1 : strlen($_GET['n'])));

    // The array form of proc_open bypasses the shell.
    proc_open(['convert', $_GET['file'], 'out.png'], [], $pipes);
    $argv = ['convert', $_GET['file']];
    proc_open($argv, [], $pipes);
    proc_open((array) $_GET['argv'], [], $pipes);
    proc_open();

    // Escaped arguments stay excluded.
    proc_open('convert ' . escapeshellarg($_GET['file']), [], $pipes);
    `convert {${escapeshellarg($_GET['file'])}}`;

    function launch($command)
    {
        system($command);
    }

    launch('backup ' . $mode);
    launch('head -c ' . strlen($_GET['body']) . ' /dev/zero');
    launch('viewer -- ' . escapeshellarg($_GET['path']));
    launch(...);
}
