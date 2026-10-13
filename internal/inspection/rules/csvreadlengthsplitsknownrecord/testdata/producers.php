<?php
// A handle assigned on several paths is not the known record.

function eitherBranch($flag, $path)
{
    if ($flag) {
        $h = fopen('data://text/plain,abcdefgh', 'r');
    } else {
        $h = fopen($path, 'r');
    }

    return fgetcsv($h, 4);
}

function eitherBranchReversed($flag, $path)
{
    if ($flag) {
        $h = fopen($path, 'r');
    } else {
        $h = fopen('data://text/plain,abcdefgh', 'r');
    }

    return fgetcsv($h, 4);
}

function replacedOnOnePath($flag, $path)
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    if ($flag) {
        $h = fopen($path, 'r');
    }

    return fgetcsv($h, 4);
}

function checkedThenReplacedOnOnePath($flag, $path)
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    if ($h === false) {
        return null;
    }
    if ($flag) {
        $h = fopen($path, 'r');
    }

    return fgetcsv($h, 4);
}

function otherProducerOnOnePath($flag, $files)
{
    if ($flag) {
        $h = fopen('data://text/plain,abcdefgh', 'r');
    } else {
        $h = $files->open();
    }

    return fgetcsv($h, 4);
}

// The first pass of the loop reads the known record.
function replacedAfterTheRead($count, $path)
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    while ($count--) {
        <warning descr="Allow enough bytes for the complete CSV record.">fgetcsv($h, 4)</warning>;
        $h = fopen($path, 'r');
    }
}
