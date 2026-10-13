<?php
// Pipes filled by one of several process launches have no known direction.

function eitherLaunch($flag)
{
    if ($flag) {
        $p = proc_open(["cat"], [0 => ["pipe", "w"]], $pipes);
    } else {
        $p = proc_open(["cat"], [0 => ["pipe", "r"]], $pipes);
    }
    if (is_resource($p)) {
        fwrite($pipes[0], "input");
    }
}

function eitherLaunchReversed($flag)
{
    if ($flag) {
        $p = proc_open(["cat"], [0 => ["pipe", "r"]], $pipes);
    } else {
        $p = proc_open(["cat"], [0 => ["pipe", "w"]], $pipes);
    }
    if (is_resource($p)) {
        fwrite($pipes[0], "input");
    }
}

function pipesReplacedOnOnePath($flag, $other)
{
    $p = proc_open(["cat"], [0 => ["pipe", "w"]], $pipes);
    if ($flag) {
        $pipes = $other;
    }
    if (is_resource($p)) {
        fwrite($pipes[0], "input");
    }
}

// The pipes are those of the one launch, whatever the handle tested holds.
function handleReplacedOnOnePath($flag, $other)
{
    $p = proc_open(["cat"], [0 => ["pipe", "w"]], $pipes);
    if ($flag) {
        $p = $other;
    }
    if (is_resource($p)) {
        <warning descr="Choose pipe modes from the child perspective.">fwrite($pipes[0], "input")</warning>;
    }
}
